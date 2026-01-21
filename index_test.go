package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndex(t *testing.T) {
	hashContent := computeHash([]byte("content"))
	hashSecret := computeHash([]byte("secret"))

	tests := []struct {
		name          string
		includeHidden bool
		setupFunc     func(string) error
		expectedCount int
		expectedMap   map[string][]string
	}{
		{
			name:          "empty_directory",
			includeHidden: false,
			setupFunc: func(dir string) error {
				return nil
			},
			expectedCount: 0,
			expectedMap:   map[string][]string{},
		},
		{
			name:          "one_file",
			includeHidden: false,
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
			},
			expectedCount: 1,
			expectedMap: map[string][]string{
				hashContent: {"file.txt"},
			},
		},
		{
			name:          "one_file_and_one_subdir",
			includeHidden: false,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				subdir := filepath.Join(dir, "subdir")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("content"), 0644)
			},
			expectedCount: 2,
			expectedMap: map[string][]string{
				hashContent: {"file.txt", "subdir/nested.txt"},
			},
		},
		{
			name:          "two_subdirs",
			includeHidden: false,
			setupFunc: func(dir string) error {
				subdir1 := filepath.Join(dir, "subdir1")
				subdir2 := filepath.Join(dir, "subdir2")
				if err := os.Mkdir(subdir1, 0755); err != nil {
					return err
				}
				if err := os.Mkdir(subdir2, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(subdir1, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(subdir2, "file.txt"), []byte("content"), 0644)
			},
			expectedCount: 2,
			expectedMap: map[string][]string{
				hashContent: {"subdir1/file.txt", "subdir2/file.txt"},
			},
		},
		{
			name:          "hidden_file_excluded",
			includeHidden: false,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("content"), 0644)
			},
			expectedCount: 1,
			expectedMap: map[string][]string{
				hashContent: {"file.txt"},
			},
		},
		{
			name:          "hidden_file_included",
			includeHidden: true,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden"), []byte("content"), 0644)
			},
			expectedCount: 2,
			expectedMap: map[string][]string{
				hashContent: {"file.txt", ".hidden"},
			},
		},
		{
			name:          "hidden_directory_excluded",
			includeHidden: false,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				hiddenDir := filepath.Join(dir, ".hidden")
				if err := os.Mkdir(hiddenDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(hiddenDir, "secret.txt"), []byte("secret"), 0644)
			},
			expectedCount: 1,
			expectedMap: map[string][]string{
				hashContent: {"file.txt"},
			},
		},
		{
			name:          "hidden_directory_included",
			includeHidden: true,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				hiddenDir := filepath.Join(dir, ".hidden")
				if err := os.Mkdir(hiddenDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(hiddenDir, "secret.txt"), []byte("secret"), 0644)
			},
			expectedCount: 2,
			expectedMap: map[string][]string{
				hashContent: {"file.txt"},
				hashSecret:  {".hidden/secret.txt"},
			},
		},
		{
			name:          "duplicate_content",
			includeHidden: false,
			setupFunc: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content"), 0644)
			},
			expectedCount: 2,
			expectedMap: map[string][]string{
				hashContent: {"file1.txt", "file2.txt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()

			if err := tt.setupFunc(testDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			idx := NewIndex(testDir, tt.includeHidden)
			count, err := idx.Index()
			if err != nil {
				t.Fatalf("Index() failed: %v", err)
			}

			if count != tt.expectedCount {
				t.Errorf("expected %d files indexed, got %d", tt.expectedCount, count)
			}

			indexPath := filepath.Join(testDir, IndexFile)
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				t.Error("index file was not created")
			}

			if len(tt.expectedMap) != len(idx.FilesByContentHash) {
				t.Errorf("expected %d different hashes, got %d", len(tt.expectedMap), len(idx.FilesByContentHash))
			}

			for expectedHash, expectedFiles := range tt.expectedMap {
				actualFiles, exists := idx.FilesByContentHash[expectedHash]
				if !exists {
					t.Errorf("expected hash %q not found in index", expectedHash)
					continue
				}

				if len(expectedFiles) != len(actualFiles) {
					t.Errorf("for hash %q: expected %d files, got %d", expectedHash, len(expectedFiles), len(actualFiles))
					continue
				}

				actualPaths := make(map[string]bool)
				for _, file := range actualFiles {
					actualPaths[file.Path] = true
				}

				for _, expectedPath := range expectedFiles {
					if !actualPaths[expectedPath] {
						t.Errorf("for hash %q: expected file %q not found", expectedHash, expectedPath)
					}
				}
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name                   string
		initialSetup           func(string) error
		changeSetup            func(string) error
		expectedAdded          int
		expectedModified       int
		expectedDeleted        int
		expectedRenamedOrMoved int
		checkFiles             func(*testing.T, *Comparison)
	}{
		{
			name: "no_changes",
			initialSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
			},
			changeSetup: func(dir string) error {
				return nil
			},
			expectedAdded:          0,
			expectedModified:       0,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 0,
		},
		{
			name: "file_added",
			initialSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644)
			},
			changeSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
			},
			expectedAdded:          1,
			expectedModified:       0,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 0,
			checkFiles: func(t *testing.T, c *Comparison) {
				if len(c.Added) > 0 && c.Added[0] != "file2.txt" {
					t.Errorf("expected 'file2.txt' to be added, got %v", c.Added)
				}
			},
		},
		{
			name: "file_deleted",
			initialSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644)
			},
			changeSetup: func(dir string) error {
				return os.Remove(filepath.Join(dir, "file2.txt"))
			},
			expectedAdded:          0,
			expectedModified:       0,
			expectedDeleted:        1,
			expectedRenamedOrMoved: 0,
			checkFiles: func(t *testing.T, c *Comparison) {
				if len(c.Deleted) > 0 && c.Deleted[0] != "file2.txt" {
					t.Errorf("expected 'file2.txt' to be deleted, got %v", c.Deleted)
				}
			},
		},
		{
			name: "file_modified",
			initialSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("original content"), 0644)
			},
			changeSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("modified content"), 0644)
			},
			expectedAdded:          0,
			expectedModified:       1,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 0,
			checkFiles: func(t *testing.T, c *Comparison) {
				if len(c.Modified) > 0 && c.Modified[0] != "file.txt" {
					t.Errorf("expected 'file.txt' to be modified, got %v", c.Modified)
				}
			},
		},
		{
			name: "file_renamed",
			initialSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "old.txt"), []byte("content"), 0644)
			},
			changeSetup: func(dir string) error {
				return os.Rename(filepath.Join(dir, "old.txt"), filepath.Join(dir, "new.txt"))
			},
			expectedAdded:          0,
			expectedModified:       0,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 1,
			checkFiles: func(t *testing.T, c *Comparison) {
				if len(c.RenamedOrMoved) > 0 {
					if c.RenamedOrMoved[0].OldPath != "old.txt" || c.RenamedOrMoved[0].NewPath != "new.txt" {
						t.Errorf("expected 'old.txt' -> 'new.txt', got %v -> %v", c.RenamedOrMoved[0].OldPath, c.RenamedOrMoved[0].NewPath)
					}
				}
			},
		},
		{
			name: "file_moved_to_subdir",
			initialSetup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
			},
			changeSetup: func(dir string) error {
				subdir := filepath.Join(dir, "subdir")
				if err := os.Mkdir(subdir, 0755); err != nil {
					return err
				}
				return os.Rename(filepath.Join(dir, "file.txt"), filepath.Join(subdir, "file.txt"))
			},
			expectedAdded:          0,
			expectedModified:       0,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 1,
			checkFiles: func(t *testing.T, c *Comparison) {
				if len(c.RenamedOrMoved) > 0 {
					if c.RenamedOrMoved[0].OldPath != "file.txt" || c.RenamedOrMoved[0].NewPath != "subdir/file.txt" {
						t.Errorf("expected 'file.txt' -> 'subdir/file.txt', got %v -> %v", c.RenamedOrMoved[0].OldPath, c.RenamedOrMoved[0].NewPath)
					}
				}
			},
		},
		{
			name: "multiple_changes",
			initialSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content1"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "file2.txt"), []byte("content2"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "file3.txt"), []byte("content3"), 0644)
			},
			changeSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "file4.txt"), []byte("content4"), 0644); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("modified content1"), 0644); err != nil {
					return err
				}
				if err := os.Remove(filepath.Join(dir, "file2.txt")); err != nil {
					return err
				}
				return os.Rename(filepath.Join(dir, "file3.txt"), filepath.Join(dir, "file3_renamed.txt"))
			},
			expectedAdded:          1, // file4
			expectedModified:       1, // file1
			expectedDeleted:        1, // file2
			expectedRenamedOrMoved: 1, // file3
		},
		{
			name: "saved_index_with_hidden_true",
			initialSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("visible"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("hidden"), 0644)
			},
			changeSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("visible modified"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("hidden modified"), 0644)
			},
			expectedAdded:          0,
			expectedModified:       2,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 0,
		},
		{
			name: "saved_index_with_hidden_false",
			initialSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("visible"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("hidden"), 0644)
			},
			changeSetup: func(dir string) error {
				if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("visible modified"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, ".hidden.txt"), []byte("hidden modified"), 0644)
			},
			expectedAdded:          0,
			expectedModified:       1,
			expectedDeleted:        0,
			expectedRenamedOrMoved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()

			if err := tt.initialSetup(testDir); err != nil {
				t.Fatalf("initial setup failed: %v", err)
			}

			includeHidden := false
			if tt.name == "saved_index_with_hidden_true" {
				includeHidden = true
			}

			idx := NewIndex(testDir, includeHidden)
			if _, err := idx.Index(); err != nil {
				t.Fatalf("Index() failed: %v", err)
			}

			if err := tt.changeSetup(testDir); err != nil {
				t.Fatalf("change setup failed: %v", err)
			}

			result, err := idx.Compare()
			if err != nil {
				t.Fatalf("Compare() failed: %v", err)
			}

			if len(result.Added) != tt.expectedAdded {
				t.Errorf("expected %d added files, got %d: %v", tt.expectedAdded, len(result.Added), result.Added)
			}
			if len(result.Modified) != tt.expectedModified {
				t.Errorf("expected %d modified files, got %d: %v", tt.expectedModified, len(result.Modified), result.Modified)
			}
			if len(result.Deleted) != tt.expectedDeleted {
				t.Errorf("expected %d deleted files, got %d: %v", tt.expectedDeleted, len(result.Deleted), result.Deleted)
			}
			if len(result.RenamedOrMoved) != tt.expectedRenamedOrMoved {
				t.Errorf("expected %d renamed/moved files, got %d: %v", tt.expectedRenamedOrMoved, len(result.RenamedOrMoved), result.RenamedOrMoved)
			}

			if tt.checkFiles != nil {
				tt.checkFiles(t, result)
			}
		})
	}
}

func TestIndexPath(t *testing.T) {
	idx := NewIndex("/tmp", false)
	expected := filepath.Join("/tmp", IndexFile)

	if idx.indexPath() != expected {
		t.Errorf("expected %s, got %s", expected, idx.indexPath())
	}
}

func TestFindAllDuplicates(t *testing.T) {
	testDir := t.TempDir()

	content1 := []byte("unique content 1")
	content2 := []byte("duplicate content")
	content3 := []byte("another duplicate content")

	hashContent2 := computeHash(content2)
	hashContent3 := computeHash(content3)

	if err := os.WriteFile(filepath.Join(testDir, "file1.txt"), content1, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file2.txt"), content2, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file3.txt"), content2, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file4.txt"), content2, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file5.txt"), content3, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file6.txt"), content3, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	idx := NewIndex(testDir, false)
	if _, err := idx.Index(); err != nil {
		t.Fatalf("indexing failed: %v", err)
	}

	duplicates := idx.FindAllDuplicates()

	if len(duplicates) != 2 {
		t.Errorf("expected 2 duplicate groups, got %d", len(duplicates))
	}

	group2, exists := duplicates[hashContent2]
	if !exists {
		t.Errorf("expected duplicate group for content2 hash %s not found", hashContent2)
	} else if len(group2) != 3 {
		t.Errorf("expected 3 files in content2 duplicate group, got %d", len(group2))
	} else {
		expectedPaths := map[string]bool{"file2.txt": true, "file3.txt": true, "file4.txt": true}
		for _, file := range group2 {
			if !expectedPaths[file.Path] {
				t.Errorf("unexpected file %s in content2 duplicate group", file.Path)
			}
		}
	}

	group3, exists := duplicates[hashContent3]
	if !exists {
		t.Errorf("expected duplicate group for content3 hash %s not found", hashContent3)
	} else if len(group3) != 2 {
		t.Errorf("expected 2 files in content3 duplicate group, got %d", len(group3))
	} else {
		expectedPaths := map[string]bool{"file5.txt": true, "file6.txt": true}
		for _, file := range group3 {
			if !expectedPaths[file.Path] {
				t.Errorf("unexpected file %s in content3 duplicate group", file.Path)
			}
		}
	}
}

func TestFindDuplicates(t *testing.T) {
	testDir := t.TempDir()

	duplicateContent := []byte("duplicate content")
	uniqueContent := []byte("unique content")

	if err := os.WriteFile(filepath.Join(testDir, "file1.txt"), duplicateContent, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file2.txt"), duplicateContent, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file3.txt"), uniqueContent, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	idx := NewIndex(testDir, false)
	if _, err := idx.Index(); err != nil {
		t.Fatalf("indexing failed: %v", err)
	}

	matches, err := idx.FindDuplicates("file1.txt")
	if err != nil {
		t.Fatalf("FindDuplicate failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("expected 2 matches (file1.txt and file2.txt), got %d: %v", len(matches), matches)
	}

	matches, err = idx.FindDuplicates("file3.txt")
	if err != nil {
		t.Fatalf("FindDuplicate failed: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("expected 1 match (only file3.txt), got %d: %v", len(matches), matches)
	}

	_, err = idx.FindDuplicates("nonexistent.txt")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
