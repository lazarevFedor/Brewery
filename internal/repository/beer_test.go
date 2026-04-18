// Package repository_test содержит тесты для слоя repository
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
	// testBeers содержит набор тестовых данных для пива, включая как валидные, так и невалидные случаи для проверки различных сценариев вставки и получения данных из репозитория.
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

	// testReview содержит тестовые данные для отзыва, которые используются в тестах вставки отзывов в репозиторий.
	testReview = entities.Review{
		Body:   "test_body",
		Rating: 4.5,
	}
)

// TestBeerRepository_InsertGetBeer содержит тесты для проверки функциональности вставки и получения пива из репозитория, включая различные сценарии, такие как успешная вставка, вставка с невалидными данными и использование неинициализированного репозитория.
//
//nolint:funlen
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
		require.Empty(t, beers, "Длина слайса должна быть равна 0")

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
		require.Empty(t, beers, "Длина слайса должна быть равна 0")

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
		require.Empty(t, beers, "Длина слайса должна быть равна 0")

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
		require.Empty(t, beers, "Длина слайса должна быть равна 0")

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с новой дочерней категорией", func(t *testing.T) {
		rootID, err := ctgRepo.GetCategoryID(ctx, "test_category")
		require.NoError(t, err)
		require.NotZero(t, rootID)

		beer := testBeers[0]
		beer.Name = "test_success_new_child_category"
		beer.Category = entities.ProductCategory{
			Name:     "test_category_child",
			ParentID: int(rootID),
		}

		beerID, err := beerRepo.InsertBeer(ctx, beer)
		require.NoError(t, err)
		require.NotZero(t, beerID)

		gotBeer, err := beerRepo.GetBeerByID(ctx, beerID)
		require.NoError(t, err)
		require.NotNil(t, gotBeer)
		require.Equal(t, beer.Category.Name, gotBeer.Category.Name)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})

	t.Run("Вставка с невалидной фичей", func(t *testing.T) {
		beer := testBeers[0]
		beer.Name = "test_failure_invalid_feature"
		beer.Features = []string{"feat_ok", "bad\x00feature"}

		beerID, err := beerRepo.InsertBeer(ctx, beer)
		require.Error(t, err)
		require.Zero(t, beerID)

		beers, err := beerRepo.GetBeers(ctx, 0, 0)
		require.NoError(t, err)
		require.Empty(t, beers)

		t.Cleanup(func() {
			cleanDB(t, ctx, "beers")
		})
	})
}

// TestBeerRepository_UpdateBeer содержит тесты для проверки функциональности обновления пива в репозитории, включая успешное обновление, попытку обновления несуществующего пива, обновление с пустым набором полей и обновление с неинициализированным репозиторием.
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

// TestBeerRepository_DeleteBeer содержит тесты для проверки функциональности удаления пива из репозитория, включая успешное удаление, попытку удаления несуществующего пива и удаление с неинициализированным репозиторием.
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

	t.Run("Удаление пива каскадно удаляет отзывы", func(t *testing.T) {
		beerID, err := beerRepo.InsertBeer(ctx, testBeers[0])
		require.NoError(t, err)
		require.NotZero(t, beerID)

		review := testReview
		review.BeerID = beerID

		reviewID, err := beerRepo.InsertReview(ctx, review)
		require.NoError(t, err)
		require.NotZero(t, reviewID)

		err = beerRepo.DeleteBeer(ctx, beerID)
		require.NoError(t, err)

		var reviewCount int
		err = testDB.QueryRow(ctx, "SELECT COUNT(*) FROM reviews WHERE beer_id = $1", beerID).Scan(&reviewCount)
		require.NoError(t, err)
		require.Zero(t, reviewCount)

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

// TestBeerRepository_InsertReview содержит тесты для проверки функциональности вставки отзывов в репозиторий, включая успешную вставку отзыва, попытку вставки отзыва с неинициализированным репозиторием и вставку отзыва с отмененным контекстом.
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

// TestBeerRepository_GetBeersByCategoryID содержит тесты для проверки функциональности получения пива по категории из репозитория, включая успешное получение, получение с лимитом и получение с неинициализированным репозиторием.
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

// TestBeerRepository_GetBeers содержит тесты для проверки функциональности получения списка пива из репозитория, включая успешную выборку с лимитом и выборку с неинициализированным репозиторием.
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

// TestBeerRepository_GetCountryID содержит тесты для проверки функциональности получения ID страны из репозитория, включая успешное получение, получение с пустым именем страны и получение с неинициализированным репозиторием.
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

// TestBeerRepository_GetCityID содержит тесты для проверки функциональности получения ID города из репозитория, включая успешное получение, получение с пустым именем города и получение с неинициализированным репозиторием.
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

// TestBeerRepository_GetFeatureIDAndInsertBeerFeature содержит тесты для проверки функциональности получения ID фичи и вставки связи beer-feature в репозитории, включая успешную вставку, получение ID фичи с неинициализированным репозиторием и вставку beer-feature с неинициализированным репозиторием.
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

// TestBeerRepository_CanceledContext содержит тесты для проверки поведения методов репозитория при использовании отмененного контекста, включая попытки вставки, обновления, удаления и получения данных с отмененным контекстом, а также проверку правильного обработки ошибок в этих случаях.
//
//nolint:funlen
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

	t.Run("Получение страны с отмененным контекстом", func(t *testing.T) {
		countryID, err := beerRepo.GetCountryID(ctx, "x")
		require.Error(t, err)
		require.Zero(t, countryID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Получение города с отмененным контекстом", func(t *testing.T) {
		cityID, err := beerRepo.GetCityID(ctx, "x", 1)
		require.Error(t, err)
		require.Zero(t, cityID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Получение feature с отмененным контекстом", func(t *testing.T) {
		featureID, err := beerRepo.GetFeatureID(ctx, "x")
		require.Error(t, err)
		require.Zero(t, featureID)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})

	t.Run("Вставка связи beer-feature с отмененным контекстом", func(t *testing.T) {
		err := beerRepo.InsertBeerFeature(ctx, 1, 1)
		require.Error(t, err)

		t.Cleanup(func() {
			cleanDB(t, baseCtx, "beers")
		})
	})
}
