package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/devocyACT/infinite-refill/internal/account"
	"github.com/devocyACT/infinite-refill/internal/clean"
	"github.com/devocyACT/infinite-refill/internal/config"
	"github.com/devocyACT/infinite-refill/internal/httpclient"
	"github.com/devocyACT/infinite-refill/internal/loop"
	"github.com/devocyACT/infinite-refill/internal/probe"
	"github.com/devocyACT/infinite-refill/internal/scheduler"
	"github.com/devocyACT/infinite-refill/internal/topup"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

var (
	Version   = "dev"
	BuildTime = "unknown"

	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "refill",
	Short: "Infinite Refill - Account renewal management tool",
	Long:  `A tool for automatic account renewal management with probing, topup, and cleanup capabilities.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			logger.SetLevel(logger.DEBUG)
		}
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a single refill cycle",
	Long:  `Executes one complete refill cycle with incremental probing and topup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create components
		accountMgr := account.NewManager(cfg.AccountsDir)

		whamClient, err := httpclient.NewClient(cfg, true)
		if err != nil {
			return fmt.Errorf("failed to create WHAM client: %w", err)
		}

		serverClient, err := httpclient.NewClient(cfg, false)
		if err != nil {
			return fmt.Errorf("failed to create server client: %w", err)
		}

		prober := probe.NewProber(&cfg.Probe, whamClient)
		topupClient := topup.NewClient(cfg.ServerURL, cfg.UserKey, &cfg.Topup, serverClient)
		refillLoop := loop.NewRefillLoop(cfg, prober, topupClient, accountMgr)

		// Run refill loop
		logger.Info("开始续杯循环")
		result, err := refillLoop.Run()
		if err != nil {
			logger.Error("续杯失败：%v", err)
			if result != nil && result.ExitCode != 0 {
				os.Exit(result.ExitCode)
			}
			return err
		}

		logger.Info("续杯完成：轮数=%d 新增=%d 删除=%d 最终=%d",
			result.Iterations, result.TotalAdded, result.TotalDeleted, result.FinalCount)

		return nil
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all accounts (full probe)",
	Long:  `Probes all accounts and generates a report without performing topup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create components
		accountMgr := account.NewManager(cfg.AccountsDir)

		whamClient, err := httpclient.NewClient(cfg, true)
		if err != nil {
			return fmt.Errorf("failed to create WHAM client: %w", err)
		}

		prober := probe.NewProber(&cfg.Probe, whamClient)

		// Load and probe all accounts
		accounts, err := accountMgr.LoadAll()
		if err != nil {
			return fmt.Errorf("加载账号失败：%w", err)
		}

		logger.Info("探测 %d 个账号", len(accounts))
		report := prober.ProbeAll(accounts)

		// Save report
		reportFile, err := probe.SaveReport(report, "out")
		if err != nil {
			return fmt.Errorf("保存报告失败：%w", err)
		}

		logger.Info("探测完成：总数=%d 成功=%d 网络失败=%d 失效=%d",
			report.Total, report.Success, report.NetFail, report.Invalid)
		logger.Info("报告已保存到：%s", reportFile)

		return nil
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean invalid and expired accounts",
	Long:  `Removes accounts that are invalid (401/429) or expired (older than configured days).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")

		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create components
		accountMgr := account.NewManager(cfg.AccountsDir)

		whamClient, err := httpclient.NewClient(cfg, true)
		if err != nil {
			return fmt.Errorf("failed to create WHAM client: %w", err)
		}

		prober := probe.NewProber(&cfg.Probe, whamClient)
		cleaner := clean.NewCleaner(&cfg.Clean, prober, accountMgr)

		// Run cleanup
		logger.Info("开始清理（应用=%v）", apply)
		report, err := cleaner.Clean(!apply, []string{})
		if err != nil {
			return fmt.Errorf("清理失败：%w", err)
		}

		// Save report
		reportFile, err := clean.SaveCleanReport(report, "out")
		if err != nil {
			logger.Warn("保存报告失败：%v", err)
		} else {
			logger.Info("报告已保存到：%s", reportFile)
		}

		logger.Info("清理完成：候选=%d 已删除=%d", report.Candidates, report.Deleted)

		return nil
	},
}

var schedulerCmd = &cobra.Command{
	Use:   "scheduler",
	Short: "Scheduler commands",
	Long:  `Start or stop the periodic scheduler for automatic refill.`,
}

var schedulerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the scheduler",
	Long:  `Starts the periodic scheduler that runs cleanup and refill at configured intervals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Create components
		accountMgr := account.NewManager(cfg.AccountsDir)

		whamClient, err := httpclient.NewClient(cfg, true)
		if err != nil {
			return fmt.Errorf("failed to create WHAM client: %w", err)
		}

		serverClient, err := httpclient.NewClient(cfg, false)
		if err != nil {
			return fmt.Errorf("failed to create server client: %w", err)
		}

		prober := probe.NewProber(&cfg.Probe, whamClient)
		topupClient := topup.NewClient(cfg.ServerURL, cfg.UserKey, &cfg.Topup, serverClient)
		refillLoop := loop.NewRefillLoop(cfg, prober, topupClient, accountMgr)
		cleaner := clean.NewCleaner(&cfg.Clean, prober, accountMgr)

		schedulerCfg := &scheduler.SchedulerConfig{
			IntervalMinutes: cfg.Scheduler.IntervalMinutes,
			LockTimeout:     cfg.Probe.WaitTimeout,
			AccountsDir:     cfg.AccountsDir,
		}

		sched := scheduler.NewScheduler(schedulerCfg, cleaner, refillLoop)

		// Start scheduler
		logger.Info("启动调度器")
		return sched.Start()
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check environment and configuration",
	Long:  `Validates the configuration and checks connectivity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		logger.Info("配置检查：")
		logger.Info("  服务器地址：%s", cfg.ServerURL)
		logger.Info("  用户密钥：%s", maskKey(cfg.UserKey))
		logger.Info("  账号目录：%s", cfg.AccountsDir)
		logger.Info("  目标池大小：%d", cfg.TargetPoolSize)
		logger.Info("  总持有上限：%d", cfg.TotalHoldLimit)
		logger.Info("  代理：%s", httpclient.GetProxyInfo(cfg))

		// Check accounts directory
		accountMgr := account.NewManager(cfg.AccountsDir)
		count, err := accountMgr.Count()
		if err != nil {
			logger.Warn("统计账号失败：%v", err)
		} else {
			logger.Info("  当前账号数：%d", count)
		}

		logger.Info("环境检查通过")
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("refill version %s (built %s)\n", Version, BuildTime)
	},
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (optional)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	cleanCmd.Flags().Bool("apply", false, "actually delete accounts (default is dry-run)")

	schedulerCmd.AddCommand(schedulerStartCmd)

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(schedulerCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
