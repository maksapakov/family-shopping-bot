package port

import "github.com/maksapakov/family-shopping-bot/internal/domain"

type LinkBuilder interface {
	ToggleURL(chatID domain.ChatID, itemID domain.ItemID) string
	UndoURL(chatID domain.ChatID) string
}
