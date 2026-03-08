package account

// Manager provides high-level account management operations
type Manager struct {
	storage *Storage
}

// NewManager creates a new Manager instance
func NewManager(accountsDir string) *Manager {
	return &Manager{
		storage: NewStorage(accountsDir),
	}
}

// GetStorage returns the underlying storage
func (m *Manager) GetStorage() *Storage {
	return m.storage
}

// LoadAll loads all accounts
func (m *Manager) LoadAll() ([]Account, error) {
	return m.storage.LoadAll()
}

// LoadByID loads an account by ID
func (m *Manager) LoadByID(accountID string) (*Account, error) {
	return m.storage.LoadByID(accountID)
}

// Save saves an account
func (m *Manager) Save(acc *Account) error {
	return m.storage.Save(acc)
}

// Delete deletes an account
func (m *Manager) Delete(acc *Account) error {
	return m.storage.Delete(acc)
}

// Backup backs up an account
func (m *Manager) Backup(acc *Account, backupDir string) error {
	return m.storage.Backup(acc, backupDir)
}

// Count returns the total number of accounts
func (m *Manager) Count() (int, error) {
	accounts, err := m.LoadAll()
	if err != nil {
		return 0, err
	}
	return len(accounts), nil
}
