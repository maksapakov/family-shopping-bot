package port

import (
	"context"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

type MessageRef string

type ChatMessenger interface {
	SendList(ctx context.Context, chatID domain.ChatID, rendered domain.RenderedList) (MessageRef, error)
	UpdateList(ctx context.Context, chatID domain.ChatID, ref MessageRef, rendered domain.RenderedList) error
}
