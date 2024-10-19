package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

var secretKey = []byte("gopher4test5key0")

func GenerateUserID() string {
	return uuid.New().String()
}

func CreateAuthCookie(userID string) *http.Cookie {
	expiration := time.Now().Add(24 * time.Hour)
	value := fmt.Sprintf("%s:%d", userID, expiration.Unix())
	signature := createSignature(value)
	cookie := &http.Cookie{
		Name:     "auth",
		Value:    base64.URLEncoding.EncodeToString([]byte(value + "." + signature)),
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	return cookie
}

func VerifyAuthCookie(cookie *http.Cookie) (string, bool) {
	decodedValue, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return "", false
	}

	parts := strings.Split(string(decodedValue), ".")
	if len(parts) != 2 {
		return "", false
	}

	value, signature := parts[0], parts[1]
	if createSignature(value) != signature {
		return "", false
	}

	valueParts := strings.Split(value, ":")
	if len(valueParts) != 2 {
		return "", false
	}

	userID, expStr := valueParts[0], valueParts[1]
	exp, err := parseInt64(expStr)
	if err != nil {
		return "", false
	}

	if time.Now().Unix() > exp {
		return "", false
	}

	return userID, true
}

func createSignature(value string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(value))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func split(s, sep string) []string {
	result := make([]string, 0, 2)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
			if len(result) == 2 {
				break
			}
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func parseInt64(s string) (int64, error) {
	var result int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid character in number: %c", ch)
		}
		result = result*10 + int64(ch-'0')
	}
	return result, nil
}
