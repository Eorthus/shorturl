package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	t.Run("GenerateUserID", func(t *testing.T) {
		userID1 := GenerateUserID()
		userID2 := GenerateUserID()
		assert.NotEqual(t, userID1, userID2, "Generated user IDs should be unique")
	})

	t.Run("CreateAuthCookie", func(t *testing.T) {
		userID := GenerateUserID()
		cookie := CreateAuthCookie(userID)

		assert.Equal(t, "auth", cookie.Name)
		assert.NotEmpty(t, cookie.Value)
		assert.True(t, cookie.Expires.After(time.Now()))
		assert.True(t, cookie.HttpOnly)
		assert.True(t, cookie.Secure)
		assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	})

	t.Run("VerifyAuthCookie", func(t *testing.T) {
		userID := GenerateUserID()
		cookie := CreateAuthCookie(userID)

		verifiedUserID, valid := VerifyAuthCookie(cookie)
		assert.True(t, valid)
		assert.Equal(t, userID, verifiedUserID)

		invalidCookie := &http.Cookie{Name: "auth", Value: "invalid"}
		_, valid = VerifyAuthCookie(invalidCookie)
		assert.False(t, valid)
	})

	t.Run("ExpiredCookie", func(t *testing.T) {
		// Тест с истекшей кукой
		userID := GenerateUserID()
		expiredCookie := CreateAuthCookie(userID)
		expiredCookie.Expires = time.Now().Add(-1 * time.Hour)
		expiredCookie.Value = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d.%s", userID, time.Now().Add(-1*time.Hour).Unix(), createSignature(fmt.Sprintf("%s:%d", userID, time.Now().Add(-1*time.Hour).Unix())))))
		_, valid := VerifyAuthCookie(expiredCookie)
		assert.False(t, valid, "Verification of expired cookie should return false")
	})
}
