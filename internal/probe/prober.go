package probe

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/config"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

const whamURL = "https://chatgpt.com/backend-api/wham/usage"

// ProbeResult represents the result of probing an account
type ProbeResult struct {
	FileName   string    `json:"file_name"`
	EmailHash  string    `json:"email_hash"`
	AccountID  string    `json:"account_id"`
	StatusCode int       `json:"status_code"`
	ProbedAt   time.Time `json:"probed_at"`
	Error      error     `json:"-"`
}

// ProbeReport contains statistics and results from probing
type ProbeReport struct {
	Total   int
	Success int
	NetFail int
	Invalid int
	Results []ProbeResult
}

// Prober handles account probing
type Prober struct {
	config     *config.ProbeConfig
	httpClient *http.Client
}

// NewProber creates a new Prober instance
func NewProber(cfg *config.ProbeConfig, httpClient *http.Client) *Prober {
	return &Prober{
		config:     cfg,
		httpClient: httpClient,
	}
}

// ProbeAccount probes a single account
func (p *Prober) ProbeAccount(acc *account.Account) ProbeResult {
	result := ProbeResult{
		FileName:  filepath.Base(acc.FilePath),
		AccountID: acc.AccountID,
		ProbedAt:  time.Now(),
	}

	// Generate email hash
	if acc.Email != "" {
		hash := md5.Sum([]byte(acc.Email))
		result.EmailHash = fmt.Sprintf("%x", hash)
	}

	// Create request with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.config.MaxTime)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", whamURL, nil)
	if err != nil {
		result.StatusCode = 0
		result.Error = err
		return result
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", acc.AccessToken))
	req.Header.Set("Chatgpt-Account-Id", acc.AccountID)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		result.StatusCode = 0
		result.Error = err
		logger.Debug("Probe failed for %s: %v", acc.AccountID, err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	logger.Debug("Probed %s: status=%d", acc.AccountID, resp.StatusCode)

	return result
}

// ProbeAll probes all accounts concurrently
func (p *Prober) ProbeAll(accounts []account.Account) *ProbeReport {
	logger.Info("开始探测 %d 个账号（并发数=%d）", len(accounts), p.config.Parallel)

	results := make([]ProbeResult, 0, len(accounts))

	// Start worker pool
	pool := NewWorkerPool(p.config.Parallel, p.config.WaitTimeout)
	pool.Start()

	// Submit jobs
	for _, acc := range accounts {
		acc := acc // capture loop variable
		pool.Submit(func() ProbeResult {
			return p.ProbeAccount(&acc)
		})
	}

	// Wait for completion
	pool.Wait()

	// Collect results
	for result := range pool.Results() {
		results = append(results, result)
	}

	// Generate report
	report := &ProbeReport{
		Total:   len(results),
		Results: results,
	}

	for _, r := range results {
		if r.StatusCode == 0 {
			report.NetFail++
		} else if r.StatusCode == 200 {
			report.Success++
		} else {
			report.Invalid++
		}
	}

	logger.Info("探测完成：总数=%d 成功=%d 网络失败=%d 失效=%d",
		report.Total, report.Success, report.NetFail, report.Invalid)

	return report
}

// ProbeSubset probes a subset of accounts
func (p *Prober) ProbeSubset(accounts []account.Account, accountIDs []string) *ProbeReport {
	// Create a map for quick lookup
	idMap := make(map[string]bool)
	for _, id := range accountIDs {
		idMap[id] = true
	}

	// Filter accounts
	subset := make([]account.Account, 0)
	for _, acc := range accounts {
		if idMap[acc.AccountID] {
			subset = append(subset, acc)
		}
	}

	logger.Info("探测子集：%d / %d 个账号", len(subset), len(accounts))
	return p.ProbeAll(subset)
}
