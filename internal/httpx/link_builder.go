package httpx

import (
	"fmt"

	"github.com/maksapakov/family-shopping-bot/internal/auth/callback"
	"github.com/maksapakov/family-shopping-bot/internal/domain"
)

type LinkBuilder struct {
	baseURL string
	signer  callback.Signer
}

func NewLinkBuilder(baseURL string, signer *callback.Signer) *LinkBuilder {
	return &LinkBuilder{
		baseURL: baseURL,
		signer:  *signer,
	}
}

func (b *LinkBuilder) ToggleURL(chatID domain.ChatID, itemID domain.ItemID) string {
	sig := b.signer.SignToggle(string(chatID), string(itemID))
	return fmt.Sprintf("%s/toggle?chat=%s&item=%s&sig=%s",
		b.baseURL, chatID, itemID, sig)
}

func (b *LinkBuilder) UndoURL(chatID domain.ChatID) string {
	sig := b.signer.SignUndo(string(chatID))
	return fmt.Sprintf("%s/undo?chat=%s&sig=%s",
		b.baseURL, chatID, sig)
}
