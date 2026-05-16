package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"5s"`
}

type Config struct {
	LogLevel          string        `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	SearchConcurrency int           `yaml:"search_concurrency" env:"SEARCH_CONCURRENCY" env-default:"1"`
	SearchRate        int           `yaml:"search_rate" env:"SEARCH_RATE" env-default:"100"`
	HTTPConfig        HTTPConfig    `yaml:"api_server"`
	WordsAddress      string        `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"words:81"`
	UpdateAddress     string        `yaml:"update_address" env:"UPDATE_ADDRESS" env-default:"update:82"`
	SearchAddress     string        `yaml:"search_address" env:"SEARCH_ADDRESS" env-default:"search:83"`
	TokenTTL          time.Duration `yaml:"token_ttl" env:"TOKEN_TTL" env-default:"24h"`
	DBAddress         string        `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:1234"`
	JWTSecret         string        `yaml:"jwt_secret" env:"JWT_SECRET" env-default:"GOGOGOOOPHER"`
	AdminName         string        `yaml:"admin_name" env:"ADMIN_USER" env-default:"admin"`
	AdminPassword     string        `yaml:"admin_password" env:"ADMIN_PASSWORD" env-default:"pass"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("Read config from file %q failed with err: %v", configPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("Read config from environment failed with err: %v", err)
		}
	}
	return cfg
}
