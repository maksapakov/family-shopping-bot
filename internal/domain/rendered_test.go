package domain

import "testing"

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
