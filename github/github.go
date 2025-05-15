package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/oliviaBahr/ez-env/crypto"
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
	// Format: "Token: ghp_..."
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Token: ") {
			return strings.TrimPrefix(line, "Token: "), nil
		}
	}

	return "", fmt.Errorf("no GitHub token found")
}

// GetRepositoryCollaborators fetches all collaborators for the current repository
func GetRepositoryCollaborators(ctx context.Context) (crypto.Collaborators, error) {
	// Get the current repository
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	// Parse the remote URL to get owner/repo
	// Format: git@github.com:owner/repo.git or https://github.com/owner/repo.git
	remoteURL := strings.TrimSpace(string(output))
	var owner, repo string
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "git@github.com:"), "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid remote URL format: %s", remoteURL)
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else if strings.HasPrefix(remoteURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(remoteURL, "https://github.com/"), "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid remote URL format: %s", remoteURL)
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else {
		return nil, fmt.Errorf("unsupported remote URL format: %s", remoteURL)
	}

	// Get GitHub token
	token, err := GetGitHubToken()
	if err != nil {
		return nil, err
	}

	// Create HTTP client with token
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators", owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Get collaborators
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collaborators: %s", resp.Status)
	}

	var collaborators = make(crypto.Collaborators)
	if err := json.NewDecoder(resp.Body).Decode(&collaborators); err != nil {
		return nil, fmt.Errorf("failed to decode collaborators: %w", err)
	}

	// Get SSH keys for each collaborator
	for login, collab := range collaborators {
		keys, err := GetUserSSHKeys(ctx, token, login)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH keys for %s: %w", login, err)
		}
		collab.SSHKeys = keys
		collaborators[login] = collab
	}

	return collaborators, nil
}

// GetUserSSHKeys fetches all SSH keys for a GitHub user
func GetUserSSHKeys(ctx context.Context, token, username string) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("https://api.github.com/users/%s/keys", username), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get SSH keys: %s", resp.Status)
	}

	var keys []struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("failed to decode SSH keys: %w", err)
	}

	sshKeys := make([]string, len(keys))
	for i, k := range keys {
		sshKeys[i] = k.Key
	}

	return sshKeys, nil
}

// GetCollaborators fetches collaborators and their SSH keys from GitHub
func GetCollaborators(owner, repo, token string) (crypto.Collaborators, error) {
	// Create HTTP client with auth
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/%s/collaborators", owner, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Get collaborators
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get collaborators: %s", resp.Status)
	}

	// Parse response
	var users []struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Get SSH keys for each collaborator
	collaborators := make(crypto.Collaborators)
	for _, user := range users {
		// Get user's SSH keys
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/users/%s/keys", user.Login), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "token "+token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH keys for %s: %w", user.Login, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get SSH keys for %s: %s", user.Login, resp.Status)
		}

		// Parse response
		var keys []struct {
			Key string `json:"key"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Add collaborator with their SSH keys
		sshKeys := make([]string, len(keys))
		for i, key := range keys {
			sshKeys[i] = key.Key
		}
		collaborators[user.Login] = &crypto.Collaborator{
			SSHKeys: sshKeys,
		}
	}

	return collaborators, nil
}
