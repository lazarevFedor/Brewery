package main

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository"
	"Brewery/migrator"
	"Brewery/pkg/postgres"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func fileData(filename string) (*os.File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open csv file: %w", err)
	}

	return file, nil
}

func parseFile(filename string) ([]entities.Beer, error) {
	file, err := fileData(filename)
	defer file.Close()
	if err != nil {
		return nil, fmt.Errorf("fileReader: %w", err)
	}

	fileReader := csv.NewReader(file)
	records, err := fileReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ReadAll: %w", err)
	}
	records = records[1:]

	beers := make([]entities.Beer, 0, len(records))

	for _, record := range records {
		beer := entities.Beer{}
		rating, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			return nil, fmt.Errorf("rating ParseFloat: %w", err)
		}
		abv, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			return nil, fmt.Errorf("abv ParseFloat: %w", err)
		}
		ibu, err := strconv.Atoi(record[4])
		if err != nil {
			return nil, fmt.Errorf("ibu Atoi: %w", err)
		}
		features := strings.Split(record[5], ", ")

		beer.Name = record[0]
		beer.Rating = float32(rating)
		beer.Description = record[2]
		beer.ABV = float32(abv)
		beer.IBU = uint8(ibu)
		beer.City = record[6]
		beer.Country = record[8]
		beer.Type = record[7]
		beer.Features = features
		beer.Category.Name = "-"

		beers = append(beers, beer)
	}
	return beers, nil
}

func fillDB(ctx context.Context, filename string, repo *repository.BeerPostgres) error {
	beers, err := parseFile(filename)
	if err != nil {
		return fmt.Errorf("parseFile: %w", err)
	}
	for _, beer := range beers {
		_, err = repo.InsertBeer(ctx, beer)
		if err != nil {
			return fmt.Errorf("InsertBeer: %w", err)
		}
	}
	return nil
}

func getBeers(ctx context.Context, repo *repository.BeerPostgres) {
	beers, err := repo.GetBeers(ctx, 1, 0)
	if err != nil {
		fmt.Print("GetBeers: ", err)
	}
	for _, beer := range beers {
		fmt.Println(beer)
	}
}

func main() {
	ctx := context.Background()
	postgresCfg := postgres.Config{
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Password: "1234",
		DB:       "brewery_db",
		MinConns: 1,
		MaxConns: 10,
	}
	pool, err := postgres.NewPool(ctx, postgresCfg)
	if err != nil {
		panic(fmt.Errorf("failed to create postgres pool: %w", err))
	}

	err = migrator.Up(pool)

	beerRepo := repository.NewBeerPostgres(pool)
	err = fillDB(ctx, "beers.csv", beerRepo)
	if err != nil {
		panic(err)
	}
}
