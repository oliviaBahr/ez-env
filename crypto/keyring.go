package crypto

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/oliviaBahr/ez-env/ssh"
)

const (
	// KeyringFile is the name of the file that stores the keyring
	KeyringFile = ".gitenv_keyring"
)

// Collaborator represents a GitHub collaborator
type Collaborator struct {
	SSHKeys      []string     `json:"ssh_keys"`
	EncryptedDEK EncryptedDEK `json:"encrypted_dek"` // Single encrypted DEK for this collaborator
}

// Collaborators represents a map of collaborator logins to their information
type Collaborators map[string]*Collaborator

// Keyring represents the keyring file
type Keyring struct {
	// Map of collaborator login to their information
	Collaborators Collaborators `json:"collaborators"`
	// DEK is the Data Encryption Key used for file encryption
	DEK *DEKManager
}

// NewKeyring creates a new keyring or loads an existing one
func NewKeyring() *Keyring {
	// Try to load existing keyring
	keyring, err := LoadKeyring()
	if err == nil {
		keyring.DEK = NewDEKManager()
		return keyring
	}

	// If loading fails, create a new keyring
	return &Keyring{
		Collaborators: make(Collaborators),
		DEK:           NewDEKManager(),
	}
}

// Save saves the keyring to disk
func (k *Keyring) Save() error {
	// Marshal the keyring to JSON
	data, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal keyring: %w", err)
	}

	// Write the keyring to disk
	if err := os.WriteFile(KeyringFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write keyring: %w", err)
	}

	return nil
}

// LoadKeyring loads the keyring from disk
func LoadKeyring() (*Keyring, error) {
	// Read the keyring file
	data, err := os.ReadFile(KeyringFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read keyring: %w", err)
	}

	// Unmarshal the keyring
	var keyring Keyring
	if err := json.Unmarshal(data, &keyring); err != nil {
		return nil, fmt.Errorf("failed to parse keyring: %w", err)
	}

	return &keyring, nil
}

// AddCollaborator adds a collaborator to the keyring
func (k *Keyring) AddCollaborator(login string, sshKeys []string) {
	k.Collaborators[login] = &Collaborator{
		SSHKeys: sshKeys,
	}
	k.Save()
}

// RemoveCollaborator removes a collaborator from the keyring
func (k *Keyring) RemoveCollaborator(login string) {
	delete(k.Collaborators, login)
	k.Save()
}

// GenerateEncryptedDEKs generates encrypted DEKs for each collaborator
func (k *Keyring) GenerateEncryptedDEKs() error {
	// Encrypt the DEK for each collaborator
	for login, collab := range k.Collaborators {
		// Try each SSH key until we find one that works
		for _, keyStr := range collab.SSHKeys {
			key, err := ssh.ParseSSHPublicKey([]byte(keyStr))
			if err != nil {
				// Skip invalid keys but continue with others
				continue
			}

			// Encrypt the DEK with the public key
			encryptedKey, err := k.DEK.EncryptDEK(key)
			if err != nil {
				// Skip keys that fail encryption but continue with others
				continue
			}

			// Store the encrypted DEK
			collab.EncryptedDEK = encryptedKey
			break // Use the first valid key that successfully encrypts
		}

		// If no valid keys were found for this collaborator
		if len(collab.EncryptedDEK) == 0 {
			return fmt.Errorf("no valid SSH keys found for collaborator %s", login)
		}
	}

	return nil
}

// GetDecryptedDEK retrieves and decrypts the DEK using the provided private key and login
func (k *Keyring) GetDecryptedDEK(privateKey *rsa.PrivateKey, login string) (UnencryptedDEK, error) {
	// Get the collaborator
	collab, exists := k.Collaborators[login]
	if !exists {
		return nil, fmt.Errorf("no collaborator found for login %s", login)
	}

	// Try to decrypt the DEK
	dek, err := DecryptDEK(collab.EncryptedDEK, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt DEK with provided private key: %w", err)
	}

	return dek, nil
}

// UpdateCollaborators updates the keyring with new collaborator information
func (k *Keyring) UpdateCollaborators(collaborators Collaborators) error {
	for login, c := range collaborators {
		k.Collaborators[login] = c
	}

	if err := k.Save(); err != nil {
		return fmt.Errorf("failed to save keyring after updating collaborators: %w", err)
	}
	return nil
}

// GetCollaborator returns a collaborator by login
func (k *Keyring) GetCollaborator(login string) (Collaborator, bool) {
	collaborator, exists := k.Collaborators[login]
	if !exists {
		return Collaborator{}, false
	}
	return *collaborator, true
}

// GetCollaboratorByKey returns a collaborator by their SSH key
func (k *Keyring) GetCollaboratorByKey(key string) (string, Collaborator, bool) {
	for login, c := range k.Collaborators {
		for _, k := range c.SSHKeys {
			if k == key {
				return login, *c, true
			}
		}
	}
	return "", Collaborator{}, false
}
