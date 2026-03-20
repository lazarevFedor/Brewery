package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"Brewery/pkg/postgres"
)


type AppConfig struct {
	Postgres postgres.Config `env:"POSTGRES"`
	Port string `env:"SERVER_PORT"`
}

func NewAppConfig() (*AppConfig, error) {
	var cfg AppConfig
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("reading env error: %w", err)
	}
	return &cfg, nil
}

func FillConfig[T any](cfg *T) (*T, error) {
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("reading env error: %w", err)
	}
	return cfg, nil
}
