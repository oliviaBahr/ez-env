package main

import (
	"fmt"
	"os"

	"github.com/olivia/ezenv/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git env <command>")
		fmt.Println("\nCommands:")
		fmt.Println("  init         Initialize ezenv in the current repository")
		fmt.Println("  add-file     Add a file to be encrypted")
		fmt.Println("  update-keys  Update collaborator keys")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "init":
		err = cmd.Init(args)
	case "add-file":
		err = cmd.AddFile(args)
	case "update-keys":
		err = cmd.UpdateKeys(args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
