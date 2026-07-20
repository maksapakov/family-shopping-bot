package domain

import (
	"fmt"
	"strings"
)

var ErrItemNotFound = fmt.Errorf("domain: item not found")
var ErrInvalidLocation = fmt.Errorf("domain: invalid location")
var ErrNothingToUndo = fmt.Errorf("domain: nothing to undo")

type ShoppingList struct {
	ID         ListID
	ChatID     ChatID
	ActiveTab  Location
	Items      []Item // слайс - упорядоченные элементы списка
	MessageRef string // platform-specific message reference (на mvp string позже тип)
	undoStack  []undoAction
}

func NewShoppingList(id ListID, chatID ChatID) *ShoppingList {
	return &ShoppingList{
		ID:        id,
		ChatID:    chatID,
		ActiveTab: LocationProducts,
		Items:     []Item{}, // Не nil!
		undoStack: []undoAction{},
	}
}

func (l *ShoppingList) ToggleItem(itemID ItemID) error {
	for i := range l.Items {
		if l.Items[i].ID == itemID {
			wasChecked := l.Items[i].IsChecked()
			wasAt := l.Items[i].CheckedAt

			l.Items[i].Toggle()

			l.undoStack = append(l.undoStack, undoAction{
				itemID:     itemID,
				wasChecked: wasChecked,
				checkedAt:  wasAt,
			})

			if len(l.undoStack) > 10 {
				l.undoStack = l.undoStack[len(l.undoStack)-10:]
			}
			return nil
		}
	}
	return ErrItemNotFound
}

func (l *ShoppingList) AddItem(item Item) {
	l.Items = append(l.Items, item)
}

func (l *ShoppingList) Undo() error {
	if len(l.undoStack) == 0 {
		return ErrNothingToUndo
	}
	last := l.undoStack[len(l.undoStack)-1]
	l.undoStack = l.undoStack[:len(l.undoStack)-1]

	for i := range l.Items {
		if l.Items[i].ID == last.itemID {
			l.Items[i].Restore(last.wasChecked, last.checkedAt)
			return nil
		}
	}
	return ErrItemNotFound
}

func (l *ShoppingList) ItemsByTab() []Item {
	result := make([]Item, 0)
	for _, item := range l.Items {
		if item.location == l.ActiveTab {
			result = append(result, item)
		}
	}
	return result
}

func (l *ShoppingList) SwitchTab(loc Location) error {
	if loc == LocationUnknown {
		return ErrInvalidLocation
	}
	l.ActiveTab = loc
	return nil
}

func (l *ShoppingList) HasItemName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, i := range l.Items {
		if strings.EqualFold(i.Name, name) {
			return true
		}
	}
	return false
}
