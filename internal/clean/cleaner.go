package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/config"
	"github.com/devocyACT/infinite-refill/internal/probe"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// CleanReport contains statistics from cleaning operation
type CleanReport struct {
	Total      int
	Probed     int
	NetFail    int
	Candidates int
	Deleted    int
	Reasons    map[string]int
}

// Cleaner handles account cleanup
type Cleaner struct {
	config     *config.CleanConfig
	prober     *probe.Prober
	accountMgr *account.Manager
}

// NewCleaner creates a new Cleaner instance
func NewCleaner(cfg *config.CleanConfig, prober *probe.Prober, accountMgr *account.Manager) *Cleaner {
	return &Cleaner{
		config:     cfg,
		prober:     prober,
		accountMgr: accountMgr,
	}
}

// Clean performs cleanup of invalid and expired accounts
func (c *Cleaner) Clean(dryRun bool, excludeIDs []string) (*CleanReport, error) {
	logger.Info("开始清理（预览模式=%v）", dryRun)

	// Load all accounts
	accounts, err := c.accountMgr.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("加载账号失败：%w", err)
	}

	report := &CleanReport{
		Total:   len(accounts),
		Reasons: make(map[string]int),
	}

	// Create exclude map
	excludeMap := make(map[string]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	// Probe all accounts
	logger.Info("探测 %d 个账号以进行清理", len(accounts))
	probeReport := c.prober.ProbeAll(accounts)
	report.Probed = probeReport.Total
	report.NetFail = probeReport.NetFail

	// Create status map
	statusMap := make(map[string]int)
	for _, result := range probeReport.Results {
		statusMap[result.AccountID] = result.StatusCode
	}

	// Identify candidates for deletion
	var candidates []account.Account
	for _, acc := range accounts {
		// Skip excluded accounts
		if excludeMap[acc.AccountID] {
			logger.Debug("Skipping excluded account: %s", acc.AccountID)
			continue
		}

		reason := c.shouldDelete(&acc, statusMap)
		if reason != "" {
			candidates = append(candidates, acc)
			report.Reasons[reason]++
			logger.Debug("Candidate for deletion: %s (reason: %s)", acc.AccountID, reason)
		}
	}

	report.Candidates = len(candidates)

	if dryRun {
		logger.Info("预览模式：将删除 %d 个账号", len(candidates))
		return report, nil
	}

	// Create backup directory
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join("out", fmt.Sprintf("清理-%s", timestamp), "backup")

	// Delete candidates
	for _, acc := range candidates {
		// Backup first
		if err := c.accountMgr.Backup(&acc, backupDir); err != nil {
			logger.Warn("备份 %s 失败：%v", acc.AccountID, err)
			continue
		}

		// Delete
		if err := c.accountMgr.Delete(&acc); err != nil {
			logger.Warn("删除 %s 失败：%v", acc.AccountID, err)
			continue
		}

		logger.Info("已删除账号：%s", acc.AccountID)
		report.Deleted++
	}

	logger.Info("清理完成：已删除 %d / %d 个候选账号", report.Deleted, report.Candidates)
	return report, nil
}

// shouldDelete determines if an account should be deleted
func (c *Cleaner) shouldDelete(acc *account.Account, statusMap map[string]int) string {
	// Check status code
	if status, ok := statusMap[acc.AccountID]; ok {
		for _, deleteStatus := range c.config.DeleteStatuses {
			if status == deleteStatus {
				return fmt.Sprintf("status_%d", status)
			}
		}
	}

	// Check expiration
	if acc.IsExpired(c.config.ExpiredDays) {
		return "expired"
	}

	return ""
}

// SaveCleanReport saves the clean report to a file
func SaveCleanReport(report *CleanReport, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(outputDir, fmt.Sprintf("clean_report_%s.txt", timestamp))

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "Clean Report\n")
	fmt.Fprintf(file, "============\n\n")
	fmt.Fprintf(file, "Total accounts: %d\n", report.Total)
	fmt.Fprintf(file, "Probed: %d\n", report.Probed)
	fmt.Fprintf(file, "Network failures: %d\n", report.NetFail)
	fmt.Fprintf(file, "Candidates: %d\n", report.Candidates)
	fmt.Fprintf(file, "Deleted: %d\n\n", report.Deleted)
	fmt.Fprintf(file, "Reasons:\n")
	for reason, count := range report.Reasons {
		fmt.Fprintf(file, "  %s: %d\n", reason, count)
	}

	return filename, nil
}
