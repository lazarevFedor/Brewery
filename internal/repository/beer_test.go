// Package repository contains layer that manipulates data in database
package repository_test

import (
	"Brewery/internal/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeerRepository_Insert(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		testBeer := entities.Beer{
			Name:        "test",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			City:        "Москва",
			Country:     "Россия",
			Type:        "Lager",
			Category: entities.ProductCategory{
				Name: "test_category",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		}

		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, ctx, ctgRepo.Pool, "beers")
		})

		var beers []entities.Beer

		beers, err = beerRepo.GetBeers(ctx, 1, 0)
		require.NoError(t, err)
		require.Len(t, beers, 1, "Длина слайса должна быть равно 1")
		require.Equal(t, testBeer.Name, beers[0].Name, "Имя вставленного и полученного товара должны быть равны")
		require.Equal(t, testBeer.City, beers[0].City, "Имя вставленного и полученного товара должны быть равны")
	})
}
