package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// SecretName is the name of the GitHub repository secret that stores the encryption key
	SecretName = "EZENV_ENCRYPTION_KEY"
	// WorkflowName is the name of the workflow for key management
	WorkflowName = "ez-env-key-management.yml"
)

// GetGitHubToken retrieves the GitHub token from environment or gh auth status
func GetGitHubToken() (string, error) {
	// First try environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Then try gh auth status
	cmd := exec.Command("gh", "auth", "status", "--show-token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GitHub token: %w", err)
	}

	// Parse the output to extract the token
	// Format: "  - Token: gho_..."
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Token: ") {
			// Extract token from line like "  - Token: gho_..."
			parts := strings.Split(line, "Token: ")
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("no GitHub token found")
}

// GetCurrentUser gets the current authenticated user
func GetCurrentUser(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "api", "user", "--jq", ".login")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	// Remove newline from output
	username := strings.TrimSpace(string(output))
	return username, nil
}

// GetRepositoryInfo gets the owner and repository name from the current git remote
func GetRepositoryInfo() (string, string, error) {
	// Get the current repository
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Parse the remote URL to get owner/repo
	// Format: git@github.com:owner/repo.git or https://github.com/owner/repo.git
	remoteURL := strings.TrimSpace(string(output))
	var owner, repo string
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid remote URL format: %s", remoteURL)
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else if strings.HasPrefix(remoteURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "https://github.com/"), "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid remote URL format: %s", remoteURL)
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else {
		return "", "", fmt.Errorf("unsupported remote URL format: %s", remoteURL)
	}

	return owner, repo, nil
}

// StoreEncryptionKey stores the encryption key as a GitHub repository secret using GitHub CLI
func StoreEncryptionKey(ctx context.Context, key []byte) error {
	// Convert key to base64
	keyB64 := base64.StdEncoding.EncodeToString(key)

	// Store the key using GitHub CLI
	cmd := exec.CommandContext(ctx, "gh", "secret", "set", SecretName, "--body", keyB64)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to store encryption key: %w", err)
	}

	return nil
}

// GetEncryptionKey retrieves the encryption key via GitHub workflow
func GetEncryptionKey(ctx context.Context) ([]byte, error) {
	currentUser, err := GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	fmt.Printf("Triggering GitHub workflow to retrieve encryption key...\n")

	// Trigger the workflow to get the key
	cmd := exec.CommandContext(ctx, "gh", "workflow", "run", WorkflowName,
		"--field", "action=get-key",
		"--field", fmt.Sprintf("user=%s", currentUser))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to trigger workflow: %w", err)
	}

	// Wait a moment for the workflow to start
	time.Sleep(2 * time.Second)

	// Get the latest workflow run for this workflow
	cmd = exec.CommandContext(ctx, "gh", "run", "list", "--workflow", WorkflowName, "--limit", "1", "--json", "databaseId,status")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow run: %w", err)
	}

	var runs []struct {
		DatabaseID int    `json:"databaseId"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(output, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse workflow runs: %w", err)
	}

	if len(runs) == 0 {
		return nil, fmt.Errorf("no workflow runs found")
	}

	runID := runs[0].DatabaseID
	fmt.Printf("Waiting for workflow run %d to complete...\n", runID)

	// Wait for the workflow to complete
	for i := 0; i < 60; i++ { // Wait up to 60 seconds
		cmd = exec.CommandContext(ctx, "gh", "run", "view", fmt.Sprintf("%d", runID), "--json", "status,conclusion")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to check workflow status: %w", err)
		}

		var run struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		}
		if err := json.Unmarshal(output, &run); err != nil {
			return nil, fmt.Errorf("failed to parse workflow status: %w", err)
		}

		if run.Status == "completed" {
			if run.Conclusion == "success" {
				fmt.Printf("✓ Workflow completed successfully\n")
				break
			} else if run.Conclusion == "failure" {
				return nil, fmt.Errorf("workflow failed with conclusion: %s", run.Conclusion)
			} else if run.Conclusion == "cancelled" {
				return nil, fmt.Errorf("workflow was cancelled")
			} else {
				return nil, fmt.Errorf("workflow completed with unexpected conclusion: %s", run.Conclusion)
			}
		} else if run.Status == "failed" {
			return nil, fmt.Errorf("workflow failed")
		} else if run.Status == "cancelled" {
			return nil, fmt.Errorf("workflow was cancelled")
		}

		// Show progress for longer waits
		if i > 0 && i%10 == 0 {
			fmt.Printf("Still waiting for workflow completion... (attempt %d/60)\n", i+1)
		}

		time.Sleep(1 * time.Second)
	}

	// Wait for artifacts to be available
	artifactName := fmt.Sprintf("encryption-key-%s", currentUser)
	fmt.Printf("Waiting for encryption key artifact to be available...\n")

	if err := waitForArtifact(ctx, runID, artifactName); err != nil {
		return nil, fmt.Errorf("failed to wait for artifact: %w", err)
	}

	// Download the artifact
	fmt.Printf("Downloading encryption key artifact...\n")

	// Remove existing file if it exists to avoid download conflicts
	os.Remove("encryption-key.txt")

	cmd = exec.CommandContext(ctx, "gh", "run", "download", fmt.Sprintf("%d", runID), "--name", artifactName)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to download artifact: %w", err)
	}

	// Read the key from the downloaded file
	keyData, err := os.ReadFile("encryption-key.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	// Clean up the downloaded file
	os.Remove("encryption-key.txt")

	// Decode the base64 key
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(keyData)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	fmt.Printf("✓ Encryption key retrieved successfully\n")
	return key, nil
}

// waitForArtifact polls the GitHub API to check if the specified artifact is available
func waitForArtifact(ctx context.Context, runID int, artifactName string) error {
	// Get repository info for API calls
	owner, repo, err := GetRepositoryInfo()
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	// Poll for artifact availability
	for i := 0; i < 30; i++ { // Wait up to 30 seconds for artifacts
		// Check if artifact exists using GitHub API
		cmd := exec.CommandContext(ctx, "gh", "api",
			fmt.Sprintf("repos/%s/%s/actions/runs/%d/artifacts", owner, repo, runID),
			"--jq", fmt.Sprintf(`.[] | select(.name == "%s") | .id`, artifactName))

		output, err := cmd.Output()
		if err == nil && len(strings.TrimSpace(string(output))) > 0 {
			// Artifact found
			return nil
		}

		// Wait before next check
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("artifact %s not available after 30 seconds", artifactName)
}
