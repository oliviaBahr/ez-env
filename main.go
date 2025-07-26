package main

import (
	"fmt"
	"os"

	"github.com/oliviaBahr/ez-env/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git ez-env <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  init         Initialize ezenv in the current repository")
		fmt.Println("  add         Add a file to be encrypted")
		fmt.Println("  remove      Remove a file from encryption")
		fmt.Println("\nKey Management:")
		fmt.Println("  - Uses GitHub Actions workflows for secure key distribution")
		fmt.Println("  - Keys stored in GitHub repository secrets")
		fmt.Println("  - Automatic access control via GitHub permissions")
		fmt.Println("\nPrerequisites:")
		fmt.Println("  - GitHub CLI (gh) installed and authenticated")
		fmt.Println("  - Repository with GitHub Actions enabled")
		fmt.Println("  - Collaborator access to the repository")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "init":
		err = cmd.Init(args)
	case "add":
		err = cmd.AddFile(args)
	case "remove":
		err = cmd.RemoveFile(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("\nAvailable commands:")
		fmt.Println("  init         Initialize ezenv in the current repository")
		fmt.Println("  add         Add a file to be encrypted")
		fmt.Println("  remove      Remove a file from encryption")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
