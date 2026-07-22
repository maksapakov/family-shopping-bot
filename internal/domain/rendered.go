package domain

type RenderedItem struct {
	ID        ItemID
	Name      string
	IsChecked bool
	ClickURL  string
}

type RenderedList struct {
	ActiveTab Location
	Items     []RenderedItem
	UndoURL   string
}

func (l *ShoppingList) Render() RenderedList {
	items := l.ItemsByTab()
	var renderedItems []RenderedItem
	for _, item := range items {
		if !item.IsChecked() {
			renderedItems = append(renderedItems, RenderedItem{
				ID:        item.ID,
				Name:      item.Name,
				IsChecked: item.IsChecked(),
				ClickURL:  "",
			})
		}
	}
	return RenderedList{
		ActiveTab: l.ActiveTab,
		Items:     renderedItems,
		UndoURL:   "",
	}
}
