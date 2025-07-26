package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// AES-256-GCM encryption constants
	keySize   = 32 // 256 bits
	nonceSize = 12
	tagSize   = 16
)

// GenerateEncryptionKey generates a new AES-256 encryption key
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// EncryptFile encrypts file contents using AES-256-GCM
// Returns the encrypted data with metadata:
// - Version (uint32)
// - Nonce (12 bytes)
// - Encrypted content
func EncryptFile(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the content
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Create the final output with metadata
	// Format: [version(4)][nonce(12)][ciphertext]
	output := make([]byte, 4+nonceSize+len(ciphertext))
	binary.BigEndian.PutUint32(output[0:4], 1) // Version 1
	copy(output[4:4+nonceSize], nonce)
	copy(output[4+nonceSize:], ciphertext)

	return output, nil
}

// DecryptFile decrypts file contents using AES-256-GCM
func DecryptFile(encrypted []byte, key []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", keySize, len(key))
	}

	if len(encrypted) < 4+nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Parse version
	version := binary.BigEndian.Uint32(encrypted[0:4])
	if version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	// Extract nonce and ciphertext
	nonce := encrypted[4 : 4+nonceSize]
	ciphertext := encrypted[4+nonceSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// IsEncryptedFile checks if a file appears to be encrypted by ez-env
func IsEncryptedFile(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	version := binary.BigEndian.Uint32(data[0:4])
	return version == 1
}
