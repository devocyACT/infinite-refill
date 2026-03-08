package downloader

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/topup"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// DownloadWithAria2c downloads accounts using aria2c for better performance
func DownloadWithAria2c(accounts []topup.AccountInfo, accountsDir string, proxyURL string, maxConcurrent int) ([]account.Account, error) {
	if len(accounts) == 0 {
		return nil, nil
	}

	// Check if aria2c is available
	if _, err := exec.LookPath("aria2c"); err != nil {
		logger.Debug("aria2c 不可用，回退到标准下载")
		return nil, fmt.Errorf("aria2c not available: %w", err)
	}

	logger.Info("使用 aria2c 并发下载 %d 个账号（并发数=%d）", len(accounts), maxConcurrent)

	// Create temp directory for downloads
	tempDir := filepath.Join(accountsDir, ".aria2c_temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败：%w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create aria2c input file
	inputFile := filepath.Join(tempDir, "input.txt")
	f, err := os.Create(inputFile)
	if err != nil {
		return nil, fmt.Errorf("创建输入文件失败：%w", err)
	}

	// Write download URLs to input file
	urlToFileName := make(map[string]string)
	for _, acc := range accounts {
		if acc.DownloadURL != "" {
			// aria2c input format: URL\n  out=filename\n
			fmt.Fprintf(f, "%s\n", acc.DownloadURL)
			fmt.Fprintf(f, "  out=%s\n", acc.FileName)
			urlToFileName[acc.DownloadURL] = acc.FileName
		}
	}
	f.Close()

	if len(urlToFileName) == 0 {
		logger.Debug("没有需要下载的 URL")
		return nil, nil
	}

	// Build aria2c command
	args := []string{
		"--input-file=" + inputFile,
		"--dir=" + tempDir,
		fmt.Sprintf("--max-concurrent-downloads=%d", maxConcurrent),
		"--max-connection-per-server=4",
		"--split=4",
		"--min-split-size=1M",
		"--connect-timeout=10",
		"--timeout=30",
		"--max-tries=3",
		"--retry-wait=2",
		"--console-log-level=warn",
		"--summary-interval=0",
		"--download-result=hide",
	}

	// Add proxy if configured
	if proxyURL != "" {
		args = append(args, "--all-proxy="+proxyURL)
		logger.Debug("aria2c 使用代理: %s", proxyURL)
	}

	// Execute aria2c
	logger.Debug("执行 aria2c 命令")
	cmd := exec.Command("aria2c", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Warn("aria2c 下载失败：%v\n%s", err, string(output))
		return nil, fmt.Errorf("aria2c 下载失败：%w", err)
	}

	logger.Debug("aria2c 下载完成")

	// Process downloaded files
	var newAccounts []account.Account
	storage := account.NewStorage(accountsDir)

	for _, accInfo := range accounts {
		var acc account.Account

		// Try auth_json first
		if len(accInfo.AuthJSON) > 0 {
			if err := json.Unmarshal(accInfo.AuthJSON, &acc); err != nil {
				logger.Warn("解析 %s 的 auth_json 失败：%v", accInfo.FileName, err)
				continue
			}
		} else if accInfo.DownloadURL != "" {
			// Load from downloaded file
			tempFile := filepath.Join(tempDir, accInfo.FileName)
			data, err := os.ReadFile(tempFile)
			if err != nil {
				logger.Warn("读取下载文件 %s 失败：%v", accInfo.FileName, err)
				continue
			}

			if err := json.Unmarshal(data, &acc); err != nil {
				logger.Warn("解析下载文件 %s 失败：%v", accInfo.FileName, err)
				continue
			}
		} else {
			logger.Warn("%s 没有 auth_json 或 download_url", accInfo.FileName)
			continue
		}

		// Set file path
		acc.FilePath = filepath.Join(accountsDir, accInfo.FileName)

		// Save account
		if err := storage.Save(&acc); err != nil {
			logger.Warn("保存账号 %s 失败：%v", accInfo.FileName, err)
			continue
		}

		logger.Info("已保存新账号：%s", accInfo.FileName)
		newAccounts = append(newAccounts, acc)
	}

	logger.Info("aria2c 下载完成：成功=%d, 失败=%d", len(newAccounts), len(accounts)-len(newAccounts))

	return newAccounts, nil
}
