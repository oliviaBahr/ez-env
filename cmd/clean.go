package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/oliviaBahr/ez-env/crypto"
)

// Clean encrypts the file content using the shared encryption key
// This is called by Git when files are staged (git add)
// Only called for files that match patterns in .gitattributes
func Clean() error {
	// Read the file content from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Check if the content is already encrypted
	if crypto.IsEncryptedFile(input) {
		// If already encrypted, just pass it through
		if _, err := os.Stdout.Write(input); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	}

	// Get encryption key
	ctx := context.Background()
	keyManager := crypto.NewKeyManager()
	key, err := keyManager.GetOrCreateEncryptionKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt the file content
	encryptedContent, err := crypto.EncryptFile(input, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt content: %w", err)
	}

	// Write the encrypted content to stdout (Git will store this in the index)
	if _, err := os.Stdout.Write(encryptedContent); err != nil {
		return fmt.Errorf("failed to write encrypted content: %w", err)
	}

	return nil
}
