package config

import (
	"errors"
	"fmt"
	"os"

	pkgconfig "github.com/jackietana/ticket-platform/pkg/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   pkgconfig.ServerConfig   `yaml:"server"`
	Postgres pkgconfig.PostgresConfig `yaml:"postgres"`
	Redis    pkgconfig.RedisConfig    `yaml:"redis"`
	Salt     string                   `yaml:"hash_salt"`
}

func NewConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	cfg.Postgres.Host = os.Getenv("APP_DB_HOST")
	cfg.Postgres.Name = os.Getenv("APP_AUTH_DB_NAME")
	cfg.Postgres.User = os.Getenv("APP_DB_USER")
	cfg.Postgres.Pass = os.Getenv("APP_DB_PASS")
	cfg.Postgres.SSLMode = os.Getenv("APP_DB_SSLMODE")

	cfg.Redis.Host = os.Getenv("APP_REDIS_HOST")
	cfg.Redis.Pass = os.Getenv("APP_REDIS_PASS")
	cfg.Salt = os.Getenv("APP_HASH_SALT")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func NewTestConfig() *Config {
	return &Config{
		Postgres: pkgconfig.PostgresConfig{Host: "localhost", Name: "test_db", User: "test_user",
			Pass: "test_pass", Port: "5432", SSLMode: "disable"},
		Redis: pkgconfig.RedisConfig{Port: "6379", Pass: "test_pass"},
		Salt:  "test_salt_0123456789",
	}
}

func (c *Config) validate() error {
	if c.Postgres.Host == "" || c.Postgres.Port == "" || c.Postgres.Name == "" ||
		c.Postgres.User == "" || c.Postgres.Pass == "" || c.Postgres.SSLMode == "" {
		return errors.New("missing psql config")
	}

	if c.Redis.Host == "" || c.Redis.Port == "" || c.Redis.Pass == "" {
		return errors.New("missing redis config")
	}

	if c.Salt == "" {
		return errors.New("missing hash salt")
	}

	return nil
}
