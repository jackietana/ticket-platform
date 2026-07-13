package config

import (
	"errors"
	"fmt"
	"os"

	pkgconfig "github.com/jackietana/ticket-platform/pkg/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server         pkgconfig.ServerConfig      `yaml:"server"`
	Postgres       pkgconfig.PostgresConfig    `yaml:"postgres"`
	Redis          pkgconfig.RedisConfig       `yaml:"redis"`
	Minio          pkgconfig.MinioConfig       `yaml:"minio"`
	AuthService    pkgconfig.AuthServiceConfig `yaml:"auth-service"`
	CatalogService CatalogServiceConfig        `yaml:"catalog-service"`
	Rabbitmq       RabbitmqConfig              `yaml:"rabbitmq"`
}

type CatalogServiceConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type RabbitmqConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
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
	cfg.Postgres.Name = os.Getenv("APP_ORDER_DB_NAME")
	cfg.Postgres.User = os.Getenv("APP_DB_USER")
	cfg.Postgres.Pass = os.Getenv("APP_DB_PASS")
	cfg.Postgres.SSLMode = os.Getenv("APP_DB_SSLMODE")

	cfg.Redis.Host = os.Getenv("APP_REDIS_HOST")
	cfg.Redis.Pass = os.Getenv("APP_REDIS_PASS")

	cfg.Minio.User = os.Getenv("APP_MINIO_USER")
	cfg.Minio.Pass = os.Getenv("APP_MINIO_PASS")

	cfg.Rabbitmq.User = os.Getenv("APP_RABBITMQ_USER")
	cfg.Rabbitmq.Pass = os.Getenv("APP_RABBITMQ_PASS")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetCatalogServiceEndpoint() string {
	return fmt.Sprintf("%s:%s", c.CatalogService.Host, c.CatalogService.Port)
}

func (c *Config) GetRabbitmqEndpoint() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.Rabbitmq.User, c.Rabbitmq.Pass, c.Rabbitmq.Host, c.Rabbitmq.Port)
}

func (c *Config) validate() error {
	if c.Server.GRPCPort == "" || c.Server.RESTPort == "" {
		return errors.New("missing server config")
	}

	if c.Postgres.Host == "" || c.Postgres.Port == "" || c.Postgres.Name == "" ||
		c.Postgres.User == "" || c.Postgres.Pass == "" || c.Postgres.SSLMode == "" {
		return errors.New("missing psql config")
	}

	if c.Redis.Host == "" || c.Redis.Port == "" || c.Redis.Pass == "" {
		return errors.New("missing redis config")
	}

	if c.Minio.InternalEndpoint == "" || c.Minio.ExternalEndpoint == "" || c.Minio.ServerPort == "" ||
		c.Minio.User == "" || c.Minio.Pass == "" {
		return errors.New("missing minio config")
	}

	if c.AuthService.Host == "" || c.AuthService.Port == "" {
		return errors.New("missing auth-service config")
	}

	if c.CatalogService.Host == "" || c.CatalogService.Port == "" {
		return errors.New("missing catalog-service config")
	}

	if c.Rabbitmq.Host == "" || c.Rabbitmq.Port == "" || c.Rabbitmq.User == "" || c.Rabbitmq.Pass == "" {
		return errors.New("missing rabbitmq config")
	}

	return nil
}
