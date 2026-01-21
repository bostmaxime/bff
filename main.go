package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var validCommands = []string{"index", "compare", "duplicates", "find"}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	isValidCommand := false
	for _, validCommand := range validCommands {
		if command == validCommand {
			isValidCommand = true
			break
		}
	}
	if !isValidCommand {
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}

	rootPath := "."
	includeHidden := false
	targetFile := ""

	argIndex := 2
	if command == "find" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: 'find' command requires a file path\n")
			fmt.Fprintf(os.Stderr, "Usage: ./bff find <file-path> [directory]\n")
			os.Exit(1)
		}
		targetFile = os.Args[argIndex]
		argIndex = 3
	}

	for i := argIndex; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--hidden" || arg == "-h" {
			// --hidden flag is only allowed for index command.
			if command != "index" {
				fmt.Fprintf(os.Stderr, "Error: --hidden flag is only allowed with 'index' command\n")
				os.Exit(1)
			}
			includeHidden = true
		} else if rootPath == "." {
			rootPath = arg
		}
	}

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid path: %v\n", err)
		os.Exit(1)
	}

	index := NewIndex(absPath, includeHidden)

	if command == "index" {
		count, err := index.Index()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Indexed %d files\n", count)
		return
	}

	if err := index.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please run 'bff index' first to create an index\n")
		os.Exit(1)
	}

	switch command {
	case "compare":
		result, err := index.Compare()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		result.Print()

	case "duplicates":
		duplicates := index.FindAllDuplicates()
		if len(duplicates) == 0 {
			fmt.Println("No duplicates found")
			return
		}

		fmt.Printf("Found %d group(s) of duplicate files:\n\n", len(duplicates))
		for hash, files := range duplicates {
			fmt.Printf("Hash: %s\n", hash)
			fmt.Printf("  %d files with identical content:\n", len(files))
			for _, file := range files {
				fmt.Printf("    - %s\n", file.Path)
			}
			fmt.Println()
		}

	case "find":
		matches, err := index.FindDuplicates(targetFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(matches) == 1 {
			fmt.Printf("File '%s' has no duplicates\n", targetFile)
		} else {
			fmt.Printf("Found %d file(s) with identical content to '%s':\n", len(matches)-1, targetFile)
			for _, match := range matches {
				if match != targetFile {
					fmt.Printf("  - %s\n", match)
				}
			}
		}
	}
}

func printUsage() {
	fmt.Println("Usage: ./bff <command> [option] [directory]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  index                - Index all files including in subdirectories (creates/updates the index file)")
	fmt.Println("                         Option: --hidden, -h to include hidden files and directories")
	fmt.Println("  compare              - Compare current state with last saved index")
	fmt.Println("  duplicates           - Find all duplicate files")
	fmt.Println("  find <path>          - Find all duplicates of a specific file")
	fmt.Println()
	fmt.Println("Directory:")
	fmt.Println("  Optional path to the directory (default: current directory)")
	fmt.Println()
	fmt.Println("Note: compare, duplicates and find commands require running index first")
	fmt.Println("Note: the hidden option is only applicable to the index command, then when using other commands the hidden settings from the saved index will be used")
}
