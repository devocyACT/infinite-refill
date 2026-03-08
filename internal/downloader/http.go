package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/topup"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// DownloadWithHTTP downloads accounts using concurrent HTTP requests
func DownloadWithHTTP(accounts []topup.AccountInfo, accountsDir string, httpClient *http.Client, maxConcurrent int) ([]account.Account, error) {
	if len(accounts) == 0 {
		return nil, nil
	}

	logger.Info("使用 HTTP 并发下载 %d 个账号（并发数=%d）", len(accounts), maxConcurrent)

	// Create accounts directory
	if err := os.MkdirAll(accountsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建账号目录失败：%w", err)
	}

	// Download files concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)
	resultChan := make(chan downloadResult, len(accounts))

	for _, accInfo := range accounts {
		wg.Add(1)
		go func(info topup.AccountInfo) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := downloadResult{
				fileName: info.FileName,
				authJSON: info.AuthJSON,
			}

			// Try auth_json first
			if len(info.AuthJSON) > 0 {
				result.success = true
			} else if info.DownloadURL != "" {
				// Download from URL
				acc, err := downloadAccountFromURL(info.DownloadURL, httpClient)
				if err != nil {
					logger.Debug("下载 %s 失败: %v", info.FileName, err)
					result.success = false
				} else {
					result.success = true
					result.account = acc
				}
			}

			resultChan <- result
		}(accInfo)
	}

	// Wait for all downloads to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	var newAccounts []account.Account
	storage := account.NewStorage(accountsDir)
	successCount := 0
	failCount := 0

	for result := range resultChan {
		var acc account.Account

		// Try auth_json first
		if len(result.authJSON) > 0 {
			if err := json.Unmarshal(result.authJSON, &acc); err != nil {
				logger.Warn("解析 %s 的 auth_json 失败：%v", result.fileName, err)
				failCount++
				continue
			}
		} else if result.success {
			// Use downloaded account
			acc = result.account
		} else {
			logger.Warn("下载 %s 失败", result.fileName)
			failCount++
			continue
		}

		// Set file path
		acc.FilePath = filepath.Join(accountsDir, result.fileName)

		// Save account
		if err := storage.Save(&acc); err != nil {
			logger.Warn("保存账号 %s 失败：%v", result.fileName, err)
			failCount++
			continue
		}

		logger.Info("已保存新账号：%s", result.fileName)
		newAccounts = append(newAccounts, acc)
		successCount++
	}

	logger.Info("HTTP 下载完成：成功=%d, 失败=%d", successCount, failCount)

	return newAccounts, nil
}

type downloadResult struct {
	fileName string
	authJSON json.RawMessage
	account  account.Account
	success  bool
}

func downloadAccountFromURL(url string, httpClient *http.Client) (account.Account, error) {
	var acc account.Account

	resp, err := httpClient.Get(url)
	if err != nil {
		return acc, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return acc, fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return acc, fmt.Errorf("读取响应失败: %w", err)
	}

	if err := json.Unmarshal(body, &acc); err != nil {
		return acc, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return acc, nil
}
