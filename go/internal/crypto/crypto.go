package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
)

// GetAESKey gets or generates AES key (must be 32 bytes)
func GetAESKey(logger *slog.Logger) ([]byte, error) {
	keyBase64 := os.Getenv("TRANSFER_KEY")
	if keyBase64 == "" {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			return nil, err
		}
		encodedKey := base64.StdEncoding.EncodeToString(key)
		logger.Info("Generated new encryption key", "key", encodedKey)
		return key, nil
	}
	return base64.StdEncoding.DecodeString(keyBase64)
}

// Encrypt data using AES-GCM
func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesGCM.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// Decrypt data using AES-GCM
func Decrypt(data []byte, key []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, errors.New("invalid data")
	}
	nonce := data[:12]
	ciphertext := data[12:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
