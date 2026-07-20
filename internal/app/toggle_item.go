package app

import (
	"context"
	"fmt"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

type ToggleItem struct {
	repo      port.ListRepository
	messenger port.ChatMessenger
	links     port.LinkBuilder
}

func NewToggleItem(
	repo port.ListRepository,
	messenger port.ChatMessenger,
	links port.LinkBuilder,
) *ToggleItem {
	return &ToggleItem{
		repo:      repo,
		messenger: messenger,
		links:     links,
	}
}

func (uc *ToggleItem) Execute(ctx context.Context, chatID domain.ChatID, itemID domain.ItemID) error {
	list, err := uc.repo.GetByChatID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get list by chatID: %w", err)
	}

	err = list.ToggleItem(itemID)
	if err != nil {
		return fmt.Errorf("toggle item: %w", err)
	}

	err = uc.repo.Save(ctx, list)

	if err != nil {
		return fmt.Errorf("save list: %w", err)
	}

	err = publishList(ctx, uc.messenger, list, uc.links)
	if err != nil {
		return fmt.Errorf("publish list: %w", err)
	}

	err = uc.repo.Save(ctx, list)
	if err != nil {
		return fmt.Errorf("save list: %w", err)
	}

	return nil
}
