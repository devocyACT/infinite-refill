package httpclient

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/devocyACT/infinite-refill/internal/config"
)

// NewClient creates an HTTP client with the specified configuration
func NewClient(cfg *config.Config, forWham bool) (*http.Client, error) {
	transport := &http.Transport{}

	// Determine proxy mode
	proxyMode := cfg.Proxy.Mode
	if cfg.Proxy.Mode == "mixed" {
		if forWham {
			proxyMode = cfg.Proxy.WhamMode
		} else {
			proxyMode = cfg.Proxy.ServerMode
		}
	}

	// Configure proxy
	switch proxyMode {
	case "direct":
		transport.Proxy = nil
	case "proxy":
		if cfg.Proxy.URL != "" {
			proxyURL, err := url.Parse(cfg.Proxy.URL)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL: %w", err)
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	case "auto":
		// Try configured proxy first, then fall back to environment
		if cfg.Proxy.URL != "" {
			proxyURL, err := url.Parse(cfg.Proxy.URL)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL: %w", err)
			}
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			transport.Proxy = http.ProxyFromEnvironment
		}
	default:
		transport.Proxy = http.ProxyFromEnvironment
	}

	// Set timeouts based on usage
	var timeout time.Duration
	if forWham {
		timeout = cfg.Probe.MaxTime
	} else {
		timeout = cfg.Topup.MaxTime
	}

	// Set transport timeouts
	transport.DialContext = (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ResponseHeaderTimeout = timeout
	transport.ExpectContinueTimeout = 1 * time.Second

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}

// GetProxyInfo returns proxy information for debugging
func GetProxyInfo(cfg *config.Config) string {
	if cfg.Proxy.Mode == "direct" {
		return "direct (no proxy)"
	}
	if cfg.Proxy.URL != "" {
		return fmt.Sprintf("%s (%s)", cfg.Proxy.URL, cfg.Proxy.Mode)
	}
	if httpProxy := os.Getenv("HTTP_PROXY"); httpProxy != "" {
		return fmt.Sprintf("%s (from HTTP_PROXY)", httpProxy)
	}
	if httpsProxy := os.Getenv("HTTPS_PROXY"); httpsProxy != "" {
		return fmt.Sprintf("%s (from HTTPS_PROXY)", httpsProxy)
	}
	return "auto (no proxy configured)"
}
