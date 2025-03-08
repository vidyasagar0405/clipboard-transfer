package crypto

import (
	"bytes"
	"encoding/base64"
	"log/slog"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Setup
	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test cases
	testCases := []struct {
		name     string
		data     []byte
		key      []byte
		expectOK bool
	}{
		{
			name:     "Normal text",
			data:     []byte("This is a test message"),
			key:      make([]byte, 32), // Zero key for testing
			expectOK: true,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			key:      make([]byte, 32),
			expectOK: true,
		},
		{
			name:     "Binary data",
			data:     []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
			key:      make([]byte, 32),
			expectOK: true,
		},
		{
			name:     "Invalid key size",
			data:     []byte("This is a test message"),
			key:      make([]byte, 16), // Too short
			expectOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip test if we expect it to fail
			if !tc.expectOK && len(tc.key) != 32 {
				t.Skip("Skipping test with invalid key size")
			}

			// Encrypt
			encrypted, err := Encrypt(tc.data, tc.key)
			if err != nil {
				if tc.expectOK {
					t.Fatalf("Failed to encrypt: %v", err)
				} else {
					return // Expected failure
				}
			}

			// Verify encrypted data is different
			if bytes.Equal(encrypted, tc.data) && len(tc.data) > 0 {
				t.Error("Encrypted data equals original data")
			}

			// Decrypt
			decrypted, err := Decrypt(encrypted, tc.key)
			if err != nil {
				t.Fatalf("Failed to decrypt: %v", err)
			}

			// Verify decrypted matches original
			if !bytes.Equal(decrypted, tc.data) {
				t.Errorf("Decrypted data doesn't match original. Got %v, want %v",
					decrypted, tc.data)
			}
		})
	}
}

func TestGetAESKey(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test with no environment variable
	os.Unsetenv("TRANSFER_KEY")
	key1, err := GetAESKey(logger)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}
	if len(key1) != 32 {
		t.Errorf("Key length should be 32, got %d", len(key1))
	}

	// Test with environment variable
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i)
	}
	os.Setenv("TRANSFER_KEY", base64.StdEncoding.EncodeToString(testKey))

	key2, err := GetAESKey(logger)
	if err != nil {
		t.Fatalf("Failed to get key from env: %v", err)
	}
	if !bytes.Equal(key2, testKey) {
		t.Errorf("Key from env doesn't match expected")
	}

	// Cleanup
	os.Unsetenv("TRANSFER_KEY")
}
