package account

import (
	"time"
)

// Account represents a ChatGPT account
type Account struct {
	Type        string    `json:"type"`
	AccessToken string    `json:"access_token"`
	AccountID   string    `json:"account_id"`
	Email       string    `json:"email"`
	FilePath    string    `json:"-"`
	ModTime     time.Time `json:"-"`
}

// IsExpired checks if the account file is older than the specified days
func (a *Account) IsExpired(days int) bool {
	if a.ModTime.IsZero() {
		return false
	}
	expiryDate := time.Now().AddDate(0, 0, -days)
	return a.ModTime.Before(expiryDate)
}
