package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// FileInfo represents info associated to a file.
type FileInfo struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// ProcessFile processes a file by reading its content and returning its hash and FileInfo.
func ProcessFile(absPath string, relPath string) (hash string, fileInfo *FileInfo, err error) {
	info, err := os.Stat(absPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to stat file: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", nil, fmt.Errorf("failed to read file for hashing: %w", err)
	}

	hash = hex.EncodeToString(hasher.Sum(nil))

	fileInfo = &FileInfo{
		Path:    relPath,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	return hash, fileInfo, nil
}
