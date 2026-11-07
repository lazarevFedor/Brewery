package entities

import "time"

type HistoryRecord struct {
	ID        int64
	ProductID int64
	Changes   string
	CreatedAt time.Time
}
