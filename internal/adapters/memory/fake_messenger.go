package memory

import (
	"context"

	"github.com/maksapakov/family-shopping-bot/internal/domain"
	"github.com/maksapakov/family-shopping-bot/internal/port"
)

type FakeMessenger struct {
	LastChatID   domain.ChatID
	LastRef      port.MessageRef
	LastRendered domain.RenderedList
	CallCount    int
}

func NewFakeMessenger() *FakeMessenger {
	return &FakeMessenger{}
}

func (f *FakeMessenger) SendList(
	ctx context.Context,
	chatID domain.ChatID,
	rendered domain.RenderedList,
) (port.MessageRef, error) {
	f.LastChatID = chatID
	f.LastRendered = rendered
	f.CallCount++
	return "fake-msg-1", nil
}

func (f *FakeMessenger) UpdateList(
	ctx context.Context,
	chatID domain.ChatID,
	ref port.MessageRef,
	rendered domain.RenderedList,
) error {
	f.LastChatID = chatID
	f.LastRef = ref
	f.LastRendered = rendered
	f.CallCount++
	return nil
}

var _ port.ChatMessenger = (*FakeMessenger)(nil)
