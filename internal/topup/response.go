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

	logger.Debug("开始下载账号：总数=%d", len(resp.Accounts))

	for i, accInfo := range resp.Accounts {
		logger.Info("处理账号 %d/%d：%s", i+1, len(resp.Accounts), accInfo.FileName)
		logger.Debug("账号信息: download_url=%s, auth_json_len=%d",
			accInfo.DownloadURL, len(accInfo.AuthJSON))

		var acc account.Account

		// Try to use auth_json first
		if len(accInfo.AuthJSON) > 0 {
			logger.Debug("使用 auth_json 解析账号")
			if err := json.Unmarshal(accInfo.AuthJSON, &acc); err != nil {
				logger.Warn("解析 %s 的 auth_json 失败：%v", accInfo.FileName, err)
				logger.Debug("auth_json 内容: %s", string(accInfo.AuthJSON))
				continue
			}
			logger.Debug("auth_json 解析成功: account_id=%s, email=%s", acc.AccountID, acc.Email)
		} else if accInfo.DownloadURL != "" {
			// Download from URL
			logger.Debug("从 URL 下载账号: %s", accInfo.DownloadURL)
			var err error
			acc, err = downloadFromURL(accInfo.DownloadURL, httpClient)
			if err != nil {
				logger.Warn("从 %s 下载账号失败：%v", accInfo.DownloadURL, err)
				continue
			}
			logger.Debug("URL 下载成功: account_id=%s, email=%s", acc.AccountID, acc.Email)
		} else {
			logger.Warn("%s 没有 auth_json 或 download_url", accInfo.FileName)
			continue
		}

		// Set file path
		acc.FilePath = filepath.Join(accountsDir, accInfo.FileName)
		logger.Debug("设置文件路径: %s", acc.FilePath)

		// Save account
		storage := account.NewStorage(accountsDir)
		if err := storage.Save(&acc); err != nil {
			logger.Warn("保存账号 %s 失败：%v", accInfo.FileName, err)
			continue
		}

		logger.Info("已保存新账号：%s", accInfo.FileName)
		newAccounts = append(newAccounts, acc)
	}

	logger.Debug("下载完成：成功=%d, 失败=%d", len(newAccounts), len(resp.Accounts)-len(newAccounts))

	return newAccounts, nil
}

func downloadFromURL(url string, httpClient *http.Client) (account.Account, error) {
	var acc account.Account

	logger.Debug("开始从 URL 下载: %s", url)

	resp, err := httpClient.Get(url)
	if err != nil {
		return acc, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug("下载响应状态码: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return acc, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return acc, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug("下载响应体长度: %d bytes", len(body))

	if err := json.Unmarshal(body, &acc); err != nil {
		logger.Debug("解析失败的 JSON: %s", string(body))
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
