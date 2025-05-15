package filter

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oliviaBahr/ez-env/crypto"
)

// Clean encrypts the file content using the shared DEK and writes a metadata block
func Clean() error {
	// Read the file content from stdin
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Load the keyring
	keyring, err := crypto.LoadKeyring()
	if err != nil {
		return fmt.Errorf("failed to load keyring: %w", err)
	}

	// Encrypt the file content using the shared DEK
	encryptedContent, err := keyring.DEK.EncryptFile(input)
	if err != nil {
		return fmt.Errorf("failed to encrypt content: %w", err)
	}

	// Write the encrypted content to stdout
	if _, err := os.Stdout.Write(encryptedContent); err != nil {
		return fmt.Errorf("failed to write encrypted content: %w", err)
	}

	return nil
}
