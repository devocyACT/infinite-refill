package topup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// DownloadAccounts downloads new accounts from the topup response
func DownloadAccounts(resp *TopupResponse, accountsDir string, httpClient *http.Client) ([]account.Account, error) {
	var newAccounts []account.Account

	for _, accInfo := range resp.Accounts {
		logger.Info("处理账号：%s", accInfo.FileName)

		var acc account.Account

		// Try to use auth_json first
		if len(accInfo.AuthJSON) > 0 {
			if err := json.Unmarshal(accInfo.AuthJSON, &acc); err != nil {
				logger.Warn("解析 %s 的 auth_json 失败：%v", accInfo.FileName, err)
				continue
			}
		} else if accInfo.DownloadURL != "" {
			// Download from URL
			var err error
			acc, err = downloadFromURL(accInfo.DownloadURL, httpClient)
			if err != nil {
				logger.Warn("从 %s 下载账号失败：%v", accInfo.DownloadURL, err)
				continue
			}
		} else {
			logger.Warn("%s 没有 auth_json 或 download_url", accInfo.FileName)
			continue
		}

		// Set file path
		acc.FilePath = filepath.Join(accountsDir, accInfo.FileName)

		// Save account
		storage := account.NewStorage(accountsDir)
		if err := storage.Save(&acc); err != nil {
			logger.Warn("保存账号 %s 失败：%v", accInfo.FileName, err)
			continue
		}

		logger.Info("已保存新账号：%s", accInfo.FileName)
		newAccounts = append(newAccounts, acc)
	}

	return newAccounts, nil
}

func downloadFromURL(url string, httpClient *http.Client) (account.Account, error) {
	var acc account.Account

	resp, err := httpClient.Get(url)
	if err != nil {
		return acc, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return acc, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return acc, fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &acc); err != nil {
		return acc, fmt.Errorf("failed to parse account JSON: %w", err)
	}

	return acc, nil
}

// SaveAccountsToFile saves accounts to individual JSON files
func SaveAccountsToFile(accounts []account.Account, accountsDir string) error {
	if err := os.MkdirAll(accountsDir, 0755); err != nil {
		return fmt.Errorf("failed to create accounts directory: %w", err)
	}

	storage := account.NewStorage(accountsDir)

	for _, acc := range accounts {
		if err := storage.Save(&acc); err != nil {
			return fmt.Errorf("failed to save account: %w", err)
		}
	}

	return nil
}
