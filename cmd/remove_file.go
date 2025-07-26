package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RemoveFile removes a file from the list of files that should be encrypted
func RemoveFile(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no file specified")
	}

	filePath := args[0]

	// Remove the file pattern from .gitattributes
	if err := removeFromGitAttributes(filePath); err != nil {
		return fmt.Errorf("failed to remove file from .gitattributes: %w", err)
	}

	fmt.Printf("✓ File removed from encryption: %s\n", filePath)
	fmt.Printf("Note: The file will no longer be encrypted on git add/commit\n")

	return nil
}

// removeFromGitAttributes removes a file pattern from .gitattributes
func removeFromGitAttributes(filePath string) error {
	// Read existing .gitattributes
	content, err := os.ReadFile(".gitattributes")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(".gitattributes file does not exist")
		}
		return fmt.Errorf("failed to read .gitattributes: %w", err)
	}

	// Check if the pattern exists
	if !containsPattern(string(content), filePath) {
		return fmt.Errorf("file pattern not found in .gitattributes: %s", filePath)
	}

	// Remove the pattern
	lines := strings.Split(string(content), "\n")
	var newLines []string
	pattern := filePath + " filter=ezenv"

	for _, line := range lines {
		if strings.TrimSpace(line) != pattern {
			newLines = append(newLines, line)
		}
	}

	// If no lines left (except empty ones), remove the file
	if len(newLines) == 0 || (len(newLines) == 1 && strings.TrimSpace(newLines[0]) == "") {
		if err := os.Remove(".gitattributes"); err != nil {
			return fmt.Errorf("failed to remove .gitattributes: %w", err)
		}
	} else {
		// Write updated content
		newContent := strings.Join(newLines, "\n")
		if !strings.HasSuffix(newContent, "\n") {
			newContent += "\n"
		}
		if err := os.WriteFile(".gitattributes", []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write .gitattributes: %w", err)
		}
	}

	// Add .gitattributes to git (or remove if deleted)
	if _, err := os.Stat(".gitattributes"); err == nil {
		addCmd := exec.Command("git", "add", ".gitattributes")
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add .gitattributes to git: %w", err)
		}
	} else {
		rmCmd := exec.Command("git", "rm", ".gitattributes")
		if err := rmCmd.Run(); err != nil {
			return fmt.Errorf("failed to remove .gitattributes from git: %w", err)
		}
	}

	return nil
}
