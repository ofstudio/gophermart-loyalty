package config

import (
	"time"
)

// Default - конфигурационная функция, возвращает конфигурацию по умолчанию.
func Default(_ *Config) (*Config, error) {
	secret, err := randSecret(64)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		DB: DB{
			RequiredVersion: 1,
		},
		Auth: Auth{
			SigningAlg: "HS512",
			TTL:        30 * 24 * time.Hour,
			SigningKey: secret,
		},
		RunAddress: "0.0.0.0:8080",
	}

	return &cfg, nil
}
