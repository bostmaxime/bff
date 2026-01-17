package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const IndexFile = "bff.json"

// Index represents a snapshot of all the files in a directory (including in subdirectories).
type Index struct {
	Files          map[string]*Content `json:"files"`
	AbsPath        string              `json:"abs_path"`
	IncludeHidden  bool                `json:"include_hidden"` // Whether hidden files are included.
	hiddenExplicit bool                `json:"-"`              // Whether --hidden was explicitly passed by the user.
}

// NewIndex initializes a new empty index for the given root path.
func NewIndex(rootPath string, includeHidden bool) *Index {
	return &Index{
		Files:          make(map[string]*Content),
		AbsPath:        rootPath,
		IncludeHidden:  includeHidden,
		hiddenExplicit: false,
	}
}

// SetHiddenExplicit marks whether --hidden was explicitly passed by the user.
func (idx *Index) SetHiddenExplicit(explicit bool) {
	idx.hiddenExplicit = explicit
}

// Index scans the directory and saves the index file as a JSON (creates it if it doesn't exist).
// It also returns the number of indexed files.
func (idx *Index) Index() (int, error) {
	if err := idx.scan(); err != nil {
		return 0, err
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(idx.indexPath(), data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write index: %w", err)
	}

	return len(idx.Files), nil
}

// Compare loads the saved index and returns differences with the current state of the directory.
func (idx *Index) Compare() (*Comparison, error) {
	indexPath := idx.indexPath()

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("index not found")
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read existing index: %w", err)
	}

	var savedIndex Index
	if err := json.Unmarshal(data, &savedIndex); err != nil {
		return nil, fmt.Errorf("failed to parse existing index: %w", err)
	}

	// Use saved setting unless user explicitly passed --hidden.
	if !idx.hiddenExplicit {
		idx.IncludeHidden = savedIndex.IncludeHidden
	}

	if err := idx.scan(); err != nil {
		return nil, err
	}

	result := &Comparison{
		Added:          []string{},
		Modified:       []string{},
		Deleted:        []string{},
		RenamedOrMoved: []RenamedOrMovedFile{},
	}

	savedHashes := make(map[string]string)
	for path, content := range savedIndex.Files {
		savedHashes[content.Hash] = path
	}

	currentHashes := make(map[string]string)
	for path, content := range idx.Files {
		currentHashes[content.Hash] = path
	}

	processedCurrent := make(map[string]bool)
	processedSaved := make(map[string]bool)

	// Check for modified files (same path, different hashes).
	for path, current := range idx.Files {
		if saved, exists := savedIndex.Files[path]; exists {
			if current.Hash != saved.Hash {
				result.Modified = append(result.Modified, path)
			}
			processedCurrent[path] = true
			processedSaved[path] = true
		}
	}

	// Check for renamed or moved files (different paths, same hash).
	for path, current := range idx.Files {
		if processedCurrent[path] {
			continue
		}
		if savedPath, exists := savedHashes[current.Hash]; exists && !processedSaved[savedPath] {
			result.RenamedOrMoved = append(result.RenamedOrMoved, RenamedOrMovedFile{
				OldPath: savedPath,
				NewPath: path,
			})
			processedCurrent[path] = true
			processedSaved[savedPath] = true
		}
	}

	// Remaining current files are added.
	for path := range idx.Files {
		if !processedCurrent[path] {
			result.Added = append(result.Added, path)
		}
	}

	// Remaining saved files are deleted.
	for path := range savedIndex.Files {
		if !processedSaved[path] {
			result.Deleted = append(result.Deleted, path)
		}
	}

	return result, nil
}

// indexPath returns the full path to the index file.
func (idx *Index) indexPath() string {
	return filepath.Join(idx.AbsPath, IndexFile)
}

// scan walks through the directory and indexes all files (including in subdirectories).
func (idx *Index) scan() error {
	err := filepath.Walk(idx.AbsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		// Ignore the index file voluntarily.
		if path == idx.indexPath() {
			return nil
		}

		if !idx.IncludeHidden && path != idx.AbsPath && isHidden(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		content, err := NewContent(path)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		relPath, err := filepath.Rel(idx.AbsPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		idx.Files[relPath] = content
		return nil
	})

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	return nil
}

// isHidden returns true if the name starts with a dot.
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}
