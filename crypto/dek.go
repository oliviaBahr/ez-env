package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// AES-256-GCM is used for both DEK and file content encryption
	keySize = 32 // 256 bits
	// Nonce size for GCM mode
	nonceSize = 12
	// GCM tag size
	tagSize = 16
)

// DEKManager contains the functions for encrypting and decrypting DEKs and files, as well as the DEK itself
type DEKManager struct {
	rawKey UnencryptedDEK
}

// EncryptedDEK represents an encrypted DEK
type EncryptedDEK []byte

// UnencryptedDEK represents an unencrypted DEK
type UnencryptedDEK []byte

// NewDEKManager creates a new empty DEKManager instance
func NewDEKManager() *DEKManager {
	return &DEKManager{}
}

// GenerateRawKey generates a new DEK
func (d *DEKManager) GenerateRawKey() error {
	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}
	d.rawKey = UnencryptedDEK(key)
	return nil
}

// EncryptFile encrypts file contents using the DEK
// Returns the encrypted data with a metadata block containing:
// - Version (uint32)
// - Nonce (12 bytes)
// - Encrypted content
func (d *DEKManager) EncryptFile(plaintext []byte) ([]byte, error) {
	if d.rawKey == nil {
		return nil, fmt.Errorf("no DEK found. Must be generated first with GenerateRawKey()")
	}

	block, err := aes.NewCipher(d.rawKey)
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

func DecryptFile(encrypted []byte, key UnencryptedDEK) ([]byte, error) {
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

// EncryptDEK encrypts the DEK using an RSA public key
func (d *DEKManager) EncryptDEK(publicKey *rsa.PublicKey) (EncryptedDEK, error) {
	if d.rawKey == nil {
		return nil, fmt.Errorf("no DEK found. Must be generated first with GenerateRawKey()")
	}

	encrypted, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		d.rawKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt DEK: %w", err)
	}

	return EncryptedDEK(encrypted), nil
}

// DecryptDEK decrypts the DEK using an RSA private key
func DecryptDEK(encryptedDEK EncryptedDEK, privateKey *rsa.PrivateKey) (UnencryptedDEK, error) {
	decrypted, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedDEK,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK: %w", err)
	}

	return UnencryptedDEK(decrypted), nil
}
