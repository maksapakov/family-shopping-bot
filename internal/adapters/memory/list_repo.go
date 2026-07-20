package memory

import (
	"context"

	"fmt"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

var ErrNotFound = fmt.Errorf("list not found")

type ListRepo struct {
	lists map[domain.ChatID]*domain.ShoppingList
}

func NewListRepo() *ListRepo {
	return &ListRepo{
		lists: make(map[domain.ChatID]*domain.ShoppingList),
	}
}

func (r *ListRepo) Save(ctx context.Context, list *domain.ShoppingList) error {
	r.lists[list.ChatID] = list
	return nil
}

func (r *ListRepo) GetByChatID(ctx context.Context, chatID domain.ChatID) (*domain.ShoppingList, error) {
	list, ok := r.lists[chatID]
	if !ok {
		return nil, ErrNotFound
	}
	return list, nil
}
