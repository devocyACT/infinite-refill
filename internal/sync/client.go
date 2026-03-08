package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// Client handles sync-all API requests
type Client struct {
	serverURL  string
	userKey    string
	httpClient *http.Client
}

// SyncAllResponse represents the sync-all API response
type SyncAllResponse struct {
	OK              bool          `json:"ok"`
	Accounts        []AccountInfo `json:"accounts"`
	AutoDisabled    bool          `json:"auto_disabled"`
	AbuseAutoBanned bool          `json:"abuse_auto_banned"`
}

// AccountInfo represents account information in the response
type AccountInfo struct {
	FileName    string          `json:"file_name"`
	DownloadURL string          `json:"download_url,omitempty"`
	AuthJSON    json.RawMessage `json:"auth_json,omitempty"`
}

// NewClient creates a new sync client
func NewClient(serverURL, userKey string, httpClient *http.Client) *Client {
	return &Client{
		serverURL:  serverURL,
		userKey:    userKey,
		httpClient: httpClient,
	}
}

// SyncAll fetches all accounts from the server
func (c *Client) SyncAll() (*SyncAllResponse, error) {
	url := fmt.Sprintf("%s/v1/refill/sync-all", c.serverURL)

	logger.Info("全量同步：POST %s", url)
	logger.Debug("Sync-all 请求 URL: %s", url)

	// Create request with empty JSON body
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Key", c.userKey)
	logger.Debug("Sync-all 请求头: X-User-Key=%s...%s", c.userKey[:8], c.userKey[len(c.userKey)-4:])

	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)
	logger.Debug("Sync-all 响应状态码: %d (耗时: %v)", resp.StatusCode, elapsed)

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug("Sync-all 响应体长度: %d bytes", len(body))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Debug("Sync-all 错误响应: %s", string(body))
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var syncResp SyncAllResponse
	if err := json.Unmarshal(body, &syncResp); err != nil {
		logger.Debug("Sync-all 响应解析失败: %s", string(body))
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	logger.Debug("Sync-all 响应: ok=%v, accounts=%d, auto_disabled=%v, abuse_auto_banned=%v",
		syncResp.OK, len(syncResp.Accounts), syncResp.AutoDisabled, syncResp.AbuseAutoBanned)

	if !syncResp.OK {
		return nil, fmt.Errorf("sync-all 请求失败")
	}

	return &syncResp, nil
}
