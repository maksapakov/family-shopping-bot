package app

import (
	"context"
	"errors"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/adapters/memory"
	"github.com/maksapakov/family-shopping-bot/internal/auth/callback"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/httpx"
)

func TestUndoItem_Execute(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))

	err := list.ToggleItem("item-1")

	if err != nil {
		t.Fatalf("toggle item fail: %s", err)
	}

	err = repo.Save(ctx, list)

	if err != nil {
		t.Fatalf("save list: %s", err)
	}

	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewUndoItem(repo, messenger, links)

	err = uc.Execute(ctx, "chat-1")

	if err != nil {
		t.Fatalf("failed to execute undo item: %s", err)
	}

	got, err := repo.GetByChatID(ctx, "chat-1")

	if err != nil {
		t.Fatalf("failed to get item by chat-1: %s", err)
	}

	if got.Items[0].IsChecked() == true {
		t.Fatalf("item should be unchecked after undo")
	}

	err = uc.Execute(ctx, "chat-1")

	if !errors.Is(err, domain.ErrNothingToUndo) {
		t.Fatalf("Execute(): error = %v, wantErr %v", err, domain.ErrNothingToUndo)
	}
}
