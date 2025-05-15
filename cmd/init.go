package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/olivia/ezenv/github"
	"github.com/olivia/ezenv/keyring"
)

// Init initializes ezenv in the current repository
func Init(args []string) error {
	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Get GitHub token from environment or gh auth
	token, err := getGitHubToken()
	if err != nil {
		return fmt.Errorf("failed to get GitHub token: %w", err)
	}

	// Get repository info
	repo, err := getRepoInfo()
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// Fetch collaborators and their SSH keys
	collaborators, err := github.GetCollaborators(repo.owner, repo.name, token)
	if err != nil {
		return fmt.Errorf("failed to get collaborators: %w", err)
	}

	// Create keyring
	if err := keyring.Create(collaborators); err != nil {
		return fmt.Errorf("failed to create keyring: %w", err)
	}

	// Set up git attributes
	if err := setupGitAttributes(); err != nil {
		return fmt.Errorf("failed to set up git attributes: %w", err)
	}

	// Configure git filters
	if err := configureGitFilters(); err != nil {
		return fmt.Errorf("failed to configure git filters: %w", err)
	}

	fmt.Println("ezenv initialized successfully!")
	return nil
}

type repoInfo struct {
	owner string
	name  string
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

func getGitHubToken() (string, error) {
	// First try environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Then try gh auth
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no GitHub token found in environment or gh auth")
	}

	return string(output), nil
}

func getRepoInfo() (repoInfo, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return repoInfo{}, fmt.Errorf("failed to get remote URL")
	}

	// Parse the remote URL to get owner and repo name
	// This is a simplified version - you might want to handle more URL formats
	url := string(output)
	// TODO: Implement proper URL parsing
	return repoInfo{
		owner: "owner", // Placeholder
		name:  "repo",  // Placeholder
	}, nil
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
