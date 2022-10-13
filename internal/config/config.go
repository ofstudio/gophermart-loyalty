package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
	"time"
)

// DB - конфигурация подключения к базе данных.
type DB struct {
	URI             string `env:"DATABASE_URI"` // URI - адрес подключения к базе данных
	RequiredVersion int64  // RequiredVersion - требуемая версия схемы базы данных
}

// Auth - конфигурация авторизации.
type Auth struct {
	JWTSigningAlg string        // JWTSigningAlg - алгоритм подписи JWT-токена
	TTL           time.Duration // TTL - время жизни авторизационного токена
	Secret        string        `env:"AUTH_SECRET"` // Secret - секрет для подписи токена
}

type Config struct {
	DB                   DB     // DB - конфигурация подключения к базе данных
	Auth                 Auth   // Auth - конфигурация авторизации
	RunAddress           string `env:"RUN_ADDRESS"`            // RunAddress - адрес и порт запуска сервиса
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"` // AccrualSystemAddress - адрес системы расчёта начислений
}

// FromCLI - конфигурационная функция, которая считывает конфигурацию приложения из переменных окружения.
//
// Флаги командной строки:
//    -a <host:port> - адрес и порт запуска сервиса
//	  -d <dsn>       - адрес подключения к базе данных
//    -r <url>       - адрес системы расчёта начислений
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
//    RUN_ADDRESS            - адрес и порт запуска сервиса
//    DATABASE_URI           - адрес подключения к базе данных
//    ACCRUAL_SYSTEM_ADDRESS - адрес системы расчёта начислений
//    AUTH_TTL               - время жизни авторизационного токена
//	  AUTH_SECRET            - секретный ключ для подписи авторизационного токена
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
	// todo validate database uri contains postgres scheme
	// todo validate accrual system address contains http/https scheme
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
	if len(c.Auth.Secret) == 0 {
		return fmt.Errorf("auth secret not set")
	}
	return nil
}
