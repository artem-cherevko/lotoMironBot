package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken  string `env:"BOT_TOKEN,required"`
	DBDsn     string `env:"DB_DSN,required"`
	RedisAddr string `env:"REDIS_ADDR,required"`
	AdminID   string `env:"ADMIN_ID"`
	PPKey     string `env:"PP_KEY,required"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
