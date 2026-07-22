package domain

import (
	"fmt"
	"testing"
)

func TestShoppingList_Render_filters_by_tab(t *testing.T) {
	list := NewShoppingList("list-1", "chat-1")
	list.AddItem(NewItem("item-1", "Milk", "mom", LocationProducts))
	list.AddItem(NewItem("item-2", "Keyboard", "dad", LocationPickup))

	rendered := list.Render()
	if len(rendered.Items) != 1 {
		t.Fatalf("rendered items count is %d, want 1", len(rendered.Items))
	}
	if rendered.Items[0].Name != "Milk" {
		t.Fatalf("rendered items name is %s, want Milk", rendered.Items[0].Name)
	}
}

func TestShoppingList_Render(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(l *ShoppingList)
		query       []string
		wantLen     int
		wantName    []string
		wantChecked bool
	}{
		{
			name:     "two_items_added",
			setup:    func(l *ShoppingList) {},
			query:    []string{"item-1", "item-2"},
			wantLen:  2,
			wantName: []string{"item-1", "item-2"},
		},
		{
			name: "two_items_added_one_toggle",
			setup: func(l *ShoppingList) {
				l.Items[0].Toggle()
			},
			query:       []string{"item-1", "item-2"},
			wantLen:     2,
			wantName:    []string{"item-1", "item-2"},
			wantChecked: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewShoppingList("list-1", "chat-1")
			for i, item := range tt.query {
				list.AddItem(NewItem(
					ItemID(fmt.Sprintf("item-%d", i)),
					item,
					"mom",
					LocationProducts,
				))
			}
			tt.setup(list)

			got := list.Render()
			if len(got.Items) != tt.wantLen {
				t.Fatalf("rendered items count is %d, want %d", len(got.Items), tt.wantLen)
			}
			for i, item := range tt.wantName {
				if got.Items[i].Name != item {
					t.Fatalf("rendered items name is %s, want %s", got.Items[i].Name, item)
				}
			}
			if tt.wantChecked != got.Items[0].IsChecked {
				t.Fatalf("rendered items count is %v, want true", got.Items[0].IsChecked)
			}
		})
	}
}
