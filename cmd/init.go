package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/oliviaBahr/ez-env/crypto"
	"github.com/oliviaBahr/ez-env/github"
)

// Init initializes ezenv in the current repository
func Init(args []string) error {
	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Create a new keyring
	keyring := crypto.NewKeyring()

	// Generate a new DEK
	if err := keyring.DEK.GenerateRawKey(); err != nil {
		return fmt.Errorf("failed to generate DEK: %w", err)
	}

	// Get collaborators and their SSH keys
	collaborators, err := github.GetRepositoryCollaborators(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get collaborators: %w", err)
	}

	// Update keyring with collaborators
	if err := keyring.UpdateCollaborators(collaborators); err != nil {
		return fmt.Errorf("failed to update keyring with collaborators: %w", err)
	}

	// Generate encrypted DEKs for each collaborator
	if err := keyring.GenerateEncryptedDEKs(); err != nil {
		return fmt.Errorf("failed to generate encrypted DEKs: %w", err)
	}

	// Save the keyring
	if err := keyring.Save(); err != nil {
		return fmt.Errorf("failed to save keyring: %w", err)
	}

	// Set up git attributes
	if err := setupGitAttributes(); err != nil {
		return fmt.Errorf("failed to set up git attributes: %w", err)
	}

	// Configure git filters
	if err := configureGitFilters(); err != nil {
		return fmt.Errorf("failed to configure git filters: %w", err)
	}

	// Add .gitenv_keyring and .gitattributes to git
	if err := addFilesToGit(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	fmt.Println("ezenv initialized successfully!")
	return nil
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

func setupGitAttributes() error {
	content := `# ezenv encrypted files
*.env filter=ezenv
`
	return os.WriteFile(".gitattributes", []byte(content), 0644)
}

func configureGitFilters() error {
	// Get the absolute path to the ezenv binary
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Configure clean filter
	cleanCmd := exec.Command("git", "config", "filter.ezenv.clean", exe+" clean")
	if err := cleanCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure clean filter: %w", err)
	}

	// Configure smudge filter
	smudgeCmd := exec.Command("git", "config", "filter.ezenv.smudge", exe+" smudge")
	if err := smudgeCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure smudge filter: %w", err)
	}

	return nil
}

func addFilesToGit() error {
	// Add .gitenv_keyring
	addKeyringCmd := exec.Command("git", "add", ".gitenv_keyring")
	if err := addKeyringCmd.Run(); err != nil {
		return fmt.Errorf("failed to add .gitenv_keyring to git: %w", err)
	}

	// Add .gitattributes
	addAttrsCmd := exec.Command("git", "add", ".gitattributes")
	if err := addAttrsCmd.Run(); err != nil {
		return fmt.Errorf("failed to add .gitattributes to git: %w", err)
	}

	return nil
}
