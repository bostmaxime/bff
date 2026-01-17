package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"
)

// Content represents data associated with a single file.
type Content struct {
	Hash    string    `json:"hash"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// NewContent creates content data for a file.
func NewContent(path string) (*Content, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	hash, err := hashFile(path)
	if err != nil {
		return nil, err
	}

	return &Content{
		Hash:    hash,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}, nil
}

// hashFile computes the SHA-256 hash of a file.
func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
