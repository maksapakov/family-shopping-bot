package port

import (
	"context"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

type ListRepository interface {
	GetByChatID(ctx context.Context, chatID domain.ChatID) (*domain.ShoppingList, error)
	Save(ctx context.Context, list *domain.ShoppingList) error
}
