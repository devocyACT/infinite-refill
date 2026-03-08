package config

import (
	"time"
)

type Config struct {
	ServerURL      string
	UserKey        string
	AccountsDir    string
	TargetPoolSize int
	TotalHoldLimit int

	Probe     ProbeConfig
	Topup     TopupConfig
	Clean     CleanConfig
	Scheduler SchedulerConfig
	Proxy     ProxyConfig
	Loop      LoopConfig
}

type ProbeConfig struct {
	Parallel       int
	ConnectTimeout time.Duration
	MaxTime        time.Duration
	WaitTimeout    int
}

type TopupConfig struct {
	ConnectTimeout time.Duration
	MaxTime        time.Duration
	Retry          int
	RetryDelay     time.Duration
}

type CleanConfig struct {
	DeleteStatuses []int
	ExpiredDays    int
}

type ProxyConfig struct {
	Mode       string
	URL        string
	WhamMode   string
	ServerMode string
}

type LoopConfig struct {
	MaxIterations int
}

type SchedulerConfig struct {
	Enabled         bool
	IntervalMinutes int
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	return &Config{
		AccountsDir:    "./accounts",
		TargetPoolSize: 10,
		TotalHoldLimit: 50,
		Probe: ProbeConfig{
			Parallel:       6,
			ConnectTimeout: 5 * time.Second,
			MaxTime:        15 * time.Second,
			WaitTimeout:    600,
		},
		Topup: TopupConfig{
			ConnectTimeout: 10 * time.Second,
			MaxTime:        180 * time.Second,
			Retry:          3,
			RetryDelay:     3 * time.Second,
		},
		Clean: CleanConfig{
			DeleteStatuses: []int{401, 429},
			ExpiredDays:    30,
		},
		Proxy: ProxyConfig{
			Mode:       "auto",
			WhamMode:   "auto",
			ServerMode: "auto",
		},
		Loop: LoopConfig{
			MaxIterations: 6,
		},
		Scheduler: SchedulerConfig{
			Enabled:         false,
			IntervalMinutes: 30,
		},
	}
}
