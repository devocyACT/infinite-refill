package account

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAccount_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		modTime  time.Time
		days     int
		expected bool
	}{
		{
			name:     "not expired",
			modTime:  time.Now().AddDate(0, 0, -10),
			days:     30,
			expected: false,
		},
		{
			name:     "expired",
			modTime:  time.Now().AddDate(0, 0, -40),
			days:     30,
			expected: true,
		},
		{
			name:     "zero time",
			modTime:  time.Time{},
			days:     30,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{
				ModTime: tt.modTime,
			}
			if got := acc.IsExpired(tt.days); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStorage_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "refill-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewStorage(tmpDir)

	// Create test account
	acc := &Account{
		Type:        "codex",
		AccessToken: "test-token",
		AccountID:   "user-123",
		Email:       "test@example.com",
	}

	// Save account
	if err := storage.Save(acc); err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Verify file exists
	if acc.FilePath == "" {
		t.Fatal("FilePath not set after save")
	}
	if _, err := os.Stat(acc.FilePath); os.IsNotExist(err) {
		t.Fatal("Account file not created")
	}

	// Load account
	loaded, err := storage.LoadByPath(acc.FilePath)
	if err != nil {
		t.Fatalf("Failed to load account: %v", err)
	}

	// Verify data
	if loaded.Type != acc.Type {
		t.Errorf("Type = %v, want %v", loaded.Type, acc.Type)
	}
	if loaded.AccessToken != acc.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loaded.AccessToken, acc.AccessToken)
	}
	if loaded.AccountID != acc.AccountID {
		t.Errorf("AccountID = %v, want %v", loaded.AccountID, acc.AccountID)
	}
	if loaded.Email != acc.Email {
		t.Errorf("Email = %v, want %v", loaded.Email, acc.Email)
	}
}

func TestStorage_LoadAll(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "refill-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewStorage(tmpDir)

	// Create multiple accounts
	accounts := []Account{
		{Type: "codex", AccessToken: "token1", AccountID: "user-1", Email: "test1@example.com"},
		{Type: "codex", AccessToken: "token2", AccountID: "user-2", Email: "test2@example.com"},
		{Type: "codex", AccessToken: "token3", AccountID: "user-3", Email: "test3@example.com"},
	}

	for i := range accounts {
		if err := storage.Save(&accounts[i]); err != nil {
			t.Fatalf("Failed to save account %d: %v", i, err)
		}
	}

	// Load all accounts
	loaded, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("Failed to load accounts: %v", err)
	}

	if len(loaded) != len(accounts) {
		t.Errorf("LoadAll() returned %d accounts, want %d", len(loaded), len(accounts))
	}
}

func TestStorage_Delete(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "refill-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewStorage(tmpDir)

	// Create and save account
	acc := &Account{
		Type:        "codex",
		AccessToken: "test-token",
		AccountID:   "user-123",
		Email:       "test@example.com",
	}

	if err := storage.Save(acc); err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	filePath := acc.FilePath

	// Delete account
	if err := storage.Delete(acc); err != nil {
		t.Fatalf("Failed to delete account: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Account file still exists after delete")
	}
}

func TestStorage_Backup(t *testing.T) {
	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "refill-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backupDir := filepath.Join(tmpDir, "backup")

	storage := NewStorage(tmpDir)

	// Create and save account
	acc := &Account{
		Type:        "codex",
		AccessToken: "test-token",
		AccountID:   "user-123",
		Email:       "test@example.com",
	}

	if err := storage.Save(acc); err != nil {
		t.Fatalf("Failed to save account: %v", err)
	}

	// Backup account
	if err := storage.Backup(acc, backupDir); err != nil {
		t.Fatalf("Failed to backup account: %v", err)
	}

	// Verify backup file exists
	backupPath := filepath.Join(backupDir, filepath.Base(acc.FilePath))
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file not created")
	}
}
