package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Setup test directory
	testDir := filepath.Join(os.TempDir(), "fastdb_test")
	err := os.MkdirAll(filepath.Join(testDir, "data"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	t.Run("successful DB initialization", func(t *testing.T) {
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		defer os.Chdir(originalWd) // Restore original working directory

		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}

		err = initDB()
		if err != nil {
			t.Errorf("Expected successful DB initialization, got error: %v", err)
		}

		// Check if DB is not nil
		if db == nil {
			t.Error("Expected db to be initialized, got nil")
		}

		if db != nil {
			db.Close()
			db = nil
		}
	})

	t.Run("already initialized", func(t *testing.T) {
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}

		err = initDB()
		if err != nil {
			t.Fatalf("Failed first initialization: %v", err)
		}

		err = initDB()
		if err != nil {
			t.Errorf("Expected nil error on second initialization, got: %v", err)
		}

		if db != nil {
			db.Close()
			db = nil
		}
	})
}
