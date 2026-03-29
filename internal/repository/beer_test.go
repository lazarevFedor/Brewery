// Package repository contains layer that manipulates data in database
package repository_test

import (
	"Brewery/internal/entities"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testBeer = entities.Beer{
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

	testReview = entities.Review{
		Body:   "test_body",
		Rating: 4.5,
	}
)

func TestBeerRepository_InsertGetBeer(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 1, "Длина слайса должна быть равна 1")
		require.Equal(t, testBeer.Name, beers[0].Name, "Имя вставленного и полученного товара должны быть равны")
		require.Equal(t, testBeer.City, beers[0].City, "Имя вставленного и полученного товара должны быть равны")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_UpdateBeer(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное обновление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		updates := map[string]any{
			"name":    "updated_name",
			"city_id": "1",
		}
		beerID, err = beerRepo.UpdateBeer(ctx, beerID, updates)

		require.NoError(t, err)
		require.NotZero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 1, "Длина слайса должна быть равна 1")
		require.Equal(t, updates["name"], beers[0].Name, "Имя обновленного и полученного товара должны быть равны")
		require.Equal(t, "test_city", beers[0].City, "Город обновленного и полученного товара должны быть равны")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_DeleteBeer(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		err = beerRepo.DeleteBeer(ctx, beerID)

		require.NoError(t, err)

		beer, _ := beerRepo.GetBeerByID(ctx, beerID)

		require.Nil(t, beer, "Функция должна вернуть nil")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_InsertReview(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		testReview.BeerID = beerID
		reviewID, err := beerRepo.InsertReview(ctx, testReview)

		require.NotZero(t, reviewID)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetBeersByCategoryID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeer)
		require.NoError(t, err, "InsertBeer Error")
		require.NotZero(t, beerID)

		beer, _ := beerRepo.GetBeerByID(ctx, beerID)
		require.NotNil(t, beer, "GetBeerByID Error")

		ctgID, _ := ctgRepo.GetCategoryID(ctx, beer.Category.Name)
		require.NotZero(t, ctgID, "GetCategoryID Error")

		beers, err := beerRepo.GetBeersByCategoryID(ctx, ctgID, 0, 0)

		require.NoError(t, err, "GetBeersByCategoryID Error")
		require.Len(t, beers, 1, "Длина слайса должна быть равна 1")
		require.Equal(t, testBeer.Name, beers[0].Name, "Имя вставленного и полученного товара должны быть равны")
		require.Equal(t, testBeer.City, beers[0].City, "Имя вставленного и полученного товара должны быть равны")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}
