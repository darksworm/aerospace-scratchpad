package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilmars/aerospace-sticky/internal/constants"
)

var defaultLogger Logger

type LogConfig struct {
	// Path to the log file
	Path string `json:"path"`
	// Log level
	Level string `json:"level"`
}

type Logger interface {
	// Info logs an informational message
	LogInfo(msg string, args ...any)
	// Error logs an error message
	LogError(msg string, args ...any)
	// Debug logs a debug message
	LogDebug(msg string, args ...any)

	// GetConfig returns the logger configuration
	GetConfig() LogConfig

	// AsJson returns the logger as a JSON object
	// In error, logs the error and returns an empty string
	AsJson(data any) string

	// Close closes the logger
	Close() error
}

type LoggerClient struct {
	logger *slog.Logger
	file   *os.File
	config LogConfig
}

func (l *LoggerClient) LogInfo(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *LoggerClient) LogError(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *LoggerClient) LogDebug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *LoggerClient) GetConfig() LogConfig {
	return l.config
}

func (l *LoggerClient) AsJson(data any) string {
	json, err := json.Marshal(data)
	if err != nil {
		l.LogError("failed to marshal data to JSON", err)
		return ""
	}
	return string(json)
}

func (l *LoggerClient) Close() error {
	if l.file != nil {
		err := l.file.Close()
		if err != nil {
			return fmt.Errorf("failed to close log file: %v", err)
		}
	}
	return nil
}

type EmptyLogger struct{}

func (l *EmptyLogger) LogInfo(msg string, args ...any) {
	// No-op
}
func (l *EmptyLogger) LogError(msg string, args ...any) {
	// No-op
}
func (l *EmptyLogger) LogDebug(msg string, args ...any) {
	// No-op
}
func (l *EmptyLogger) Close() error {
	// No-op
	return nil
}
func (l *EmptyLogger) GetConfig() LogConfig {
	// No-op
	return LogConfig{
		Path:  "/tmp/aerospace-marks.log",
		Level: "DISABLED",
	}
}
func (l *EmptyLogger) AsJson(data any) string {
	// No-op
	return ""
}

// NewLogger creates a new logger instance
// It accepts a path to a file where logs will be written
// and a boolean indicating whether to log to stdout as well
func NewLogger() (Logger, error) {
	path := os.Getenv(constants.EnvAeroSpaceScratchpadLogsPath)
	if path == "" {
		path = "/tmp/aerospace-scratchpad.log"
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	configLogLevel := os.Getenv(constants.EnvAeroSpaceScratchpadLogsLevel)
	if configLogLevel == "" {
		return &EmptyLogger{}, nil
	}

	logLevel := slog.LevelError
	if configLogLevel != "" {
		switch configLogLevel {
		case "DEBUG":
			logLevel = slog.LevelDebug
		case "INFO":
			logLevel = slog.LevelInfo
		case "WARN":
			logLevel = slog.LevelWarn
		default:
			logLevel = slog.LevelError
		}
	}

	textHandler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: logLevel,
	})

	newLogger := slog.New(textHandler)
	logClient := &LoggerClient{
		logger: newLogger,
		file:   file,
		config: LogConfig{
			Path:  path,
			Level: configLogLevel,
		},
	}

	return logClient, nil
}

func SetDefaultLogger(logger Logger) {
	// Set the default logger to the provided logger
	defaultLogger = logger
}

func GetDefaultLogger() Logger {
	if defaultLogger == nil {
		panic("Unrecoverable error because default logger is not set")
	}
	return defaultLogger
}
