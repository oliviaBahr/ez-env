package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/oliviaBahr/ez-env/crypto"
)

// AddFile adds a file to .gitattributes with filter=ezenv and encrypts it using the shared DEK
func AddFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no file specified")
	}

	filePath := args[0]

	// Add the file to .gitattributes
	attrEntry := fmt.Sprintf("%s filter=ezenv\n", filePath)
	if err := os.WriteFile(".gitattributes", []byte(attrEntry), 0644); err != nil {
		return fmt.Errorf("failed to update .gitattributes: %w", err)
	}

	// Load the keyring
	keyring, err := crypto.LoadKeyring()
	if err != nil {
		return fmt.Errorf("failed to load keyring: %w", err)
	}

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Encrypt the file content using the shared DEK
	encryptedContent, err := keyring.DEK.EncryptFile(content)
	if err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	// Write the encrypted content back to the file
	if err := os.WriteFile(filePath, encryptedContent, 0644); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	// Add the file to git
	addCmd := exec.Command("git", "add", filePath)
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add file to git: %w", err)
	}

	return nil
}
