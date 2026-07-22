package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksapakov/family-shopping-bot/internal/adapters/matrix"
	"github.com/maksapakov/family-shopping-bot/internal/adapters/memory"
	"github.com/maksapakov/family-shopping-bot/internal/adapters/sqlite"
	"github.com/maksapakov/family-shopping-bot/internal/app"
	"github.com/maksapakov/family-shopping-bot/internal/auth/callback"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/httpx"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("family shopping bot starting")

	ctx := context.Background()
	dbPath := "shopping.db"
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		dbPath = v
	}
	repo, err := sqlite.Open(dbPath)
	if err != nil {
		slog.Error("failed to open database", "path", dbPath, "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			slog.Error("failed to close repository", "error", err)
		}
	}()

	var messenger port.ChatMessenger

	hs, token, room := os.Getenv("MATRIX_HOMESERVER"), os.Getenv("MATRIX_ACCESS_TOKEN"), os.Getenv("MATRIX_ROOM_ID")
	secret := os.Getenv("CALLBACK_SECRET")
	signer, err := callback.NewSigner(secret)
	if err != nil {
		slog.Error("failed to create signer", "error", err)
		os.Exit(1)
	}
	baseURL := os.Getenv("CALLBACK_BASE_URL")
	links := httpx.NewLinkBuilder(baseURL, signer)

	if hs == "" || token == "" || room == "" {
		messenger = memory.NewFakeMessenger()
		slog.Info("using fake messenger", "hint", "set MATRIX_* to enabled")
	} else {
		m, err := matrix.NewMessenger(hs, token, room)
		if err != nil {
			slog.Error("failed to create messenger", "error", err)
			os.Exit(1)
		}
		messenger = m
		slog.Info("using matrix messenger", "hs", hs)

		client := &http.Client{Timeout: 90 * time.Second}

		addItem := app.NewAddItem(repo, messenger, links)
		listener := matrix.NewListener(client, hs, token, room, addItem)
		go func() {
			if err := listener.Run(ctx); err != nil {
				slog.Error("failed to run listener", "error", err)
			}
		}()
	}

	toggleUc := app.NewToggleItem(repo, messenger, links)
	undoUc := app.NewUndoItem(repo, messenger, links)

	if os.Getenv("DEMO") == "1" {
		if err := runDemo(ctx, toggleUc, repo, signer); err != nil {
			slog.Error("demo failure", "error", err)
			os.Exit(1)
		}
	} else {
		slog.Info("demo skipped", "hint", "set DEMO=1 to seed demo-chat")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /toggle", func(w http.ResponseWriter, r *http.Request) {
		chatID := domain.ChatID(r.URL.Query().Get("chat"))
		itemID := domain.ItemID(r.URL.Query().Get("item"))

		if chatID == "" || itemID == "" {
			http.Error(w, "chat or item id required", http.StatusBadRequest)
			return
		}

		sig := r.URL.Query().Get("sig")
		if !signer.VerifyToggle(string(chatID), string(itemID), sig) {
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		err := toggleUc.Execute(r.Context(), chatID, itemID)
		if errors.Is(err, domain.ErrItemNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		if err != nil {
			slog.Error("toggle failure", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /undo", func(w http.ResponseWriter, r *http.Request) {
		chatID := domain.ChatID(r.URL.Query().Get("chat"))
		if chatID == "" {
			http.Error(w, "chat id required", http.StatusBadRequest)
			return
		}

		sig := r.URL.Query().Get("sig")
		if !signer.VerifyUndo(string(chatID), sig) {
			http.Error(w, "invalid signature", http.StatusForbidden)
			return
		}

		err := undoUc.Execute(r.Context(), chatID)
		if errors.Is(err, domain.ErrNothingToUndo) {
			http.Error(w, "nothing to undo", http.StatusConflict)
			return
		}
		if err != nil {
			slog.Error("undo failure", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	srv := &http.Server{
		Addr:    ":8181",
		Handler: mux,
	}

	go func() {
		slog.Info("starting http server", "addr", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server failure", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "error", err)
	}

	slog.Info("bye")
}

func runDemo(
	ctx context.Context,
	toggle *app.ToggleItem,
	repo port.ListRepository,
	signer *callback.Signer,
) error {
	got, err := repo.GetByChatID(ctx, "demo-chat")
	if errors.Is(err, sqlite.ErrNotFound) {
		list := domain.NewShoppingList("demo-list", "demo-chat")
		list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
		if err := repo.Save(ctx, list); err != nil {
			return fmt.Errorf("save list: %w", err)
		}
		slog.Info("before toggle", "checked", list.Items[0].IsChecked())
		if err := toggle.Execute(ctx, "demo-chat", "item-1"); err != nil {
			return fmt.Errorf("toggle item-1: %w", err)
		}
		got, err = repo.GetByChatID(ctx, "demo-chat")
		if err != nil {
			return fmt.Errorf("get demo-chat: %w", err)
		}
		slog.Info("after toggle", "got", got.Items[0].IsChecked(), "message_ref", got.MessageRef)
		slog.Info("demo urls",
			"toggle", fmt.Sprintf("/toggle?chat=demo-chat&item=item-1&sig=%s",
				signer.SignToggle("demo-chat", "item-1")),
			"undo", fmt.Sprintf("/undo?chat=demo-chat&sig=%s",
				signer.SignUndo("demo-chat")),
		)
		return nil
	}
	if err != nil {
		return fmt.Errorf("get demo-chat: %w", err)
	}
	slog.Info("demo list already exists", "elements", len(got.Items))
	return nil
}
