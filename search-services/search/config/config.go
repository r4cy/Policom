package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type SearchServer struct {
	Address  string        `yaml:"address" env:"SEARCH_ADDRESS" env-default:"localhost:83"`
	IndexTTL time.Duration `yaml:"index_ttl" env:"INDEX_TTL" env-default:"20s"`
}

type Config struct {
	LogLevel     string       `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	SearchServer SearchServer `yaml:"search_server" `
	DBAddress    string       `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	WordsAddress string       `yaml:"words_address" env:"WORDS_ADDRESS" env-default:"localhost:81"`
	BrokerAddress string `yaml:"broker_address" env:"BROKER_ADDRESS" env-default:"localhost:4222"`
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
