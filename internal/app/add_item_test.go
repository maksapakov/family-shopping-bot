package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/adapters/memory"
	"github.com/maksapakov/family-shopping-bot/internal/auth/callback"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/httpx"
)

func TestAddItem_Execute(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	list.MessageRef = "msg-1"

	err := repo.Save(ctx, list)
	if err != nil {
		t.Fatalf("failed to save list: %s", err)
	}

	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewAddItem(repo, messenger, links)

	err = uc.Execute(ctx, "chat-1", "Potato", "dad")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}

	got, _ := repo.GetByChatID(ctx, "chat-1")
	if got.Items[1].Name != "Potato" {
		t.Fatalf("item got Name = %v, want %v", got.Items[1].Name, "Potato")
	}
}

func TestAddItem_ExecuteMany(t *testing.T) {
	tests := []struct {
		name      string
		existing  []string
		input     []string
		wantLen   int
		wantCalls int
	}{
		{
			name:      "skip dupes add bread",
			existing:  []string{"Milk"},
			input:     []string{"milk", "Bread", "bread"},
			wantLen:   2,
			wantCalls: 1,
		},
		{
			name:      "all dupes no publish",
			existing:  []string{"Milk"},
			input:     []string{"milk"},
			wantLen:   1,
			wantCalls: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := memory.NewListRepo()
			list := domain.NewShoppingList("list-1", "chat-1")
			list.MessageRef = "msg-1"
			for i, name := range tt.existing {
				list.AddItem(domain.NewItem(domain.ItemID(fmt.Sprintf("item-%d", i)),
					name,
					"mom",
					domain.LocationProducts))
			}
			err := repo.Save(ctx, list)
			if err != nil {
				t.Fatalf("failed to save list: %s", err)
			}

			messenger := memory.NewFakeMessenger()
			signer, _ := callback.NewSigner("test-secret")
			links := httpx.NewLinkBuilder("http://localhost:8181", signer)
			uc := NewAddItem(repo, messenger, links)

			err = uc.ExecuteMany(ctx, "chat-1", tt.input, "dad")
			if err != nil {
				t.Fatalf("failed to execute: %s", err)
			}
			got, _ := repo.GetByChatID(ctx, "chat-1")
			if len(got.Items) != tt.wantLen {
				t.Fatalf("len(items) = %d, want %d", len(got.Items), tt.wantLen)
			}
			if messenger.CallCount != tt.wantCalls {
				t.Fatalf("len(items) = %d, want %d", messenger.CallCount, tt.wantCalls)
			}
		})
	}
}

func TestAddItem_ExecuteMany_oneDupes(t *testing.T) {
	// В БД уже есть Milk. Вызываем ExecuteMany c ["milk", "Bread", "bread"].
	// Должно стать 2 items (Milk, Bread). Messenger вызывается 1 раз
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	list.MessageRef = "msg-1"
	err := repo.Save(ctx, list)
	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewAddItem(repo, messenger, links)
	err = uc.ExecuteMany(ctx, "chat-1", []string{"milk", "Bread", "bread"}, "mom")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}
	got, err := repo.GetByChatID(ctx, "chat-1")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("len(item) = %v, want %v", len(got.Items), 2)
	}
	if !got.HasItemName("Milk") || !got.HasItemName("Bread") {
		t.Fatalf("names: %#v", got.Items)
	}
	if messenger.CallCount != 1 {
		t.Fatalf("messenger.CallCount = %v, want %v", messenger.CallCount, 1)
	}
}

func TestAddItem_ExecuteMany_allDupes(t *testing.T) {
	// «Уже есть Milk. ExecuteMany(["milk"]). Список по-прежнему 1 item. CallCount = 0.»
	ctx := context.Background()
	repo := memory.NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")
	list.AddItem(domain.NewItem("item-1", "Milk", "mom", domain.LocationProducts))
	list.MessageRef = "msg-1"
	err := repo.Save(ctx, list)
	messenger := memory.NewFakeMessenger()
	signer, _ := callback.NewSigner("test-secret")
	links := httpx.NewLinkBuilder("http://localhost:8181", signer)
	uc := NewAddItem(repo, messenger, links)
	err = uc.ExecuteMany(ctx, "chat-1", []string{"milk"}, "mom")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}
	got, err := repo.GetByChatID(ctx, "chat-1")
	if err != nil {
		t.Fatalf("failed to execute: %s", err)
	}
	if len(got.Items) != 1 {
		t.Fatalf("len(item) = %v, want %v", len(got.Items), 2)
	}
	if messenger.CallCount != 0 {
		t.Fatalf("messenger.CallCount = %v, want %v", messenger.CallCount, 1)
	}
}
