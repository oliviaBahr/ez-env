package cmd

import (
	"context"
	"fmt"

	"github.com/oliviaBahr/ez-env/crypto"
	"github.com/oliviaBahr/ez-env/github"
)

// UpdateKeys re-fetches collaborator list and SSH keys, updates .gitenv_keyring, and encrypts the shared DEK to new public keys
func UpdateKeys() error {
	// Get collaborators and their SSH keys
	collaborators, err := github.GetRepositoryCollaborators(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get collaborators: %w", err)
	}

	// Load the keyring
	keyring, err := crypto.LoadKeyring()
	if err != nil {
		return fmt.Errorf("failed to load keyring: %w", err)
	}

	// Update keyring with new collaborators
	if err := keyring.UpdateCollaborators(collaborators); err != nil {
		return fmt.Errorf("failed to update keyring with collaborators: %w", err)
	}

	// Generate encrypted DEKs for any new collaborators
	if err := keyring.GenerateEncryptedDEKs(); err != nil {
		return fmt.Errorf("failed to generate encrypted DEKs: %w", err)
	}

	fmt.Println("Collaborator keys updated successfully!")
	return nil
}
