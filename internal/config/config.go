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
	// DatabaseURI - адрес подключения к базе данных
	DatabaseURI     string `env:"DATABASE_URI"`
	RequiredVersion int64
}

type Config struct {
	// RunAddress - адрес и порт запуска сервиса
	RunAddress string `env:"RUN_ADDRESS"`

	// DB - конфигурация подключения к базе данных
	DB DB

	// AccrualSystemAddress - адрес системы расчёта начислений
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

	// AuthTTL - время жизни авторизационного токена
	AuthTTL time.Duration `env:"AUTH_TTL"`

	// AuthSecret - секретный ключ для подписи авторизационного токена
	AuthSecret string `env:"AUTH_SECRET,unset"`

	// PasswdMinLen - минимальная длина пароля
	PasswdMinLen int
}

// Default - конфигурационная функция, возвращает конфигурацию по умолчанию.
func Default(_ *Config) (*Config, error) {
	secret, err := randSecret(64)
	if err != nil {
		return nil, err
	}
	cfg := Config{
		RunAddress: "0.0.0.0:8080",
		DB: DB{
			RequiredVersion: 1,
		},
		AuthTTL:      time.Minute * 60 * 24 * 30,
		AuthSecret:   secret,
		PasswdMinLen: 8,
	}
	return &cfg, nil
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
	cli.StringVar(&cfg.DB.DatabaseURI, "d", cfg.DB.DatabaseURI, "адрес подключения к базе данных")
	cli.DurationVar(&cfg.AuthTTL, "t", cfg.AuthTTL, "время жизни авторизационного токена")
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
	if len(c.AuthSecret) == 0 {
		return fmt.Errorf("auth secret not set")
	}
	return nil
}
