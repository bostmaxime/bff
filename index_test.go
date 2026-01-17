package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		name          string
		includeHidden bool
		setupFunc     func(string) error
		expectedCount int
		expectedFiles []string
	}{
		{
			name:          "empty_directory",
			includeHidden: false,
			setupFunc: func(dir string) error {
				return nil
			},
			expectedCount: 0,
			expectedFiles: []string{},
		},
		{
			name:          "one_file",
			includeHidden: false,
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
			},
			expectedCount: 1,
			expectedFiles: []string{"file.txt"},
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
			expectedFiles: []string{"file.txt", "subdir/nested.txt"},
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
			expectedFiles: []string{"subdir1/file.txt", "subdir2/file.txt"},
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
			expectedFiles: []string{"file.txt"},
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
			expectedFiles: []string{"file.txt", ".hidden"},
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
			expectedFiles: []string{"file.txt"},
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
			expectedFiles: []string{"file.txt", ".hidden/secret.txt"},
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

			for _, expectedFile := range tt.expectedFiles {
				if _, exists := idx.Files[expectedFile]; !exists {
					t.Errorf("expected file %q to be in index", expectedFile)
				}
			}

			if len(idx.Files) != tt.expectedCount {
				t.Errorf("expected %d files in index, got %d", tt.expectedCount, len(idx.Files))
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()

			if err := tt.initialSetup(testDir); err != nil {
				t.Fatalf("initial setup failed: %v", err)
			}

			idx := NewIndex(testDir, false)
			if _, err := idx.Index(); err != nil {
				t.Fatalf("Index() failed: %v", err)
			}

			if err := tt.changeSetup(testDir); err != nil {
				t.Fatalf("change setup failed: %v", err)
			}

			idx2 := NewIndex(testDir, false)
			result, err := idx2.Compare()
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

func TestIsHidden(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"hidden", ".git", true},
		{"not_hidden", "normal.txt", false},
	}

	for _, tt := range tests {
		result := isHidden(tt.value)
		if result != tt.expected {
			t.Errorf("on %q expected %v, got %v", tt.value, tt.expected, result)
		}
	}
}
