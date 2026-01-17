package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]

	rootPath := "."
	includeHidden := false
	hiddenExplicit := false

	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--hidden" || arg == "-h" {
			includeHidden = true
			hiddenExplicit = true
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
	index.SetHiddenExplicit(hiddenExplicit)

	switch command {
	case "index":
		count, err := index.Index()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Indexed %d files\n", count)

	case "compare":
		result, err := index.Compare()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		result.Print()

	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: ./bff <command> [options] [directory]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  index   - Index all files including in subdirectories (creates the index file if not created yet)")
	fmt.Println("  compare - Compare current state with last saved index")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --hidden, -h  Include hidden files and directories")
	fmt.Println("                (compare uses saved setting in the JSON file unless it is overridden)")
	fmt.Println()
	fmt.Println("Directory:")
	fmt.Println("  Optional path (default: current directory)")
}
