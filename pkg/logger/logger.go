package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	currentLevel = INFO
	logger       = log.New(os.Stdout, "", 0)
)

func SetLevel(level Level) {
	currentLevel = level
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Debug(format string, args ...interface{}) {
	if currentLevel <= DEBUG {
		msg := fmt.Sprintf(format, args...)
		logger.Printf("[%s] DEBUG: %s", timestamp(), msg)
	}
}

func Info(format string, args ...interface{}) {
	if currentLevel <= INFO {
		msg := fmt.Sprintf(format, args...)
		logger.Printf("[%s] INFO: %s", timestamp(), msg)
	}
}

func Warn(format string, args ...interface{}) {
	if currentLevel <= WARN {
		msg := fmt.Sprintf(format, args...)
		logger.Printf("[%s] WARN: %s", timestamp(), msg)
	}
}

func Error(format string, args ...interface{}) {
	if currentLevel <= ERROR {
		msg := fmt.Sprintf(format, args...)
		logger.Printf("[%s] ERROR: %s", timestamp(), msg)
	}
}

func Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	logger.Printf("[%s] FATAL: %s", timestamp(), msg)
	os.Exit(1)
}
