package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// HistoryRecordPostgres реализует интерфейс HistoryRecordRepository для работы с записями истории в базе данных PostgreSQL.
type HistoryRecordPostgres struct {
	Pool *pgxpool.Pool
}

// NewHistoryRecordPostgres создает новый экземпляр HistoryRecordPostgres с использованием предоставленного пула соединений к базе данных.
func NewHistoryRecordPostgres(pool *pgxpool.Pool) *HistoryRecordPostgres {
	return &HistoryRecordPostgres{Pool: pool}
}
