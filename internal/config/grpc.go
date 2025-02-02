package config

// GRPCConfig содержит настройки gRPC сервера
type GRPCConfig struct {
	// Адрес gRPC сервера
	Address string `env:"GRPC_ADDRESS" envDefault:":50051" json:"address"`
	// Максимальный размер сообщения (default: 4MB)
	MaxMessageSize int `env:"GRPC_MAX_MESSAGE_SIZE" envDefault:"4194304" json:"max_message_size"`
	// Включить отражение (reflection) для gRPC
	EnableReflection bool `env:"GRPC_ENABLE_REFLECTION" envDefault:"false" json:"enable_reflection"`
}

// WithGRPCConfig устанавливает конфигурацию gRPC сервера
func (b *ConfigBuilder) WithGRPCConfig(addr string) *ConfigBuilder {
	b.config.GRPC.Address = addr
	return b
}

// WithGRPCLimits устанавливает ограничения для gRPC сервера
func (b *ConfigBuilder) WithGRPCLimits(maxMessageSize int) *ConfigBuilder {
	b.config.GRPC.MaxMessageSize = maxMessageSize
	return b
}
