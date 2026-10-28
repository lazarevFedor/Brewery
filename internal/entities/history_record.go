package entities

import "time"

type HistoryRecord struct {
	ID        int
	ProductID int
	Changes   string
	CreatedAt time.Time
}
