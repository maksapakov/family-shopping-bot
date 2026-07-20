package app

import (
	"context"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/adapters/memory"
	"github.com/maksapakov/family-shopping-bot/internal/auth/callback"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/httpx"
)

func TestToggleItem_Execute(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	list.MessageRef = "msg-42"

	err := repo.Save(ctx, list)

	if err != nil {
		t.Fatalf("failed to save list: %s", err)
	}

	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewToggleItem(repo, messenger, links)

	err = uc.Execute(ctx, "chat-1", "item-1")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}

	got, _ := repo.GetByChatID(ctx, "chat-1")
	if got.Items[0].IsChecked() != true {
		t.Fatalf("got %v, want %v", got.Items[0].IsChecked(), true)
	}
	if messenger.CallCount != 1 {
		t.Fatalf("CallCount = %d, want %d", messenger.CallCount, 1)
	}
	if len(messenger.LastRendered.Items) != 1 {
		t.Fatalf("LastRendered.Items = %d, want %d", len(messenger.LastRendered.Items), 1)
	}
	if !messenger.LastRendered.Items[0].IsChecked {
		t.Fatalf("LastRendered.Items[0].IsChecked = %v, want %v", messenger.LastRendered.Items[0].IsChecked, true)
	}
}

func TestToggleItem_Execute_Fail(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1") // заведомо отсутствует Item
	err := repo.Save(ctx, list)

	if err != nil {
		t.Fatalf("failed to save list: %s", err)
	}

	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewToggleItem(repo, messenger, links)

	err = uc.Execute(ctx, "chat-1", "item-1")
	if err == nil {
		t.Fatalf("want error, got nil")
	}
}

func TestToggleItem_Execute_sends_when_no_message_ref(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewListRepo()
	messenger := memory.NewFakeMessenger()

	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	err := repo.Save(ctx, list)
	if err != nil {
		t.Fatalf("failed to save list: %v", err)
	}

	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewToggleItem(repo, messenger, links)

	err = uc.Execute(ctx, "chat-1", "item-1")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got, _ := repo.GetByChatID(ctx, "chat-1")
	if got.MessageRef == "" {
		t.Fatalf("MessageRef should be saved after first send, got %v, want %v", got.MessageRef, "")
	}
	if messenger.CallCount != 1 {
		t.Fatalf("CallCount = %d, want %d", messenger.CallCount, 1)
	}
}
