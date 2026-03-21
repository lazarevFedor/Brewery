package main

import (
	"Brewery/internal/entities"
	repository "Brewery/internal/repository/beer"
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

	// Name, Rating, Description, ABV (%), IBU, Features, City, Category, Country
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
		err = repo.InsertBeer(ctx, beer)
		if err != nil {
			return fmt.Errorf("InsertBeer: %w", err)
		}
	}
	return nil
}

func getBeers(ctx context.Context, repo *repository.BeerPostgres) {
	beers, err := repo.GetBeers(ctx)
	if err != nil {
		fmt.Print("GetBeers: ", err)
	}
	for _, beer := range beers {
		fmt.Println(beer)
	}
}

func main() {
	filename := "beers.csv"

	ctx := context.Background()
	cfg := postgres.Config{
		Host:     "localhost",
		Port:     5432,
		DB:       "db",
		Username: "POSTGRES_USER",
		Password: "password",
		MinConns: 1,
		MaxConns: 3,
	}

	pool, err := postgres.NewPool(ctx, cfg)
	if err != nil {
		fmt.Println(err)
	}

	defer pool.Close()

	if err = migrator.Up(pool); err != nil {
		fmt.Println("fillDB: ", err)
	}

	repo := repository.NewBeerPostgres(pool)
	if err = fillDB(ctx, filename, repo); err != nil {
		fmt.Print("fillDB: ", err)
	}
}
