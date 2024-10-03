package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/mmatongo/flowline/pkg/config"
	"github.com/sirupsen/logrus"
)

type App struct {
	Logger *logrus.Logger
	Print  func(a ...interface{})
	Error  error
}

func logToFileAndOutput(output io.Writer) io.Writer {
	LogDir := config.NewConfig().LogDir

	file, err := os.OpenFile(LogDir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		logrus.Fatalf("failed to open log file: %s", err)
	}
	return io.MultiWriter(file, output)
}

func NewLogger() *App {
	logger := logrus.New()
	logger.SetOutput(logToFileAndOutput(os.Stdout))
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	return &App{
		Logger: logger,
		Print: func(a ...interface{}) {
			logger.Info(fmt.Sprint(a...))
		},
		Error: nil,
	}
}
