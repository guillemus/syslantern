package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"syslantern/validate"
)

type Config struct {
	DBPath       string `validate:"required"`
	Port         string `validate:"required"`
	AssetVersion string `validate:"required"`
	Dev          bool
}

func ParseConfig() Config {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic(fmt.Sprintf("load .env file: %v", err))
	}

	cfg := Config{
		DBPath:       GetEnvOr("DB_PATH", "./tmp/syslantern.db"),
		Port:         GetEnvOr("PORT", "3000"),
		AssetVersion: fmt.Sprintf("%d", time.Now().Unix()),
		Dev:          os.Getenv("DEV") == "true",
	}

	if err := validate.V.Struct(cfg); err != nil {
		panic(fmt.Sprintf("validate env vars: %v", err))
	}

	return cfg
}

func GetEnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
