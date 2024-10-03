package config

import (
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	LogDir  string
	BaseURL string
	APIKey  string
}

func NewConfig() *Config {
	c := &Config{
		LogDir:  filepath.Join("/tmp/", time.Now().Local().Format("2006-01-02T15:04")+"-flowline.log"),
		BaseURL: getEnv("BASE_URL", ""),
		APIKey:  getEnv("API_KEY", ""),
	}

	return c
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
