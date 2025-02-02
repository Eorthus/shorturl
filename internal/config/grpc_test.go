package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCConfigEnv(t *testing.T) {
	// Сохраняем оригинальные значения
	envVars := map[string]string{
		"GRPC_ADDRESS":           ":12345",
		"GRPC_MAX_MESSAGE_SIZE":  "8388608",
		"GRPC_ENABLE_REFLECTION": "true",
	}

	for k, v := range envVars {
		oldVal := os.Getenv(k)
		t.Cleanup(func() {
			os.Setenv(k, oldVal)
		})
		os.Setenv(k, v)
	}

	builder := NewConfigBuilder()
	builder, err := builder.FromEnv()
	require.NoError(t, err)

	cfg := builder.Build()

	assert.Equal(t, ":12345", cfg.GRPC.Address)
	assert.Equal(t, 8388608, cfg.GRPC.MaxMessageSize)
	assert.True(t, cfg.GRPC.EnableReflection)
}

func TestGRPCConfigDefaults(t *testing.T) {
	// Очищаем все переменные окружения, которые могут повлиять на тест
	envVars := []string{
		"GRPC_ADDRESS",
		"GRPC_MAX_MESSAGE_SIZE",
		"GRPC_ENABLE_REFLECTION",
	}

	for _, env := range envVars {
		oldVal := os.Getenv(env)
		t.Cleanup(func() {
			os.Setenv(env, oldVal)
		})
		os.Unsetenv(env)
	}

	builder := NewConfigBuilder()
	builder, err := builder.FromEnv()
	require.NoError(t, err)

	cfg := builder.Build()

	assert.Equal(t, ":50051", cfg.GRPC.Address, "Default GRPC address should be :50051")
	assert.Equal(t, 4194304, cfg.GRPC.MaxMessageSize, "Default max message size should be 4MB")
	assert.False(t, cfg.GRPC.EnableReflection, "Reflection should be disabled by default")
}

func TestGRPCConfigJSON(t *testing.T) {
	// Сначала загружаем дефолтные значения
	builder := NewConfigBuilder()
	builder, err := builder.FromEnv()
	require.NoError(t, err)

	// Затем применяем JSON конфигурацию
	tempFile, err := os.CreateTemp("", "config*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	config := `{
        "grpc": {
            "address": ":7777",
            "max_message_size": 1048576,
            "enable_reflection": true
        }
    }`
	err = os.WriteFile(tempFile.Name(), []byte(config), 0644)
	require.NoError(t, err)

	builder, err = builder.FromJSON(tempFile.Name())
	require.NoError(t, err, "Failed to load JSON config")

	// Получим сырой JSON для проверки
	rawConfig, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err, "Failed to read config file")
	t.Logf("Raw config: %s", string(rawConfig))

	cfg := builder.Build()
	t.Logf("Resulting config: %+v", cfg.GRPC)
	assert.Equal(t, ":7777", cfg.GRPC.Address)
	assert.Equal(t, 1048576, cfg.GRPC.MaxMessageSize)
	assert.True(t, cfg.GRPC.EnableReflection)
}
