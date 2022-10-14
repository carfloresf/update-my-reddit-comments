package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		HTTP   `yaml:"http"`
		DB     `yaml:"db"`
		Reddit `yaml:"reddit"`
	}

	HTTP struct {
		Addr string `env-required:"true" yaml:"address" env:"HTTP_ADDR"`
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	DB struct {
		DBFile string `env-required:"true" yaml:"db_file" env:"DB_FILE"`
	}

	Reddit struct {
		Username string `env-required:"true" yaml:"username" env:"REDDIT_USERNAME"`
		Password string `env-required:"true" yaml:"password" env:"REDDIT_PASSWORD"`
		ClientID string `env-required:"true" yaml:"client_id" env:"REDDIT_CLIENT_ID"`
		Secret   string `env-required:"true" yaml:"secret" env:"REDDIT_SECRET"`
	}
)

func ReadConfig(path string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(path, cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
