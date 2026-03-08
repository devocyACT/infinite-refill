package loop

import (
	"fmt"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/config"
	"github.com/devocyACT/infinite-refill/internal/downloader"
	"github.com/devocyACT/infinite-refill/internal/probe"
	"github.com/devocyACT/infinite-refill/internal/topup"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// LoopResult contains the result of the refill loop
type LoopResult struct {
	Iterations   int
	TotalAdded   int
	TotalDeleted int
	FinalCount   int
	ExitCode     int
}

// RefillLoop handles the incremental refill loop
type RefillLoop struct {
	config      *config.Config
	prober      *probe.Prober
	topupClient *topup.Client
	accountMgr  *account.Manager
}

// NewRefillLoop creates a new RefillLoop instance
func NewRefillLoop(cfg *config.Config, prober *probe.Prober, topupClient *topup.Client, accountMgr *account.Manager) *RefillLoop {
	return &RefillLoop{
		config:      cfg,
		prober:      prober,
		topupClient: topupClient,
		accountMgr:  accountMgr,
	}
}

// Run executes the refill loop
func (rl *RefillLoop) Run() (*LoopResult, error) {
	result := &LoopResult{}

	// Track new account IDs for incremental probing
	var incrementalIDs []string

	for iteration := 1; iteration <= rl.config.Loop.MaxIterations; iteration++ {
		logger.Info("=== 第 %d/%d 轮 ===", iteration, rl.config.Loop.MaxIterations)

		// Load current accounts
		accounts, err := rl.accountMgr.LoadAll()
		if err != nil {
			return nil, fmt.Errorf("加载账号失败：%w", err)
		}

		logger.Info("当前账号数量：%d", len(accounts))

		// Probe accounts
		var probeReport *probe.ProbeReport
		if iteration == 1 {
			// First iteration: probe all accounts
			logger.Info("第一轮：探测所有账号")
			probeReport = rl.prober.ProbeAll(accounts)
		} else if len(incrementalIDs) > 0 {
			// Subsequent iterations: probe only new accounts
			logger.Info("增量探测：%d 个新账号", len(incrementalIDs))
			probeReport = rl.prober.ProbeSubset(accounts, incrementalIDs)
		} else {
			// No new accounts to probe
			logger.Info("没有新账号需要探测，结束循环")
			break
		}

		// Count invalid accounts
		invalidCount := probeReport.Invalid + probeReport.NetFail

		// Determine if topup is needed
		needTopup := false
		if invalidCount > 0 {
			logger.Info("发现 %d 个失效账号", invalidCount)
			needTopup = true
		}
		if len(accounts) < rl.config.TargetPoolSize {
			logger.Info("账号数量（%d）低于目标（%d）", len(accounts), rl.config.TargetPoolSize)
			needTopup = true
		}

		if !needTopup {
			logger.Info("无需续杯，结束循环")
			break
		}

		// Check total hold limit
		availableHold := rl.config.TotalHoldLimit - len(accounts)
		if availableHold <= 0 {
			logger.Warn("已达到总持有上限（%d/%d）", len(accounts), rl.config.TotalHoldLimit)
			break
		}

		// Prepare topup request
		accountIDs := make([]string, 0, len(accounts))
		for _, acc := range accounts {
			accountIDs = append(accountIDs, acc.AccountID)
		}

		// Only include invalid reports (401, 429) in topup request
		invalidReports := make([]probe.ProbeResult, 0)
		for _, result := range probeReport.Results {
			if result.StatusCode == 401 || result.StatusCode == 429 {
				invalidReports = append(invalidReports, result)
			}
		}

		logger.Debug("Topup 请求：account_ids=%d, invalid_reports=%d", len(accountIDs), len(invalidReports))

		topupReq := &topup.TopupRequest{
			TargetPoolSize: rl.config.TargetPoolSize,
			Reports:        invalidReports,
			AccountIDs:     accountIDs,
		}

		// Send topup request
		logger.Info("发送 topup 请求...")
		topupResp, err := rl.topupClient.Topup(topupReq)
		if err != nil {
			return nil, fmt.Errorf("topup 请求失败：%w", err)
		}

		// Check response
		exitCode, err := topup.CheckResponse(topupResp)
		if err != nil {
			result.ExitCode = exitCode
			return result, err
		}

		// Download new accounts using optimized HTTP concurrent download
		logger.Info("下载 %d 个新账号", len(topupResp.Accounts))
		newAccounts, err := downloader.DownloadWithHTTP(topupResp.Accounts, rl.config.AccountsDir, rl.topupClient.GetHTTPClient(), 6)
		if err != nil {
			logger.Warn("下载账号失败：%v", err)
		}

		result.TotalAdded += len(newAccounts)

		// Update incremental IDs for next iteration
		incrementalIDs = make([]string, 0, len(newAccounts))
		for _, acc := range newAccounts {
			incrementalIDs = append(incrementalIDs, acc.AccountID)
		}

		// Delete invalid accounts
		for _, probeResult := range probeReport.Results {
			if probeResult.StatusCode == 401 || probeResult.StatusCode == 429 {
				acc, err := rl.accountMgr.LoadByID(probeResult.AccountID)
				if err != nil {
					logger.Warn("加载账号 %s 失败（准备删除）：%v", probeResult.AccountID, err)
					continue
				}

				if err := rl.accountMgr.Delete(acc); err != nil {
					logger.Warn("删除账号 %s 失败：%v", probeResult.AccountID, err)
					continue
				}

				logger.Info("已删除失效账号：%s（状态码=%d）", probeResult.AccountID, probeResult.StatusCode)
				result.TotalDeleted++
			}
		}

		result.Iterations = iteration

		// Check if we should continue
		if len(incrementalIDs) == 0 && invalidCount == 0 {
			logger.Info("没有新账号且无失效账号，结束循环")
			break
		}
	}

	// Get final count
	accounts, err := rl.accountMgr.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("加载最终账号列表失败：%w", err)
	}
	result.FinalCount = len(accounts)

	logger.Info("续杯循环完成：轮数=%d 新增=%d 删除=%d 最终=%d",
		result.Iterations, result.TotalAdded, result.TotalDeleted, result.FinalCount)

	return result, nil
}
