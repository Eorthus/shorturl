package config

import (
	"flag"
)

// Config структура содержит конфигурационные параметры приложения
type Config struct {
	ServerAddress string
	BaseURL       string
}

// ParseFlags инициализирует конфигурацию из флагов командной строки
func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base address for shortened URL")

	flag.Parse()

	return cfg
}
