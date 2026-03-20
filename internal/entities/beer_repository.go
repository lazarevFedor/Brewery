package entities

import(
	"context"

)

// BeerRepository определяет контракт для хранения и получения данных о пиве.
type BeerRepository interface{
	
	// InsertBeer сохраняет новую сущность Beer в хранилище.
	InsertBeer(ctx context.Context, beer Beer) error

	// GetBeers возвращает список всех сортов пива.
	GetBeers(ctx context.Context) ([]Beer, error)
}