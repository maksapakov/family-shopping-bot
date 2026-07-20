package app

import (
	"context"
	"fmt"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

func publishList(
	ctx context.Context,
	messenger port.ChatMessenger,
	list *domain.ShoppingList,
	links port.LinkBuilder,
) error {
	rendered := list.Render()
	for i := range rendered.Items {
		rendered.Items[i].ClickURL = links.ToggleURL(list.ChatID, rendered.Items[i].ID)
	}
	rendered.UndoURL = links.UndoURL(list.ChatID)

	if list.MessageRef == "" {
		ref, err := messenger.SendList(ctx, list.ChatID, rendered)
		if err != nil {
			return fmt.Errorf("messenger.SendList: %w", err)
		}
		list.MessageRef = string(ref)
		return nil
	}

	ref := port.MessageRef(list.MessageRef)
	err := messenger.UpdateList(ctx, list.ChatID, ref, rendered)
	if err != nil {
		return fmt.Errorf("messenger.UpdateList: %w", err)
	}
	return nil
}
