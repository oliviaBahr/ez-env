package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AddFile adds a file to the list of files that should be encrypted
func AddFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no file specified")
	}

	filePath := args[0]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Add the file pattern to .gitattributes
	if err := addToGitAttributes(filePath); err != nil {
		return fmt.Errorf("failed to add file to .gitattributes: %w", err)
	}

	fmt.Printf("✓ File added for encryption: %s\n", filePath)
	fmt.Printf("Note: The file will be encrypted on next git add/commit\n")

	return nil
}

// addToGitAttributes adds a file pattern to .gitattributes
func addToGitAttributes(filePath string) error {
	// Read existing .gitattributes
	content, err := os.ReadFile(".gitattributes")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read .gitattributes: %w", err)
	}

	// Check if the pattern already exists
	pattern := filePath + " filter=ezenv\n"
	if os.IsNotExist(err) {
		// Create new .gitattributes
		content = []byte("# ezenv encrypted files\n" + pattern)
	} else {
		// Check if pattern already exists
		if !containsPattern(string(content), filePath) {
			// Append to existing content
			content = append(content, []byte(pattern)...)
		}
	}

	// Write .gitattributes
	if err := os.WriteFile(".gitattributes", content, 0644); err != nil {
		return fmt.Errorf("failed to write .gitattributes: %w", err)
	}

	// Add .gitattributes to git
	addCmd := exec.Command("git", "add", ".gitattributes")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add .gitattributes to git: %w", err)
	}

	return nil
}

// containsPattern checks if a file pattern already exists in .gitattributes
func containsPattern(content, filePath string) bool {
	pattern := filePath + " filter=ezenv"
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == pattern {
			return true
		}
	}
	return false
}
