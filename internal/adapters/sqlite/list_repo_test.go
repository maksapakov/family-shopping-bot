package sqlite

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

func TestOpen_creates_db(t *testing.T) {
	wd, _ := os.Getwd()
	_ = wd

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	repo, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func(repo *ListRepo) {
		err := repo.Close()
		if err != nil {
			slog.Error("close", "error", err)
		}
	}(repo)
}

func TestListRepo_SaveAndGet(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	repo, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))

	if err := repo.Save(ctx, list); err != nil {
		t.Fatalf("failed to save list: %v", err)
	}

	got, err := repo.GetByChatID(ctx, "chat-1")
	if err != nil {
		t.Fatalf("failed to get by chat id: %v", err)
	}
	if got == nil {
		t.Fatalf("got nil list")
	}
	if got.ID != list.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, list.ID)
	}
	if got.ChatID != list.ChatID {
		t.Fatalf("got chatID = %q, want %q", got.ChatID, list.ChatID)
	}
	if len(got.Items) != 1 {
		t.Fatalf("got len(items) = %d, want %d", len(got.Items), 1)
	}
	if got.Items[0].Name != "Milk" {
		t.Fatalf("got items[0].Name = %q, want %q", got.Items[0].Name, "Milk")
	}
}

func TestListRepo_SaveAndGet_Trim(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	repo, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer func() {
		_ = repo.Close()
	}()

	list := domain.NewShoppingList("list-1", "chat-1")
	for i := range 11 {
		list.AddItem(domain.NewItem(
			domain.ItemID(fmt.Sprintf("item-%d", i)),
			fmt.Sprintf("Name-%d", i),
			"mom",
			domain.LocationProducts,
		))
	}
	for _, i := range list.Items {
		err := list.ToggleItem(i.ID)
		if err != nil {
			t.Fatalf("toggle error = %v", err)
		}
	}

	if err := repo.Save(ctx, list); err != nil {
		t.Fatalf("failed to save list: %v", err)
	}

	got, err := repo.GetByChatID(ctx, "chat-1")
	if err != nil {
		t.Fatalf("failed to get by chat id: %v", err)
	}
	if got == nil {
		t.Fatalf("got nil list")
	}
	if got.ID != list.ID {
		t.Fatalf("got.ID = %q, want %q", got.ID, list.ID)
	}
	if got.ChatID != list.ChatID {
		t.Fatalf("got chatID = %q, want %q", got.ChatID, list.ChatID)
	}
	if len(got.Items) != 10 {
		t.Fatalf("got len(items) = %d, want %d", len(got.Items), 10)
	}
	for _, item := range got.Items {
		if item.ID == "item-0" {
			t.Fatalf("list contains item item-0")
		}
	}
}
