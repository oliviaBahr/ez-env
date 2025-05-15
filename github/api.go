package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Collaborator represents a GitHub user with their SSH keys
type Collaborator struct {
	Login   string
	SSHKeys []string
}

// GetCollaborators fetches all collaborators for a repository and their SSH keys
func GetCollaborators(owner, repo, token string) ([]Collaborator, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get collaborators
	opts := &github.ListCollaboratorsOptions{
		Affiliation: "all",
	}
	collaborators, _, err := client.Repositories.ListCollaborators(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list collaborators: %w", err)
	}

	var result []Collaborator
	for _, collab := range collaborators {
		// Get user's SSH keys
		keys, _, err := client.Users.ListKeys(ctx, *collab.Login, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH keys for %s: %w", *collab.Login, err)
		}

		// Convert keys to strings
		var sshKeys []string
		for _, key := range keys {
			sshKeys = append(sshKeys, *key.Key)
		}

		result = append(result, Collaborator{
			Login:   *collab.Login,
			SSHKeys: sshKeys,
		})
	}

	return result, nil
}
