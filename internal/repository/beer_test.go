// Package repository contains layer that manipulates data in database
package repository_test

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"Brewery/pkg/logger"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testBeers = []entities.Beer{
		{
			Name:        "test_success",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			Amount:      100,
			Unit:        "Литр",
			City:        "Москва",
			Country:     "Россия",
			Category: entities.ProductCategory{
				Name: "test_category",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		},
		{
			Name:        "test_failure_empty_category",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			Amount:      100,
			Unit:        "Литр",
			City:        "Москва",
			Country:     "Россия",
			Category: entities.ProductCategory{
				Name: "",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		},
		{
			Name: "test_failure_uninitialized_repository",
		},
		{
			Name:        "test_failure_null_category",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			Amount:      100,
			Unit:        "Литр",
			City:        "Москва",
			Country:     "Россия",
			Category: entities.ProductCategory{
				Name: "null",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		},
		{
			Name:        "test_failure_empty_city",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			Amount:      100,
			Unit:        "Литр",
			City:        "",
			Country:     "",
			Category: entities.ProductCategory{
				Name: "test_category",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		},
		{
			Name:        "test_failure_empty_country",
			Rating:      4.7,
			Description: "test description",
			ABV:         4.7,
			IBU:         100,
			Amount:      100,
			Unit:        "Литр",
			City:        "Москва",
			Country:     "",
			Category: entities.ProductCategory{
				Name: "test_category",
			},
			Features: []string{"feat1", "feat2", "feat3"},
		},
	}

	testReview = entities.Review{
		Body:   "test_body",
		Rating: 4.5,
	}
)

func TestBeerRepository_InsertGetBeer(t *testing.T) {
	ctx := t.Context()
	ctx, err := logger.NewLoggerContext(ctx, true)
	if err != nil {
		panic(fmt.Errorf("failed to create logger context: %w", err))
	}

	t.Run("Успешная вставка", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err)
		require.NotZero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 1, "Длина слайса должна быть равна 1")
		require.Equal(t, testBeers[0].Name, beers[0].Name, "Имя вставленного и полученного товара должны быть равны")
		require.Equal(t, testBeers[0].City, beers[0].City, "Имя вставленного и полученного товара должны быть равны")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с пустой категорией", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[1])
		require.Error(t, err)
		require.Zero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 0, "Длина слайса должна быть равна 0")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с не инициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}
		beerID, err := uninitializedBeerRepo.InsertBeer(ctx, testBeers[1])
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с null категорией вторым корнем", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[3])
		require.Error(t, err)
		require.Zero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 0, "Длина слайса должна быть равна 0")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с пустым городом", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[4])
		require.Error(t, err)
		require.Zero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 0, "Длина слайса должна быть равна 0")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с пустой страной", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[5])
		require.Error(t, err)
		require.Zero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)

		require.NoError(t, err)
		require.Len(t, beers, 0, "Длина слайса должна быть равна 0")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_UpdateBeer(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное обновление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
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

	t.Run("Обновление несуществующего пива", func(t *testing.T) {
		updates := map[string]any{
			"name": "updated_name",
		}

		beerID, err := beerRepo.UpdateBeer(ctx, 999999, updates)
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Обновление с пустым набором полей", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err)
		require.NotZero(t, beerID)

		beerID, err = beerRepo.UpdateBeer(ctx, beerID, map[string]any{})
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Обновление с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		updates := map[string]any{
			"name": "updated_name",
		}
		beerID, err := uninitializedBeerRepo.UpdateBeer(ctx, 1, updates)
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_DeleteBeer(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
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

	t.Run("Удаление несуществующего пива", func(t *testing.T) {
		err := beerRepo.DeleteBeer(ctx, 999999)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Удаление с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		err := uninitializedBeerRepo.DeleteBeer(ctx, 1)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_InsertReview(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное удаление", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
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

	t.Run("Вставка отзыва с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		reviewID, err := uninitializedBeerRepo.InsertReview(ctx, testReview)
		require.Error(t, err)
		require.Zero(t, reviewID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetBeersByCategoryID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err, "InsertBeer Error")
		require.NotZero(t, beerID)

		beer, _ := beerRepo.GetBeerByID(ctx, beerID)
		require.NotNil(t, beer, "GetBeerByID Error")

		ctgID, _ := ctgRepo.GetCategoryID(ctx, beer.Category.Name)
		require.NotZero(t, ctgID, "GetCategoryID Error")

		beers, err := beerRepo.GetBeersByCategoryID(ctx, ctgID, 0, 0)

		require.NoError(t, err, "GetBeersByCategoryID Error")
		require.Len(t, beers, 1, "Длина слайса должна быть равна 1")
		require.Equal(t, testBeers[0].Name, beers[0].Name, "Имя вставленного и полученного товара должны быть равны")
		require.Equal(t, testBeers[0].City, beers[0].City, "Имя вставленного и полученного товара должны быть равны")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Лимит при выборке по категории", func(t *testing.T) {
		firstBeer := testBeers[0]
		firstBeer.Name = "test_first"
		secondBeer := testBeers[0]
		secondBeer.Name = "test_second"

		firstID, err := beerRepo.InsertBeer(ctx, firstBeer)
		require.NoError(t, err)
		require.NotZero(t, firstID)

		secondID, err := beerRepo.InsertBeer(ctx, secondBeer)
		require.NoError(t, err)
		require.NotZero(t, secondID)

		ctgID, err := ctgRepo.GetCategoryID(ctx, firstBeer.Category.Name)
		require.NoError(t, err)
		require.NotZero(t, ctgID)

		beers, err := beerRepo.GetBeersByCategoryID(ctx, ctgID, 1, 0)
		require.NoError(t, err)
		require.Len(t, beers, 1)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Выборка по категории с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		beers, err := uninitializedBeerRepo.GetBeersByCategoryID(ctx, 1, 0, 0)
		require.Error(t, err)
		require.Nil(t, beers)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetBeers(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная выборка с лимитом", func(t *testing.T) {
		firstBeer := testBeers[0]
		firstBeer.Name = "get_beers_first"
		secondBeer := testBeers[0]
		secondBeer.Name = "get_beers_second"

		_, err := beerRepo.InsertBeer(ctx, firstBeer)
		require.NoError(t, err)

		_, err = beerRepo.InsertBeer(ctx, secondBeer)
		require.NoError(t, err)

		beers, err := beerRepo.GetBeers(ctx, 1, 0)
		require.NoError(t, err)
		require.Len(t, beers, 1)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Выборка с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		beers, err := uninitializedBeerRepo.GetBeers(ctx, 0, 0)
		require.Error(t, err)
		require.Nil(t, beers)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetBeerByID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное получение по ID", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err)
		require.NotZero(t, beerID)

		beer, err := beerRepo.GetBeerByID(ctx, beerID)
		require.NoError(t, err)
		require.NotNil(t, beer)
		require.Equal(t, testBeers[0].Name, beer.Name)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Получение несуществующего пива", func(t *testing.T) {
		beer, err := beerRepo.GetBeerByID(ctx, 999999)
		require.Error(t, err)
		require.Nil(t, beer)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Получение с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		beer, err := uninitializedBeerRepo.GetBeerByID(ctx, 1)
		require.Error(t, err)
		require.Nil(t, beer)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetCountryID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное получение ID страны", func(t *testing.T) {
		countryID, err := beerRepo.GetCountryID(ctx, "test_country")
		require.NoError(t, err)
		require.NotZero(t, countryID)

		countryIDSecond, err := beerRepo.GetCountryID(ctx, "test_country")
		require.NoError(t, err)
		require.Equal(t, countryID, countryIDSecond)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Пустое имя страны", func(t *testing.T) {
		countryID, err := beerRepo.GetCountryID(ctx, "")
		require.Error(t, err)
		require.Zero(t, countryID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Получение страны с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		countryID, err := uninitializedBeerRepo.GetCountryID(ctx, "test_country")
		require.Error(t, err)
		require.Zero(t, countryID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetCityID(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешное получение ID города", func(t *testing.T) {
		countryID, err := beerRepo.GetCountryID(ctx, "test_country")
		require.NoError(t, err)
		require.NotZero(t, countryID)

		cityID, err := beerRepo.GetCityID(ctx, "test_city", countryID)
		require.NoError(t, err)
		require.NotZero(t, cityID)

		cityIDSecond, err := beerRepo.GetCityID(ctx, "test_city", countryID)
		require.NoError(t, err)
		require.Equal(t, cityID, cityIDSecond)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Пустое имя города", func(t *testing.T) {
		cityID, err := beerRepo.GetCityID(ctx, "", 1)
		require.Error(t, err)
		require.Zero(t, cityID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Получение города с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		cityID, err := uninitializedBeerRepo.GetCityID(ctx, "test_city", 1)
		require.Error(t, err)
		require.Zero(t, cityID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_GetFeatureIDAndInsertBeerFeature(t *testing.T) {
	ctx := t.Context()

	t.Run("Успешная вставка связи beer-feature", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err)
		require.NotZero(t, beerID)

		featureID, err := beerRepo.GetFeatureID(ctx, "manual_feature")
		require.NoError(t, err)
		require.NotZero(t, featureID)

		err = beerRepo.InsertBeerFeature(ctx, featureID, beerID)
		require.NoError(t, err)

		err = beerRepo.InsertBeerFeature(ctx, featureID, beerID)
		require.NoError(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Получение feature с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		featureID, err := uninitializedBeerRepo.GetFeatureID(ctx, "feat")
		require.Error(t, err)
		require.Zero(t, featureID)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка beer-feature с неинициализированным репозиторием", func(t *testing.T) {
		uninitializedBeerRepo := repository.BeerPostgres{}

		err := uninitializedBeerRepo.InsertBeerFeature(ctx, 1, 1)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

func TestBeerRepository_CanceledContext(t *testing.T) {
	baseCtx := t.Context()
	ctx, cancel := context.WithCancel(baseCtx)
	cancel()

	t.Run("Вставка пива с отмененным контекстом", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Обновление пива с отмененным контекстом", func(t *testing.T) {
		beerID, err := beerRepo.UpdateBeer(ctx, 1, map[string]any{"name": "x"})
		require.Error(t, err)
		require.Zero(t, beerID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Удаление пива с отмененным контекстом", func(t *testing.T) {
		err := beerRepo.DeleteBeer(ctx, 1)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Вставка отзыва с отмененным контекстом", func(t *testing.T) {
		reviewID, err := beerRepo.InsertReview(ctx, entities.Review{Body: "x", Rating: 1.0, BeerID: 1})
		require.Error(t, err)
		require.Zero(t, reviewID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Получение списка пива с отмененным контекстом", func(t *testing.T) {
		beers, err := beerRepo.GetBeers(ctx, 0, 0)
		require.Error(t, err)
		require.Nil(t, beers)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Получение пива по категории с отмененным контекстом", func(t *testing.T) {
		beers, err := beerRepo.GetBeersByCategoryID(ctx, 1, 0, 0)
		require.Error(t, err)
		require.Nil(t, beers)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})
}
