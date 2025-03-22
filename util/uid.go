package util

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecret(length int) (string, error) {
	// Generate a random sequence of bytes
	var bytes = make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode the bytes using Base64 encoding
	secret := base64.URLEncoding.EncodeToString(bytes)

	return secret, nil
}
