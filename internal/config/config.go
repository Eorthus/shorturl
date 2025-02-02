// Package config предоставляет структуры и функции для конфигурации сервиса.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
)

// ServerConfig содержит настройки HTTP сервера
type ServerConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:"localhost:8080" json:"server_address" `
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080" json:"base_url"`
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

// TLSConfig содержит настройки TLS/HTTPS
type TLSConfig struct {
	EnableHTTPS bool   `env:"ENABLE_HTTPS" envDefault:"false" json:"enable_https"`
	CertFile    string `env:"CERT_FILE" envDefault:"server.crt" json:"cert_file"`
	KeyFile     string `env:"KEY_FILE" envDefault:"server.key" json:"key_file"`
	ConfigFile  string `env:"CONFIG" envDefault:""`
}

// StorageConfig содержит настройки хранилища
type StorageConfig struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"url_storage.json" json:"file_storage_path"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"" json:"database_dsn"`
}

// Config содержит все компоненты конфигурации сервиса
type Config struct {
	Server     ServerConfig
	Storage    StorageConfig
	TLS        TLSConfig
	GRPC       GRPCConfig // добавляем поле для gRPC конфигурации
	ConfigFile string     `env:"CONFIG" envDefault:""`
}

// ConfigBuilder реализует паттерн строителя для конфигурации
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder создает новый экземпляр строителя конфигурации
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: &Config{},
	}
}

// WithServerConfig устанавливает конфигурацию сервера
func (b *ConfigBuilder) WithServerConfig(addr, baseURL string) *ConfigBuilder {
	b.config.Server.ServerAddress = addr
	b.config.Server.BaseURL = baseURL
	return b
}

// WithStorageConfig устанавливает конфигурацию хранилища
func (b *ConfigBuilder) WithStorageConfig(filePath, dbDSN string) *ConfigBuilder {
	b.config.Storage.FileStoragePath = filePath
	b.config.Storage.DatabaseDSN = dbDSN
	return b
}

// В ConfigBuilder добавляем метод для установки доверенной подсети
func (b *ConfigBuilder) WithTrustedSubnet(subnet string) *ConfigBuilder {
	b.config.Server.TrustedSubnet = subnet
	return b
}

// WithTLSConfig устанавливает конфигурацию TLS
func (b *ConfigBuilder) WithTLSConfig(enable bool, certFile, keyFile string) *ConfigBuilder {
	b.config.TLS.EnableHTTPS = enable
	b.config.TLS.CertFile = certFile
	b.config.TLS.KeyFile = keyFile
	return b
}

// FromEnv загружает конфигурацию из переменных окружения
func (b *ConfigBuilder) FromEnv() (*ConfigBuilder, error) {
	if err := env.Parse(&b.config.Server); err != nil {
		return nil, err
	}
	if err := env.Parse(&b.config.Storage); err != nil {
		return nil, err
	}
	if err := env.Parse(&b.config.TLS); err != nil {
		return nil, err
	}
	if err := env.Parse(&b.config.GRPC); err != nil { // Добавляем парсинг GRPC конфига
		return nil, err
	}
	if err := env.Parse(b.config); err != nil {
		return nil, err
	}
	return b, nil
}

// FromFlags загружает конфигурацию из флагов командной строки
func (b *ConfigBuilder) FromFlags() *ConfigBuilder {
	// Флаги для Server
	flag.StringVar(&b.config.Server.ServerAddress, "a", b.config.Server.ServerAddress, "HTTP server address")
	flag.StringVar(&b.config.Server.BaseURL, "b", b.config.Server.BaseURL, "Base address for shortened URL")
	flag.StringVar(&b.config.Server.TrustedSubnet, "t", b.config.Server.TrustedSubnet, "Trusted subnet in CIDR notation")

	// Флаги для Storage
	flag.StringVar(&b.config.Storage.FileStoragePath, "f", b.config.Storage.FileStoragePath, "File storage path")
	flag.StringVar(&b.config.Storage.DatabaseDSN, "d", b.config.Storage.DatabaseDSN, "Database connection string")

	// Флаги для TLS
	flag.BoolVar(&b.config.TLS.EnableHTTPS, "s", b.config.TLS.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&b.config.TLS.CertFile, "cert", b.config.TLS.CertFile, "Path to SSL certificate file")
	flag.StringVar(&b.config.TLS.KeyFile, "key", b.config.TLS.KeyFile, "Path to SSL private key file")

	// Общие флаги
	flag.StringVar(&b.config.ConfigFile, "c", b.config.ConfigFile, "Path to configuration file")
	flag.StringVar(&b.config.ConfigFile, "config", b.config.ConfigFile, "Path to configuration file")

	// Флаги для gRPC
	flag.StringVar(&b.config.GRPC.Address, "grpc-addr", b.config.GRPC.Address, "gRPC server address")
	flag.IntVar(&b.config.GRPC.MaxMessageSize, "grpc-max-size", b.config.GRPC.MaxMessageSize, "Maximum message size for gRPC")
	flag.BoolVar(&b.config.GRPC.EnableReflection, "grpc-reflection", b.config.GRPC.EnableReflection, "Enable reflection for gRPC")

	flag.Parse()
	return b
}

// FromJSON загружает конфигурацию из JSON файла
func (b *ConfigBuilder) FromJSON(filename string) (*ConfigBuilder, error) {
	if filename == "" {
		return b, nil
	}

	// Добавляем расширение .json, если его нет
	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}

	jsonCfg, err := LoadJSON(filename)
	if err != nil {
		return nil, err
	}

	if jsonCfg != nil {
		fmt.Printf("Applying JSON config to builder: %+v\n", jsonCfg)
		b.config.ApplyJSON(jsonCfg)
		fmt.Printf("Config after applying JSON: %+v\n", b.config)
	}

	return b, nil
}

// Build собирает и возвращает готовую конфигурацию
func (b *ConfigBuilder) Build() *Config {
	return b.config
}

// LoadConfig загружает полную конфигурацию используя паттерн строителя
func LoadConfig() (*Config, error) {
	builder := NewConfigBuilder()

	// Загружаем конфигурацию из переменных окружения
	builder, err := builder.FromEnv()
	if err != nil {
		return nil, err
	}

	// Загружаем конфигурацию из флагов командной строки
	builder.FromFlags()

	// Определяем путь к файлу конфигурации
	configFile := os.Getenv("CONFIG")
	if configFile == "" {
		if flag.Lookup("config") != nil {
			configFile = flag.Lookup("config").Value.String()
		}
	}

	// Загружаем конфигурацию из JSON файла если указан
	if configFile != "" {
		builder, err = builder.FromJSON(configFile)
		if err != nil {
			return nil, err
		}
	}

	return builder.Build(), nil
}
