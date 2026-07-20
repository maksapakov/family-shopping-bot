package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

func TestListRepo_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	repo := NewListRepo()
	list := domain.NewShoppingList("list-1", "chat-1")

	err := repo.Save(ctx, list)
	if err != nil {
		t.Fatalf("Save(): error = %v, want nil", err)
	}

	got, err := repo.GetByChatID(ctx, "chat-1")
	if err != nil {
		t.Fatalf("GetByChatId(): error = %v, want nil", err)
	}
	if got == nil {
		t.Fatalf("GetByChatId(): got nil = %v", got)
	}
	if got.ID != list.ID {
		t.Fatalf("GetByChatId(): got id = %v, want %v", got.ID, list.ID)
	}
	if got.ChatID != list.ChatID {
		t.Fatalf("GetByChatId(): got chat_id = %v, want %v", got.ChatID, list.ChatID)
	}
}

func TestListRepo_not_found(t *testing.T) {
	ctx := context.Background()
	repo := NewListRepo()

	_, err := repo.GetByChatID(ctx, "no-such-chat")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByChatId(): error = %v, want ErrNotFound", err)
	}
}

var _ port.ListRepository = (*ListRepo)(nil)
