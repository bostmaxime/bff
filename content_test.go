package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewContent(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "hello.txt")
	testContent := []byte("Hello, World!")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	content, err := NewContent(testFile)
	if err != nil {
		t.Fatalf("NewContent() failed: %v", err)
	}

	expectedHash, _ := hashFile(testFile)
	info, _ := os.Stat(testFile)
	expectedContent := &Content{
		Hash:    expectedHash,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	if !reflect.DeepEqual(*content, *expectedContent) {
		t.Errorf("not equal: got %v, want %v", *content, *expectedContent)
	}
}

func TestHashFile(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "hello.txt")
	testContent := []byte("Hello, World!")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash1, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("failed to hash file: %v", err)
	}

	if hash1 == "" {
		t.Error("expected non-empty hash")
	}

	hash2, err := hashFile(testFile)
	if err != nil {
		t.Fatalf("failed to hash file: %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("expected same hash for same content, got %s and %s", hash1, hash2)
	}
}
