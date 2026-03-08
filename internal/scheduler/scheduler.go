package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/devocyACT/infinite-refill/internal/clean"
	"github.com/devocyACT/infinite-refill/internal/loop"
	"github.com/devocyACT/infinite-refill/pkg/logger"
)

// Scheduler handles periodic execution of refill tasks
type Scheduler struct {
	config     *SchedulerConfig
	cleaner    *clean.Cleaner
	refillLoop *loop.RefillLoop
	lockFile   string
	stopChan   chan struct{}
}

// SchedulerConfig contains scheduler configuration
type SchedulerConfig struct {
	IntervalMinutes int
	LockTimeout     int
	AccountsDir     string
}

// NewScheduler creates a new Scheduler instance
func NewScheduler(cfg *SchedulerConfig, cleaner *clean.Cleaner, refillLoop *loop.RefillLoop) *Scheduler {
	lockFile := filepath.Join(cfg.AccountsDir, ".refill.lock")
	return &Scheduler{
		config:     cfg,
		cleaner:    cleaner,
		refillLoop: refillLoop,
		lockFile:   lockFile,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	logger.Info("启动调度器（间隔=%d 分钟）", s.config.IntervalMinutes)

	ticker := time.NewTicker(time.Duration(s.config.IntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	if err := s.runOnce(); err != nil {
		logger.Error("初始运行失败：%v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.runOnce(); err != nil {
				logger.Error("定时运行失败：%v", err)
			}
		case <-s.stopChan:
			logger.Info("调度器已停止")
			return nil
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// runOnce executes one cycle of cleanup and refill
func (s *Scheduler) runOnce() error {
	// Acquire lock
	if !s.acquireLock() {
		logger.Warn("获取锁失败，可能有其他实例正在运行")
		return fmt.Errorf("获取锁失败")
	}
	defer s.releaseLock()

	logger.Info("=== 开始定时运行 ===")

	// Run cleanup
	logger.Info("运行清理...")
	cleanReport, err := s.cleaner.Clean(false, []string{})
	if err != nil {
		logger.Error("清理失败：%v", err)
	} else {
		logger.Info("清理完成：已删除 %d 个账号", cleanReport.Deleted)
	}

	// Run refill loop
	logger.Info("运行续杯循环...")
	loopResult, err := s.refillLoop.Run()
	if err != nil {
		logger.Error("续杯循环失败：%v", err)

		// Check for special exit codes
		if loopResult != nil {
			if loopResult.ExitCode == 4 {
				logger.Error("服务器已禁用自动续杯，停止调度器")
				s.Stop()
				return fmt.Errorf("自动续杯已禁用")
			}
			if loopResult.ExitCode == 5 {
				logger.Error("服务器检测到滥用已封禁，停止调度器")
				s.Stop()
				return fmt.Errorf("滥用已封禁")
			}
		}

		return err
	}

	logger.Info("续杯循环完成：新增=%d 删除=%d 最终=%d",
		loopResult.TotalAdded, loopResult.TotalDeleted, loopResult.FinalCount)

	logger.Info("=== 定时运行完成 ===")
	return nil
}

// acquireLock attempts to acquire the lock file
func (s *Scheduler) acquireLock() bool {
	// Check if lock file exists
	if info, err := os.Stat(s.lockFile); err == nil {
		// Lock file exists, check if it's stale
		age := time.Since(info.ModTime())
		if age < time.Duration(s.config.LockTimeout)*time.Second {
			logger.Debug("Lock file exists and is fresh (age=%v)", age)
			return false
		}
		logger.Warn("Lock file is stale (age=%v), removing", age)
		os.Remove(s.lockFile)
	}

	// Create lock file
	pid := os.Getpid()
	if err := os.WriteFile(s.lockFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		logger.Error("Failed to create lock file: %v", err)
		return false
	}

	logger.Debug("Acquired lock (pid=%d)", pid)
	return true
}

// releaseLock releases the lock file
func (s *Scheduler) releaseLock() {
	if err := os.Remove(s.lockFile); err != nil {
		logger.Warn("Failed to remove lock file: %v", err)
	} else {
		logger.Debug("Released lock")
	}
}
