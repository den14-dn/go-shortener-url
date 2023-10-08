// Package sign is designed to sign user data and validate it.
package sign

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const sizeKey = 16

var key []byte

func signData(data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// ValidateID The user ID is validated.
func ValidateID(value string) bool {
	data, err := hex.DecodeString(value)
	if err != nil {
		return false
	}
	sign := signData(data[:16])
	return hmac.Equal(sign, data[16:])
}

// UserID A new user ID is created.
func UserID() string {
	id, _ := generateRandom(sizeKey)
	return hex.EncodeToString(append(id, signData(id)...))
}

func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
