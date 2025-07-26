package crypto

import (
	"context"
	"fmt"

	"github.com/oliviaBahr/ez-env/github"
)

// KeyManager handles encryption key storage and retrieval
type KeyManager struct{}

// NewKeyManager creates a new key manager
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// GetOrCreateEncryptionKey retrieves the existing encryption key or creates a new one
func (km *KeyManager) GetOrCreateEncryptionKey(ctx context.Context) ([]byte, error) {
	fmt.Println("Retrieving encryption key via GitHub workflow...")

	// First try to get the existing key via workflow
	key, err := github.GetEncryptionKey(ctx)
	if err != nil {
		// If getting the key fails, create a new one
		fmt.Println("No existing encryption key found. Creating new key...")
		key, err = GenerateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}

		// Store the new key in GitHub secrets
		if err := github.StoreEncryptionKey(ctx, key); err != nil {
			return nil, fmt.Errorf("failed to store encryption key: %w", err)
		}

		fmt.Println("✓ New encryption key created and stored in GitHub repository secrets")
	} else {
		fmt.Println("✓ Existing encryption key retrieved from GitHub repository secrets")
	}

	return key, nil
}
