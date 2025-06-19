package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger is a simple structured logger
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

// Global logger instance
var defaultLogger = &Logger{
	level:  InfoLevel,
	logger: log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile),
}

// SetLevel sets the global log level
func SetLevel(levelStr string) {
	level := InfoLevel
	switch strings.ToLower(levelStr) {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn", "warning":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	}
	defaultLogger.level = level
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	defaultLogger.log(DebugLevel, format, args...)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	defaultLogger.log(InfoLevel, format, args...)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	defaultLogger.log(WarnLevel, format, args...)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	defaultLogger.log(ErrorLevel, format, args...)
}

// WithField returns a logger with a field (simplified version)
func WithField(key string, value interface{}) *Logger {
	return defaultLogger // Simplified for now
}

// WithError returns a logger with an error field
func WithError(err error) *Logger {
	return defaultLogger // Simplified for now
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	levelStr := ""
	switch level {
	case DebugLevel:
		levelStr = "[DEBUG]"
	case InfoLevel:
		levelStr = "[INFO] "
	case WarnLevel:
		levelStr = "[WARN] "
	case ErrorLevel:
		levelStr = "[ERROR]"
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("%s %s", levelStr, msg)
}

// Implement methods on Logger for chaining
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}