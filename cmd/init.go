package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/oliviaBahr/ez-env/crypto"
	"github.com/oliviaBahr/ez-env/workflows"
)

// Init initializes ezenv in the current repository
func Init(args []string) error {
	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	ctx := context.Background()

	// Create key manager and get/create encryption key
	fmt.Println("Setting up ez-env with GitHub Actions workflow-based key management...")
	keyManager := crypto.NewKeyManager()
	key, err := keyManager.GetOrCreateEncryptionKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get or create encryption key: %w", err)
	}

	// Write the workflow file to the repository
	if err := writeWorkflowFile(); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	// Set up git attributes (will be populated as files are added)
	if err := setupGitAttributes(); err != nil {
		return fmt.Errorf("failed to set up git attributes: %w", err)
	}

	// Configure git filters
	if err := configureGitFilters(); err != nil {
		return fmt.Errorf("failed to configure git filters: %w", err)
	}

	// Add .gitattributes to git
	if err := addGitAttributesToGit(); err != nil {
		return fmt.Errorf("failed to add .gitattributes to git: %w", err)
	}

	// Add workflow file to git
	if err := addWorkflowToGit(); err != nil {
		return fmt.Errorf("failed to add workflow to git: %w", err)
	}

	fmt.Println("✓ ezenv initialized successfully!")
	fmt.Printf("✓ Encryption key: %d bytes\n", len(key))
	fmt.Println("✓ Git filters configured")
	fmt.Println("✓ .gitattributes created")
	fmt.Println("✓ GitHub workflow created")
	fmt.Println("\nKey Management:")
	fmt.Println("  - Encryption key stored in GitHub repository secrets")
	fmt.Println("  - Key distribution via GitHub Actions workflow")
	fmt.Println("  - Access controlled by repository permissions")
	fmt.Println("\nNext steps:")
	fmt.Println("  - Use 'git ez-env add <file>' to specify files for encryption")
	fmt.Println("  - Use 'git add <file>' to stage files (they'll be encrypted automatically)")
	fmt.Println("  - Push changes to enable workflow-based key management for collaborators")

	return nil
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

func writeWorkflowFile() error {
	fmt.Println("Setting up GitHub workflow...")

	// Get the current working directory (repository root)
	repoPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Write the workflow file
	if err := workflows.WriteWorkflowFile(repoPath); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	fmt.Println("✓ GitHub workflow created")
	return nil
}

func setupGitAttributes() error {
	content := `# ezenv encrypted files
# Files will be added here when you run 'git ez-env add <file>'
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

	// Configure clean filter to run on add/commit
	cleanCmd := exec.Command("git", "config", "filter.ezenv.clean", exe+" clean")
	if err := cleanCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure clean filter: %w", err)
	}

	// Configure smudge filter to run on checkout
	smudgeCmd := exec.Command("git", "config", "filter.ezenv.smudge", exe+" smudge")
	if err := smudgeCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure smudge filter: %w", err)
	}

	// Enable the filter to run automatically
	requiredCmd := exec.Command("git", "config", "filter.ezenv.required", "true")
	if err := requiredCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure filter as required: %w", err)
	}

	return nil
}

func addGitAttributesToGit() error {
	// Add .gitattributes
	addAttrsCmd := exec.Command("git", "add", ".gitattributes")
	if err := addAttrsCmd.Run(); err != nil {
		return fmt.Errorf("failed to add .gitattributes to git: %w", err)
	}

	return nil
}

func addWorkflowToGit() error {
	// Add the workflow file
	addWorkflowCmd := exec.Command("git", "add", ".github/workflows/ez-env-key-management.yml")
	if err := addWorkflowCmd.Run(); err != nil {
		return fmt.Errorf("failed to add workflow to git: %w", err)
	}

	return nil
}
