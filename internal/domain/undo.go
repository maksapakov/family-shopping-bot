package domain

import "time"

type undoAction struct {
	itemID     ItemID
	wasChecked bool
	checkedAt  time.Time
}
