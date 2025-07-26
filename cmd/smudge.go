package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/oliviaBahr/ez-env/crypto"
)

// Smudge decrypts the file content using the shared encryption key
// This is called by Git when files are checked out (git checkout, git pull)
// Only called for files that match patterns in .gitattributes
func Smudge() error {
	// Read the encrypted file content from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Check if the content is encrypted
	if !crypto.IsEncryptedFile(input) {
		// If not encrypted, just pass it through
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

	// Decrypt the file content
	plaintext, err := crypto.DecryptFile(input, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt content: %w", err)
	}

	// Write the plaintext content to stdout (Git will write this to the working tree)
	if _, err := os.Stdout.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write plaintext content: %w", err)
	}

	return nil
}
