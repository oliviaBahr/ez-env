package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"crypto/rsa"
)

// LoadLocalSSHPrivateKey loads the local SSH private key
func LoadLocalSSHPrivateKey() (*rsa.PrivateKey, error) {
	// Get the SSH key from the SSH agent
	cmd := exec.Command("ssh-add", "-L")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key from agent: %w", err)
	}

	// Verify we have at least one key
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, fmt.Errorf("no SSH keys found in agent")
	}

	// Get the private key path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Try common SSH key locations
	keyPaths := []string{
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
	}

	var privateKeyPath string
	for _, path := range keyPaths {
		if _, err := os.Stat(path); err == nil {
			privateKeyPath = path
			break
		}
	}

	if privateKeyPath == "" {
		return nil, fmt.Errorf("no SSH private key found")
	}

	// Read the private key
	privateKeyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Parse the private key
	privateKey, err := ParseSSHPrivateKey(privateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}
