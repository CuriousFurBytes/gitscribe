package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var (
	Logger  *log.Logger
	LogFile string
)

func Init() error {
	f, err := os.CreateTemp("", "gitscribe-*.log")
	if err != nil {
		return err
	}
	LogFile = f.Name()
	Logger = log.New(f)
	Logger.SetLevel(log.DebugLevel)
	Logger.SetTimeFormat("15:04:05")
	Logger.SetReportCaller(false)
	return nil
}

func Debug(msg string, args ...any) {
	if Logger != nil {
		Logger.Debug(msg, args...)
	}
}

func Info(msg string, args ...any) {
	if Logger != nil {
		Logger.Info(msg, args...)
	}
}

func Warn(msg string, args ...any) {
	if Logger != nil {
		Logger.Warn(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if Logger != nil {
		Logger.Error(msg, args...)
	}
}
