package crypto

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olivia/ezenv/github"
)

// Keyring represents the structure of the .gitenv_keyring file
type Keyring struct {
	Collaborators []github.Collaborator `json:"collaborators"`
}

// Create creates a new .gitenv_keyring file with the given collaborators
func Create(collaborators []github.Collaborator) error {
	keyring := Keyring{
		Collaborators: collaborators,
	}

	data, err := json.MarshalIndent(keyring, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keyring: %w", err)
	}

	if err := os.WriteFile(".gitenv_keyring", data, 0644); err != nil {
		return fmt.Errorf("failed to write keyring file: %w", err)
	}

	return nil
}

// Load loads the .gitenv_keyring file
func Load() (*Keyring, error) {
	data, err := os.ReadFile(".gitenv_keyring")
	if err != nil {
		return nil, fmt.Errorf("failed to read keyring file: %w", err)
	}

	var keyring Keyring
	if err := json.Unmarshal(data, &keyring); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keyring: %w", err)
	}

	return &keyring, nil
}

// Update updates the keyring with new collaborators
func Update(collaborators []github.Collaborator) error {
	return Create(collaborators)
}
