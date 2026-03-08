package account

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage handles account file operations
type Storage struct {
	accountsDir string
}

// NewStorage creates a new Storage instance
func NewStorage(accountsDir string) *Storage {
	return &Storage{
		accountsDir: accountsDir,
	}
}

// LoadAll loads all account files from the accounts directory
func (s *Storage) LoadAll() ([]Account, error) {
	if err := os.MkdirAll(s.accountsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create accounts directory: %w", err)
	}

	entries, err := os.ReadDir(s.accountsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts directory: %w", err)
	}

	var accounts []Account
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(s.accountsDir, entry.Name())
		acc, err := s.LoadByPath(filePath)
		if err != nil {
			// Log error but continue loading other accounts
			continue
		}
		accounts = append(accounts, *acc)
	}

	return accounts, nil
}

// LoadByPath loads an account from a specific file path
func (s *Storage) LoadByPath(filePath string) (*Account, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read account file: %w", err)
	}

	var acc Account
	if err := json.Unmarshal(data, &acc); err != nil {
		return nil, fmt.Errorf("failed to parse account JSON: %w", err)
	}

	// Get file modification time
	info, err := os.Stat(filePath)
	if err == nil {
		acc.ModTime = info.ModTime()
	}

	acc.FilePath = filePath
	return &acc, nil
}

// LoadByID loads an account by its account ID
func (s *Storage) LoadByID(accountID string) (*Account, error) {
	accounts, err := s.LoadAll()
	if err != nil {
		return nil, err
	}

	for _, acc := range accounts {
		if acc.AccountID == accountID {
			return &acc, nil
		}
	}

	return nil, fmt.Errorf("account not found: %s", accountID)
}

// Save saves an account to a file
func (s *Storage) Save(acc *Account) error {
	if err := os.MkdirAll(s.accountsDir, 0755); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	// Generate filename if not set
	if acc.FilePath == "" {
		filename := s.generateFilename(acc)
		acc.FilePath = filepath.Join(s.accountsDir, filename)
	}

	data, err := json.MarshalIndent(acc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account: %w", err)
	}

	if err := os.WriteFile(acc.FilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write account file: %w", err)
	}

	return nil
}

// Delete deletes an account file
func (s *Storage) Delete(acc *Account) error {
	if acc.FilePath == "" {
		return fmt.Errorf("account file path is empty")
	}

	if err := os.Remove(acc.FilePath); err != nil {
		return fmt.Errorf("failed to delete account file: %w", err)
	}

	return nil
}

// Backup backs up an account file to a backup directory
func (s *Storage) Backup(acc *Account, backupDir string) error {
	if acc.FilePath == "" {
		return fmt.Errorf("account file path is empty")
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	filename := filepath.Base(acc.FilePath)
	backupPath := filepath.Join(backupDir, filename)

	data, err := os.ReadFile(acc.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read account file: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// generateFilename generates a filename for an account
func (s *Storage) generateFilename(acc *Account) string {
	if acc.Email != "" {
		hash := md5.Sum([]byte(acc.Email))
		return fmt.Sprintf("%x.json", hash)
	}
	if acc.AccountID != "" {
		return fmt.Sprintf("%s.json", acc.AccountID)
	}
	return fmt.Sprintf("account_%d.json", time.Now().Unix())
}
