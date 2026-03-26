// Package config содержит функции для считывания конфигурационны переменных для приложения
package config

import (
	"Brewery/pkg/postgres"

	"github.com/ilyakaznacheev/cleanenv"
)

type AppConfig struct {
	Postgres postgres.Config `env-prefix:"POSTGRES_"`
	Port     string          `env:"SERVER_PORT"`
}

func NewAppConfig() *AppConfig {
	var cfg AppConfig
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil
	}
	return &cfg
}
