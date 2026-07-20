package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

var ErrEmptyInput = errors.New("empty input")

type AddItem struct {
	repo      port.ListRepository
	messenger port.ChatMessenger
	links     port.LinkBuilder
}

func NewAddItem(
	repo port.ListRepository,
	messenger port.ChatMessenger,
	links port.LinkBuilder,
) *AddItem {
	return &AddItem{
		repo:      repo,
		messenger: messenger,
		links:     links,
	}
}

func (uc *AddItem) Execute(
	ctx context.Context,
	chatID domain.ChatID,
	name string,
	addedBy string,
) error {
	space := strings.TrimSpace(name)
	if len(space) == 0 {
		return ErrEmptyInput
	}
	return uc.ExecuteMany(ctx, chatID, []string{name}, addedBy)
}

func (uc *AddItem) ExecuteMany(
	ctx context.Context,
	chatID domain.ChatID,
	names []string,
	addedBy string,
) error {
	list, err := uc.repo.GetByChatID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get list: %w", err)
	}

	added := 0
	for _, name := range names {
		name := strings.TrimSpace(name)
		if len(name) == 0 {
			continue
		}
		if list.HasItemName(name) {
			continue
		}
		id := fmt.Sprintf("item-%d", time.Now().UnixNano())
		list.AddItem(domain.NewItem(domain.ItemID(id), name, addedBy, domain.LocationProducts))
		added++
	}

	if added == 0 {
		return nil
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
