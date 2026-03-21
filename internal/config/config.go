package config

import (
	"fmt"

	"Brewery/pkg/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Postgres postgres.Config `env:"POSTGRES"`
	Port     string          `env:"SERVER_PORT"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func FillConfig[T any](cfg *T) (*T, error) {
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("reading env error: %w", err)
	}

	return cfg, nil
}
