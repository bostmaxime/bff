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
	FilesByContentHash map[string][]*FileInfo `json:"files_by_content_hash"`
	AbsPath            string                 `json:"abs_path"`
	IncludeHidden      bool                   `json:"include_hidden"` // Whether hidden files are included.
}

// NewIndex initializes a new empty index for the given root path.
func NewIndex(rootPath string, includeHidden bool) *Index {
	return &Index{
		FilesByContentHash: make(map[string][]*FileInfo),
		AbsPath:            rootPath,
		IncludeHidden:      includeHidden,
	}
}

// Index scans the directory and saves the index file as a JSON (creates it if it doesn't exist).
// It also returns the number of indexed files.
func (idx *Index) Index() (int, error) {
	indexedFilesCount, err := idx.scan()
	if err != nil {
		return 0, err
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(idx.indexPath(), data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write index: %w", err)
	}

	return indexedFilesCount, nil
}

// scan walks through the directory and indexes all files (including in subdirectories).
// It also returns the total number of files indexed.
func (idx *Index) scan() (int, error) {
	var indexedFilesCount int

	err := filepath.Walk(idx.AbsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		// Ignore the index file voluntarily.
		if path == idx.indexPath() {
			return nil
		}

		if !idx.IncludeHidden && path != idx.AbsPath && strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(idx.AbsPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		hash, fileInfo, err := ProcessFile(path, relPath)
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", path, err)
		}

		idx.FilesByContentHash[hash] = append(idx.FilesByContentHash[hash], fileInfo)

		indexedFilesCount++

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("scan failed: %w", err)
	}

	return indexedFilesCount, nil
}

// indexPath returns the full path to the index file.
func (idx *Index) indexPath() string {
	return filepath.Join(idx.AbsPath, IndexFile)
}

// Load loads an existing index from the JSON file into the current Index struct.
func (idx *Index) Load() error {
	indexPath := idx.indexPath()

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return fmt.Errorf("index not found at %s", indexPath)
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	if err := json.Unmarshal(data, idx); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	return nil
}

// Compare compares the loaded index with the current state of the directory.
// The index must be loaded before calling this method.
func (idx *Index) Compare() (*Comparison, error) {
	savedIndex := Index{
		FilesByContentHash: idx.FilesByContentHash,
	}

	idx.FilesByContentHash = make(map[string][]*FileInfo)
	if _, err := idx.scan(); err != nil {
		return nil, fmt.Errorf("failed to rescan current directory: %w", err)
	}

	result := &Comparison{
		Added:          []string{},
		Modified:       []string{},
		Deleted:        []string{},
		RenamedOrMoved: []RenamedOrMovedFile{},
	}

	savedHashByPath := make(map[string]string)
	for hash, files := range savedIndex.FilesByContentHash {
		for _, file := range files {
			savedHashByPath[file.Path] = hash
		}
	}

	currentHashByPath := make(map[string]string)
	for hash, files := range idx.FilesByContentHash {
		for _, file := range files {
			currentHashByPath[file.Path] = hash
		}
	}

	processedCurrent := make(map[string]bool)
	processedSaved := make(map[string]bool)

	// Check for modified files (same path, different hashes).
	for path, currentHash := range currentHashByPath {
		if savedHash, exists := savedHashByPath[path]; exists {
			if currentHash != savedHash {
				result.Modified = append(result.Modified, path)
			}
			processedCurrent[path] = true
			processedSaved[path] = true
		}
	}

	// Check for renamed or moved files (different paths, same hash).
	for currentPath, currentHash := range currentHashByPath {
		if processedCurrent[currentPath] {
			continue
		}
		if savedFiles, exists := savedIndex.FilesByContentHash[currentHash]; exists {
			for _, savedFile := range savedFiles {
				if processedSaved[savedFile.Path] {
					continue
				}

				result.RenamedOrMoved = append(result.RenamedOrMoved, RenamedOrMovedFile{
					OldPath: savedFile.Path,
					NewPath: currentPath,
				})
				processedCurrent[currentPath] = true
				processedSaved[savedFile.Path] = true
				break
			}
		}
	}

	// Remaining current files are added.
	for path := range currentHashByPath {
		if !processedCurrent[path] {
			result.Added = append(result.Added, path)
		}
	}

	// Remaining saved files are deleted.
	for path := range savedHashByPath {
		if !processedSaved[path] {
			result.Deleted = append(result.Deleted, path)
		}
	}

	return result, nil
}

// FindAllDuplicates returns a map of content hashes to lists of FileInfo for files that have duplicate content.
// The index must be loaded before calling this method.
func (idx *Index) FindAllDuplicates() map[string][]*FileInfo {
	duplicates := make(map[string][]*FileInfo)

	for hash, files := range idx.FilesByContentHash {
		if len(files) > 1 {
			duplicates[hash] = files
		}
	}

	return duplicates
}

// FindDuplicates searches for all files that have the same content hash as the one of the provided path.
// It includes the target file path itself in the results.
// The index must be loaded before calling this method.
func (idx *Index) FindDuplicates(targetPath string) ([]string, error) {
	var targetHash string

	for hash, files := range idx.FilesByContentHash {
		for _, file := range files {
			if file.Path == targetPath {
				targetHash = hash
				break
			}
		}
		if targetHash != "" {
			break
		}
	}

	if targetHash == "" {
		return nil, fmt.Errorf("file %q not found in index", targetPath)
	}

	matchingPaths := []string{}
	if files, exists := idx.FilesByContentHash[targetHash]; exists {
		for _, file := range files {
			matchingPaths = append(matchingPaths, file.Path)
		}
	}

	return matchingPaths, nil
}
