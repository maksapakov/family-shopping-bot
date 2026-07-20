package matrix

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

func TestMessenger_SendList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("got HTTP method %s, want %s", r.Method, http.MethodPut)
		}
		got := r.Header.Get("Authorization")
		if got != "Bearer token" {
			t.Errorf("got Authorization header %s, want %s", got, "Bearer token")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"event_id":"$evt1:test"}`))
	}))
	defer srv.Close()

	m, err := NewMessenger(srv.URL, "token", "!room:test")
	if err != nil {
		t.Fatal("NewMessenger: failed", err)
	}

	rendered := domain.RenderedList{
		Items: []domain.RenderedItem{
			{
				ID:        "",
				Name:      "Milk",
				IsChecked: false,
				ClickURL:  "http://example/toggle",
			},
		},
		UndoURL: "http://example/undo",
	}

	ref, err := m.SendList(context.Background(), "chat-1", rendered)
	if err != nil {
		t.Fatal("SendList: failed", err)
	}
	if ref != "$evt1:test" {
		t.Fatalf("SendList: failed ref = %q, want %q", ref, "$evt1:test")
	}
}

func TestMessenger_UpdateList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("got HTTP method %s, want %s", r.Method, http.MethodPut)
		}
		gotAuth := r.Header.Get("Authorization")
		if gotAuth != "Bearer token" {
			t.Errorf("got Authorization header %s, want %s", gotAuth, "Bearer token")
		}

		w.Header().Set("Content-Type", "application/json")
		body, _ := io.ReadAll(r.Body)
		var got struct {
			RelatesTo struct {
				RelType string `json:"rel_type"`
				EventID string `json:"event_id"`
			} `json:"m.relates_to"`
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal response body: %v\nraw: %s", err, body)
		}
		if got.RelatesTo.RelType != "m.replace" {
			t.Errorf("rel_type = %q, want %q", got.RelatesTo.RelType, "m.replace")
		}
		if got.RelatesTo.EventID != "$evt1:test" {
			t.Errorf("rel_type = %q, want %q", got.RelatesTo.EventID, "$evt1:test")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"event_id":"$evt2:test"}`))
	}))
	defer srv.Close()

	m, err := NewMessenger(srv.URL, "token", "!room:test")
	if err != nil {
		t.Fatal("NewMessenger: failed", err)
	}

	rendered := domain.RenderedList{
		Items: []domain.RenderedItem{
			{
				ID:        "",
				Name:      "Milk",
				IsChecked: false,
				ClickURL:  "http://example/toggle",
			},
		},
		UndoURL: "http://example/undo",
	}

	err = m.UpdateList(context.Background(), "chat-1", "$evt1:test", rendered)
	if err != nil {
		t.Fatal("UpdateList: failed", err)
	}
}
