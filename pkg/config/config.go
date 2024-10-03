package config

import (
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	LogDir  string
	BaseURL string
	APIKey  string
	Client  http.Client
}

func NewConfig() *Config {
	c := &Config{
		LogDir:  filepath.Join("/tmp/", time.Now().Local().Format("2006-01-02T15:04")+"-flowline.log"),
		BaseURL: getEnv("BASE_URL", ""),
		APIKey:  getEnv("API_KEY", ""),
		Client:  http.Client{},
	}

	return c
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
