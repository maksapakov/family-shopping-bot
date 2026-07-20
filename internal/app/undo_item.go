package app

import (
	"context"
	"fmt"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

type UndoItem struct {
	repo      port.ListRepository
	messenger port.ChatMessenger
	links     port.LinkBuilder
}

func NewUndoItem(
	repo port.ListRepository,
	messenger port.ChatMessenger,
	links port.LinkBuilder,
) *UndoItem {
	return &UndoItem{
		repo:      repo,
		messenger: messenger,
		links:     links,
	}
}

func (uc *UndoItem) Execute(ctx context.Context, chatID domain.ChatID) error {
	list, err := uc.repo.GetByChatID(ctx, chatID)

	if err != nil {
		return fmt.Errorf("get list by chatID: %w", err)
	}

	err = list.Undo()

	if err != nil {
		return fmt.Errorf("undo list: %w", err)
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
