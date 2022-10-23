package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"golang.org/x/sync/errgroup"
)

// DB - конфигурация подключения к базе данных.
type DB struct {
	URI             string `env:"DATABASE_URI"` // URI - адрес подключения к базе данных
	RequiredVersion int64  // RequiredVersion - требуемая версия схемы базы данных
}

// Auth - конфигурация авторизации.
type Auth struct {
	SigningKey string        `env:"AUTH_SECRET"` // SigningKey - ключ для подписи токена
	SigningAlg string        // SigningAlg - алгоритм подписи JWT-токена
	TTL        time.Duration `env:"AUTH_TTL"` // TTL - время жизни авторизационного токена
}

// IntegrationAccrual - конфигурация интеграции с системой расчёта начислений.
type IntegrationAccrual struct {
	Address      string        `env:"ACCRUAL_SYSTEM_ADDRESS"`       // Address - адрес системы расчёта начислений
	PollInterval time.Duration `env:"ACCRUAL_SYSTEM_POLL_INTERVAL"` // PollInterval - интервал опроса системы расчёта начислений
	Timeout      time.Duration `env:"ACCRUAL_SYSTEM_TIMEOUT"`       // Timeout - таймаут запросов к системе расчёта начислений
}

type Config struct {
	DB                 DB     // DB - конфигурация подключения к базе данных
	Auth               Auth   // Auth - конфигурация авторизации
	IntegrationAccrual        // IntegrationAccrual - конфигурация интеграции с системой расчёта начислений
	RunAddress         string `env:"RUN_ADDRESS"` // RunAddress - адрес и порт запуска сервиса
}

// FromCLI - конфигурационная функция, которая считывает конфигурацию приложения из переменных окружения.
//
// Флаги командной строки:
//    -a <host:port> - адрес и порт запуска сервиса
//	  -d <dsn>       - адрес подключения к базе данных
//    -r <url>       - адрес системы расчёта начислений
//    -p <duration>  - интервал опроса системы расчёта начислений
//    -m <duration>  - таймаут запросов к системе расчёта начислений
//    -t <duration>  - время жизни авторизационного токена
//
// Если какие-либо значения не заданы в командной строке, то используются значения переданные в cfg.
func FromCLI(cfg *Config) (*Config, error) {
	return fromCLI(cfg, os.Args[1:]...)
}

// fromCLI - логика для FromCLI.
// Вынесена отдельно в целях тестирования.
func fromCLI(cfg *Config, arguments ...string) (*Config, error) {
	// Парсим командную строку
	cli := flag.NewFlagSet("config", flag.ExitOnError)
	cli.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "адрес и порт запуска сервиса")
	cli.StringVar(&cfg.DB.URI, "d", cfg.DB.URI, "адрес подключения к базе данных")
	cli.StringVar(&cfg.IntegrationAccrual.Address, "r", cfg.IntegrationAccrual.Address, "адрес системы расчёта начислений")
	cli.DurationVar(&cfg.IntegrationAccrual.PollInterval, "p", cfg.IntegrationAccrual.PollInterval, "интервал опроса системы расчёта начислений")
	cli.DurationVar(&cfg.IntegrationAccrual.Timeout, "m", cfg.IntegrationAccrual.Timeout, "таймаут запросов к системе расчёта начислений")
	cli.DurationVar(&cfg.Auth.TTL, "t", cfg.Auth.TTL, "время жизни авторизационного токена")
	if err := cli.Parse(arguments); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// FromEnv - конфигурационная функция, которая читывает конфигурацию приложения из переменных окружения.
//
// Переменные окружения:
//    RUN_ADDRESS                  - адрес и порт запуска сервиса
//    DATABASE_URI                 - адрес подключения к базе данных
//    ACCRUAL_SYSTEM_ADDRESS       - адрес системы расчёта начислений
//    ACCRUAL_SYSTEM_TIMEOUT       - таймаут запросов к системе расчёта начислений
//    ACCRUAL_SYSTEM_POLL_INTERVAL - интервал опроса системы расчёта начислений
//    AUTH_TTL                     - время жизни авторизационного токена
//    AUTH_SECRET                  - секретный ключ для подписи авторизационного токена
//
// Если какие-либо переменные окружения не заданы, то используются значения переданные в cfg.
func FromEnv(cfg *Config) (*Config, error) {
	// Получаем параметры из окружения
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	if err = cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate - проверяет конфигурацию на валидность
func (c *Config) validate() error {
	g := &errgroup.Group{}
	g.Go(c.validateAuthSecret)
	g.Go(c.validateServerAddr)
	return g.Wait()
}

// validateServerAddr - проверяет адрес для запуска HTTP-сервера.
func (c *Config) validateServerAddr() error {
	if c.RunAddress == "" {
		return fmt.Errorf("empty server address")
	}
	_, err := net.ResolveTCPAddr("tcp", c.RunAddress)
	if err != nil {
		return fmt.Errorf("invalid server address")
	}
	return nil
}

// validateAuthSecret - проверяет чтобы ключ авторизации был не пустым.
func (c *Config) validateAuthSecret() error {
	if len(c.Auth.SigningKey) == 0 {
		return fmt.Errorf("auth secret not set")
	}
	return nil
}
