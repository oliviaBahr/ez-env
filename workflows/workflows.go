package workflows

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed ez-env-key-management.yml
var workflowFS embed.FS

// WriteWorkflowFile writes the embedded workflow file to the repository
func WriteWorkflowFile(repoPath string) error {
	// Create the .github/workflows directory
	workflowsDir := filepath.Join(repoPath, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory: %w", err)
	}

	// Read the embedded workflow file
	workflowContent, err := workflowFS.ReadFile("ez-env-key-management.yml")
	if err != nil {
		return fmt.Errorf("failed to read embedded workflow file: %w", err)
	}

	// Write the workflow file to the repository
	workflowPath := filepath.Join(workflowsDir, "ez-env-key-management.yml")
	if err := os.WriteFile(workflowPath, workflowContent, 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}
