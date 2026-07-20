package domain

import "time"

type Item struct {
	ID        ItemID
	Name      string
	isChecked bool      // unexported - снаружи меняем только через методы
	CheckedAt time.Time // когда вычеркнут (для undo/TTL)
	AddedBy   string    // кто добавил - display name, не user ID
	location  Location  // unexported - на какой вкладке
}

func NewItem(id ItemID, name, addedBy string, loc Location) Item {
	return Item{
		ID:       id,
		Name:     name,
		AddedBy:  addedBy,
		location: loc,
	}
}

func (i *Item) IsChecked() bool {
	return i.isChecked
}

func (i *Item) Toggle() {
	i.isChecked = !i.isChecked
	if i.isChecked {
		i.CheckedAt = time.Now()
	} else {
		i.CheckedAt = time.Time{} // zero = "не вычеркнуто"
	}
}

func (i *Item) Restore(wasChecked bool, checkedAt time.Time) {
	i.isChecked = wasChecked
	i.CheckedAt = checkedAt
}

func (i *Item) Location() Location {
	return i.location
}
