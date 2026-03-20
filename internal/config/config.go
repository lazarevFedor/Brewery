package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	//Postgres postgres.Config `env:"POSTGRES"`
	Port string `env:"SERVER_PORT"`
}

func NewConfig() (*Config, error) {
	var config Config
	if err := cleanenv.ReadEnv(&config); err != nil {
		return nil, fmt.Errorf("reading env error: %w", err)
	}

	return &config, nil
}
