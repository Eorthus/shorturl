// internal/utils/utils_test.go

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateShortID(t *testing.T) {
	id := GenerateShortID()

	// Проверяем, что длина сгенерированного идентификатора составляет 8 символов
	assert.Equal(t, 8, len(id), "Длина сгенерированного ID должна быть 8 символов")

	// Можно добавить проверку, что ID уникален, сгенерировав несколько значений
	secondID := GenerateShortID()
	assert.NotEqual(t, id, secondID, "Два идентификатора должны быть разными")
}
