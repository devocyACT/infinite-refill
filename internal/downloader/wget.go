package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/topup"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// DownloadWithWget downloads accounts using wget for better compatibility
func DownloadWithWget(accounts []topup.AccountInfo, accountsDir string, proxyURL string, maxConcurrent int) ([]account.Account, error) {
	if len(accounts) == 0 {
		return nil, nil
	}

	// Check if wget is available
	if _, err := exec.LookPath("wget"); err != nil {
		logger.Debug("wget 不可用，回退到标准下载")
		return nil, fmt.Errorf("wget not available: %w", err)
	}

	logger.Info("使用 wget 并发下载 %d 个账号（并发数=%d）", len(accounts), maxConcurrent)

	// Create temp directory for downloads
	tempDir := filepath.Join(accountsDir, ".wget_temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download files concurrently using wget
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrent)
	resultChan := make(chan downloadResult, len(accounts))

	for _, accInfo := range accounts {
		if accInfo.DownloadURL == "" && len(accInfo.AuthJSON) == 0 {
			continue
		}

		wg.Add(1)
		go func(info topup.AccountInfo) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := downloadResult{
				fileName: info.FileName,
				authJSON: info.AuthJSON,
			}

			if info.DownloadURL != "" {
				tempFile := filepath.Join(tempDir, info.FileName)
				result.success = downloadFileWithWget(info.DownloadURL, tempFile, proxyURL)
				result.tempFile = tempFile
			} else {
				result.success = true
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
		} else if result.success && result.tempFile != "" {
			// Load from downloaded file
			data, err := os.ReadFile(result.tempFile)
			if err != nil {
				logger.Warn("读取下载文件 %s 失败：%v", result.fileName, err)
				failCount++
				continue
			}

			if err := json.Unmarshal(data, &acc); err != nil {
				logger.Warn("解析下载文件 %s 失败：%v", result.fileName, err)
				failCount++
				continue
			}
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

	logger.Info("wget 下载完成：成功=%d, 失败=%d", successCount, failCount)

	return newAccounts, nil
}

type downloadResult struct {
	fileName string
	tempFile string
	authJSON json.RawMessage
	success  bool
}

func downloadFileWithWget(url, outputFile, proxyURL string) bool {
	args := []string{
		"--quiet",
		"--timeout=30",
		"--dns-timeout=10",
		"--connect-timeout=10",
		"--read-timeout=30",
		"--tries=3",
		"--waitretry=2",
		"--no-check-certificate", // Skip SSL verification for faster connection
		"--output-document=" + outputFile,
	}

	// Add proxy if configured
	if proxyURL != "" {
		// Parse proxy URL to determine type
		if strings.HasPrefix(proxyURL, "http://") || strings.HasPrefix(proxyURL, "https://") {
			args = append(args, "--execute", "use_proxy=yes")
			args = append(args, "--execute", "http_proxy="+proxyURL)
			args = append(args, "--execute", "https_proxy="+proxyURL)
			logger.Debug("wget 使用代理: %s", proxyURL)
		} else if strings.HasPrefix(proxyURL, "socks5://") {
			// wget doesn't support socks5 directly, skip proxy
			logger.Debug("wget 不支持 socks5 代理，使用直连")
		}
	}

	args = append(args, url)

	cmd := exec.Command("wget", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("wget 下载失败: %s - %v: %s", url, err, string(output))
		return false
	}

	return true
}
