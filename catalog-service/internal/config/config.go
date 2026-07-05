package config

import (
	"errors"
	"fmt"
	"os"

	pkgConfig "github.com/jackietana/ticket-platform/pkg/config"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   pkgConfig.ServerConfig   `yaml:"server"`
	Postgres pkgConfig.PostgresConfig `yaml:"postgres"`
	Minio    MinioConfig              `yaml:"minio"`
}

type MinioConfig struct {
	InternalEndpoint string `yaml:"internal_endpoint"`
	ExternalEndpoint string `yaml:"external_endpoint"`
	ServerPort       string `yaml:"server_port"`
	User             string `yaml:"minio_user"`
	Pass             string `yaml:"minio_pass"`
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
	cfg.Postgres.Name = os.Getenv("APP_CATALOG_DB_NAME")
	cfg.Postgres.User = os.Getenv("APP_DB_USER")
	cfg.Postgres.Pass = os.Getenv("APP_DB_PASS")
	cfg.Postgres.SSLMode = os.Getenv("APP_DB_SSLMODE")

	cfg.Minio.User = os.Getenv("APP_MINIO_USER")
	cfg.Minio.Pass = os.Getenv("APP_MINIO_PASS")
	cfg.Minio.ServerPort = os.Getenv("APP_MINIO_SRV_PORT")

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Server.GRPCPort == "" || c.Server.RESTPort == "" {
		return errors.New("missing server config")
	}

	if c.Postgres.Host == "" || c.Postgres.Port == "" || c.Postgres.Name == "" ||
		c.Postgres.User == "" || c.Postgres.Pass == "" || c.Postgres.SSLMode == "" {
		return errors.New("missing psql config")
	}

	if c.Minio.InternalEndpoint == "" || c.Minio.ExternalEndpoint == "" || c.Minio.ServerPort == "" ||
		c.Minio.User == "" || c.Minio.Pass == "" {
		return errors.New("missing minio config")
	}

	return nil
}
