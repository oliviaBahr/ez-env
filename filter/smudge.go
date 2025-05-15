package filter

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oliviaBahr/ez-env/crypto"
	"github.com/oliviaBahr/ez-env/ssh"
)

// Smudge decrypts the file content using the shared DEK and replaces it with plaintext in the working tree.
func Smudge() error {
	// Read the encrypted file content from stdin
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Load the keyring
	keyring, err := crypto.LoadKeyring()
	if err != nil {
		return fmt.Errorf("failed to load keyring: %w", err)
	}

	// Get the local SSH private key
	privateKey, err := ssh.LoadLocalSSHPrivateKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	// Decrypt the DEK using the private key
	dek, err := keyring.GetDecryptedDEK(privateKey, "your-github-username")
	if err != nil {
		return fmt.Errorf("failed to decrypt DEK: %w", err)
	}

	// Decrypt the file content using the DEK
	plaintext, err := crypto.DecryptFile(input, dek)
	if err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Write the plaintext content to stdout
	if _, err := os.Stdout.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write plaintext content: %w", err)
	}

	return nil
}
