package config

import (
	"crypto/rand"
	"encoding/base64"
)

// randSecret - генерирует случайный ключ.
func randSecret(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
