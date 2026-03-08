package topup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/devocyACT/infinite-refill/internal/config"
	"github.com/devocyACT/infinite-refill/internal/probe"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// Client handles topup API requests
type Client struct {
	serverURL  string
	userKey    string
	config     *config.TopupConfig
	httpClient *http.Client
}

// TopupRequest represents a topup request
type TopupRequest struct {
	TargetPoolSize int                 `json:"target_pool_size"`
	Reports        []probe.ProbeResult `json:"reports"`
	AccountIDs     []string            `json:"account_ids"`
}

// TopupResponse represents a topup response
type TopupResponse struct {
	OK              bool          `json:"ok"`
	Accounts        []AccountInfo `json:"accounts"`
	AutoDisabled    bool          `json:"auto_disabled"`
	AbuseAutoBanned bool          `json:"abuse_auto_banned"`
	TotalHoldLimit  int           `json:"total_hold_limit"`
}

// AccountInfo represents account information in the response
type AccountInfo struct {
	FileName    string          `json:"file_name"`
	DownloadURL string          `json:"download_url,omitempty"`
	AuthJSON    json.RawMessage `json:"auth_json,omitempty"`
}

// NewClient creates a new topup client
func NewClient(serverURL, userKey string, cfg *config.TopupConfig, httpClient *http.Client) *Client {
	return &Client{
		serverURL:  serverURL,
		userKey:    userKey,
		config:     cfg,
		httpClient: httpClient,
	}
}

// GetHTTPClient returns the HTTP client
func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

// Topup sends a topup request to the server
func (c *Client) Topup(req *TopupRequest) (*TopupResponse, error) {
	url := fmt.Sprintf("%s/v1/refill/topup", c.serverURL)

	var lastErr error
	for attempt := 1; attempt <= c.config.Retry; attempt++ {
		logger.Info("Topup 请求尝试 %d/%d", attempt, c.config.Retry)

		resp, err := c.sendRequest(url, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		logger.Warn("Topup 尝试 %d 失败：%v", attempt, err)

		if attempt < c.config.Retry {
			logger.Info("将在 %v 后重试...", c.config.RetryDelay)
			time.Sleep(c.config.RetryDelay)
		}
	}

	return nil, fmt.Errorf("topup 在 %d 次尝试后失败：%w", c.config.Retry, lastErr)
}

func (c *Client) sendRequest(url string, req *TopupRequest) (*TopupResponse, error) {
	// Marshal request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-Key", c.userKey)

	// Send request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(body))
	}

	// Parse response
	var resp TopupResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// CheckResponse checks the topup response for special conditions
func CheckResponse(resp *TopupResponse) (exitCode int, err error) {
	if resp.AutoDisabled {
		return 4, fmt.Errorf("服务器已禁用自动续杯")
	}
	if resp.AbuseAutoBanned {
		return 5, fmt.Errorf("服务器检测到滥用，已自动封禁")
	}
	if !resp.OK {
		return 2, fmt.Errorf("topup 请求失败")
	}
	return 0, nil
}
