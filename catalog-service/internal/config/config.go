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
	User             string
	Pass             string
}

func NewConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	cfg.Postgres = pkgConfig.NewPostgresConfig()

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	cfg.Minio.User = os.Getenv("APP_MINIO_USER")
	cfg.Minio.Pass = os.Getenv("APP_MINIO_PASS")

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
