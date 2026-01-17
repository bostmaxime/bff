package main

import "fmt"

// Comparison contains the results of comparing two different indexes of a directory,
// at different times for example.
type Comparison struct {
	Added          []string
	Modified       []string
	Deleted        []string
	RenamedOrMoved []RenamedOrMovedFile
}

type RenamedOrMovedFile struct {
	OldPath string
	NewPath string
}

// hasChanges returns true if there are any changes.
func (c *Comparison) hasChanges() bool {
	return len(c.Added) > 0 || len(c.Modified) > 0 || len(c.Deleted) > 0 || len(c.RenamedOrMoved) > 0
}

// Print outputs the comparison in a readable format.
func (c *Comparison) Print() {
	if !c.hasChanges() {
		fmt.Println("No changes detected")
		return
	}

	if len(c.Added) > 0 {
		fmt.Println("\nAdded:")
		for _, path := range c.Added {
			fmt.Println("  +", path)
		}
	}

	if len(c.Modified) > 0 {
		fmt.Println("\nModified:")
		for _, path := range c.Modified {
			fmt.Println("  ~", path)
		}
	}

	if len(c.RenamedOrMoved) > 0 {
		fmt.Println("\nRenamed/Moved:")
		for _, file := range c.RenamedOrMoved {
			fmt.Printf("  → %s -> %s\n", file.OldPath, file.NewPath)
		}
	}

	if len(c.Deleted) > 0 {
		fmt.Println("\nDeleted:")
		for _, path := range c.Deleted {
			fmt.Println("  -", path)
		}
	}

	fmt.Printf("\n%d added, %d modified, %d renamed/moved, %d deleted\n",
		len(c.Added), len(c.Modified), len(c.RenamedOrMoved), len(c.Deleted))
}
