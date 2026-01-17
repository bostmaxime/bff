package main

import (
	"testing"
)

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name     string
		comp     *Comparison
		expected bool
	}{
		{
			"no_changes",
			&Comparison{[]string{}, []string{}, []string{}, []RenamedOrMovedFile{}},
			false,
		},
		{
			"added",
			&Comparison{[]string{"file.txt"}, []string{}, []string{}, []RenamedOrMovedFile{}},
			true,
		},
		{
			"modified",
			&Comparison{[]string{}, []string{"file.txt"}, []string{}, []RenamedOrMovedFile{}},
			true,
		},
		{
			"renamed_or_moved",
			&Comparison{[]string{}, []string{}, []string{}, []RenamedOrMovedFile{{OldPath: "old.txt", NewPath: "new.txt"}}},
			true,
		},
		{
			"deleted",
			&Comparison{[]string{}, []string{}, []string{"file.txt"}, []RenamedOrMovedFile{}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.comp.hasChanges() != tt.expected {
				t.Errorf("hasChanges() = %v, want %v", tt.comp.hasChanges(), tt.expected)
			}
		})
	}
}
