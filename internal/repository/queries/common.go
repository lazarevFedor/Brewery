// Package queries содержит функции для сборки запросов к базе данных
package queries

import sq "github.com/Masterminds/squirrel"

// psql - это экземпляр StatementBuilder, настроенный на использование формата плейсхолдеров для PostgreSQL.
// Используется во всех функциях пакета для сборки SQL запросов.
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
