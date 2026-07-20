package domain

import (
	"errors"
	"testing"
)

func TestShoppingList_ToggleItem(t *testing.T) {
	tests := []struct {
		name      string
		itemID    ItemID
		location  Location
		wantErr   error
		wantCheck bool
	}{
		{
			name:      "toggle existing item",
			itemID:    "item-1",
			location:  LocationProducts,
			wantErr:   nil,
			wantCheck: true,
		},
		{
			name:    "toggle missing item",
			itemID:  "ghost",
			wantErr: ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewShoppingList("list-1", "chat-1")
			list.AddItem(NewItem("item-1", "Молоко", "мама", 1))

			err := list.ToggleItem(tt.itemID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ToggleItem() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			item := list.Items[0]
			if item.IsChecked() != tt.wantCheck {
				t.Errorf("isChecked() = %v, location = %v, want %v", item.IsChecked(), tt.location, tt.wantCheck)
			}
		})
	}
}

func TestUndo_restores_previous_state(t *testing.T) {
	list := NewShoppingList("list-1", "chat-1")
	list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))

	err := list.ToggleItem("item-1")
	if err != nil {
		t.Fatalf("ToggleItem() error = %v, want nil", err)
	}

	if !list.Items[0].IsChecked() {
		t.Fatalf("items[0].IsChecked() = false, want true")
	}

	err = list.Undo()
	if err != nil {
		t.Fatalf("Undo() error = %v, want nil", err)
	}

	if list.Items[0].IsChecked() {
		t.Fatalf("items[0].IsChecked() = true, want false")
	}

	err = list.Undo()
	if !errors.Is(err, ErrNothingToUndo) {
		t.Fatalf("Undo() error = %v, want ErrNothingToUndo", err)
	}
}

func TestShoppingList_HasItemName(t *testing.T) {
	list := NewShoppingList("list-1", "chat-1")

	if list.HasItemName("Milk") {
		t.Fatal("empty list HasItemName() = true, want false")
	}

	list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))

	if !list.HasItemName("Milk") {
		t.Fatal("exact name HasItemName() = false, want true")
	}
	if !list.HasItemName("milk") {
		t.Fatal("case insensitive list HasItemName() = false, want true")
	}
	if list.HasItemName("Bread") {
		t.Fatal("other name HasItemName() = true, want false")
	}
}

func TestShoppingList_HasItemName_table(t *testing.T) {
	tests := []struct {
		name  string
		setup func(list *ShoppingList)
		query string
		want  bool
	}{
		{
			name:  "empty list",
			setup: func(l *ShoppingList) {},
			query: "Milk",
			want:  false,
		},
		{
			name: "exact match",
			setup: func(l *ShoppingList) {
				l.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
			},
			query: "Milk",
			want:  true,
		},
		{
			name: "case insensitive",
			setup: func(l *ShoppingList) {
				l.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
			},
			query: "milk",
			want:  true,
		},
		{
			name: "third item match",
			setup: func(l *ShoppingList) {
				l.AddItem(NewItem("item-1", "A", "x", LocationProducts))
				l.AddItem(NewItem("item-2", "B", "x", LocationProducts))
				l.AddItem(NewItem("item-3", "Bread", "x", LocationProducts))
			},
			query: "bread",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewShoppingList("list-1", "chat-1")
			tt.setup(list)

			got := list.HasItemName(tt.query)
			if got != tt.want {
				t.Fatalf("HasItemName(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
