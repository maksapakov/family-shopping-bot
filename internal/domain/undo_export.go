package domain

import "time"

type UndoSnapshot struct {
	ItemID     ItemID
	WasChecked bool
	CheckedAt  time.Time
}

func (l *ShoppingList) UndoSnapshot() []UndoSnapshot {
	out := make([]UndoSnapshot, len(l.undoStack))
	for i, a := range l.undoStack {
		out[i] = UndoSnapshot{
			ItemID:     a.itemID,
			WasChecked: a.wasChecked,
			CheckedAt:  a.checkedAt,
		}
	}
	return out
}

func (l *ShoppingList) RestoreUndoSnapshot(snaps []UndoSnapshot) {
	l.undoStack = make([]undoAction, 0, len(snaps))
	for _, s := range snaps {
		l.undoStack = append(l.undoStack, undoAction{
			itemID:     s.ItemID,
			wasChecked: s.WasChecked,
			checkedAt:  s.CheckedAt,
		})
	}
}
