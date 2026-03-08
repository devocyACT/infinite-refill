package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load loads configuration from environment variables and optional YAML file
func Load(configFile string) (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from YAML file if provided
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err == nil {
			if err := viper.Unmarshal(cfg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}
	}

	// Override with environment variables (priority over YAML)
	loadFromEnv(cfg)

	// Validate required fields
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadFromEnv(cfg *Config) {
	if v := os.Getenv("SERVER_URL"); v != "" {
		cfg.ServerURL = v
	}
	if v := os.Getenv("USER_KEY"); v != "" {
		cfg.UserKey = v
	}
	if v := os.Getenv("ACCOUNTS_DIR"); v != "" {
		cfg.AccountsDir = v
	}
	if v := os.Getenv("TARGET_POOL_SIZE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.TargetPoolSize = i
		}
	}
	if v := os.Getenv("TOTAL_HOLD_LIMIT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.TotalHoldLimit = i
		}
	}

	// Probe config
	if v := os.Getenv("PROBE_PARALLEL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Probe.Parallel = i
		}
	}
	if v := os.Getenv("PROBE_CONNECT_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Probe.ConnectTimeout = time.Duration(i) * time.Second
		}
	}
	if v := os.Getenv("PROBE_MAX_TIME"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Probe.MaxTime = time.Duration(i) * time.Second
		}
	}
	if v := os.Getenv("PROBE_WAIT_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Probe.WaitTimeout = i
		}
	}

	// Topup config
	if v := os.Getenv("TOPUP_CONNECT_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Topup.ConnectTimeout = time.Duration(i) * time.Second
		}
	}
	if v := os.Getenv("TOPUP_MAX_TIME"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Topup.MaxTime = time.Duration(i) * time.Second
		}
	}
	if v := os.Getenv("TOPUP_RETRY"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Topup.Retry = i
		}
	}

	// Clean config
	if v := os.Getenv("CLEAN_DELETE_STATUSES"); v != "" {
		statuses := []int{}
		for _, s := range strings.Split(v, ",") {
			if i, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
				statuses = append(statuses, i)
			}
		}
		if len(statuses) > 0 {
			cfg.Clean.DeleteStatuses = statuses
		}
	}
	if v := os.Getenv("CLEAN_EXPIRED_DAYS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Clean.ExpiredDays = i
		}
	}

	// Proxy config
	if v := os.Getenv("PROXY_MODE"); v != "" {
		cfg.Proxy.Mode = v
	}
	if v := os.Getenv("PROXY_URL"); v != "" {
		cfg.Proxy.URL = v
	}
	if v := os.Getenv("WHAM_PROXY_MODE"); v != "" {
		cfg.Proxy.WhamMode = v
	}
	if v := os.Getenv("SERVER_PROXY_MODE"); v != "" {
		cfg.Proxy.ServerMode = v
	}

	// Loop config
	if v := os.Getenv("LOOP_MAX_ITERATIONS"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Loop.MaxIterations = i
		}
	}

	// Scheduler config
	if v := os.Getenv("SCHEDULER_ENABLED"); v != "" {
		cfg.Scheduler.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("SCHEDULER_INTERVAL_MINUTES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.Scheduler.IntervalMinutes = i
		}
	}
}

func validate(cfg *Config) error {
	if cfg.ServerURL == "" {
		return fmt.Errorf("SERVER_URL is required")
	}
	if cfg.UserKey == "" {
		return fmt.Errorf("USER_KEY is required")
	}
	if cfg.AccountsDir == "" {
		return fmt.Errorf("ACCOUNTS_DIR is required")
	}
	return nil
}
