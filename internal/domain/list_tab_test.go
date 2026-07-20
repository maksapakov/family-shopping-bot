package domain

import (
	"errors"
	"testing"
)

func TestShoppingList_SwitchTab(t *testing.T) {
	tests := []struct {
		name     string
		switchTo Location
		wantName string
		wantErr  error
	}{
		{
			name:     "products tab",
			switchTo: LocationProducts,
			wantName: "Milk",
			wantErr:  nil,
		},
		{
			name:     "pickupTab",
			switchTo: LocationPickup,
			wantName: "Keyboard",
			wantErr:  nil,
		},
		{
			name:     "unknown location",
			switchTo: LocationUnknown,
			wantErr:  ErrInvalidLocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewShoppingList("list-1", "chat-1")
			list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
			list.AddItem(NewItem("item-2", "Keyboard", "dad", LocationPickup))

			err := list.SwitchTab(tt.switchTo)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("SwitchTab() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			items := list.ItemsByTab()
			if len(items) != 1 {
				t.Errorf("len(items) = %d, want 1", len(items))
			}
			if tt.wantName != items[0].Name {
				t.Errorf("items[0].Name() = %v, want %v", items[0].Name, tt.wantName)
			}
		})
	}
}

func TestItemsByTab_filters_by_active_tab(t *testing.T) {
	list := NewShoppingList("list-1", "chat-1")

	list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
	list.AddItem(NewItem("item-2", "Keyboard", "dad", LocationPickup))

	items := list.ItemsByTab()

	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want %d", len(items), 1)
	}
	if items[0].Name != "Milk" {
		t.Fatalf("items[0].Name() = %s, want %s", items[0].Name, "Milk")
	}
}

func TestSwitchTab_changes_visible_items(t *testing.T) {
	list := NewShoppingList("list-1", "chat-1")
	list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
	list.AddItem(NewItem("item-2", "Keyboard", "dad", LocationPickup))

	err := list.SwitchTab(LocationPickup)
	if err != nil {
		t.Fatalf("SwitchTab() error = %v, want nil", err)
	}

	items := list.ItemsByTab()
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want %d", len(items), 1)
	}
	if items[0].Name != "Keyboard" {
		t.Fatalf("items[0].Name() = %s, want %s", items[0].Name, "Keyboard")
	}
}
