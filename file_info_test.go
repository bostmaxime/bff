package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessFile(t *testing.T) {
	relPath := "hello.txt"

	testDir := t.TempDir()
	absPath := filepath.Join(testDir, relPath)
	testContent := []byte("Hello, World!")

	if err := os.WriteFile(absPath, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	hash, fileInfo, err := ProcessFile(absPath, relPath)
	if err != nil {
		t.Fatalf("ProcessFile() failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	info, _ := os.Stat(absPath)
	expectedFileInfo := &FileInfo{
		Path:    relPath,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	if !reflect.DeepEqual(*fileInfo, *expectedFileInfo) {
		t.Errorf("FileInfo not equal: got %v, want %v", *fileInfo, *expectedFileInfo)
	}

	hash2, _, err := ProcessFile(absPath, relPath)
	if err != nil {
		t.Fatalf("ProcessFile() second call failed: %v", err)
	}

	if hash != hash2 {
		t.Errorf("expected same hash for same file, got %s and %s", hash, hash2)
	}
}
