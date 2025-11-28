// Package logger provides structured logging for database operations.
//
// Logging Features:
//   - Structured JSON logging for machine parsing
//   - Multiple log levels: Debug, Info, Warn, Error
//   - File-based logging with append mode
//   - Thread-safe concurrent access
//   - Timestamp and level information
//
// Usage:
//
//	logger := logger.New("my-database.log", logger.LevelInfo)
//	defer logger.Close()
//	logger.Info("Database started", nil)
//
// Logs are appended to the specified file in JSON format for
// easy parsing and analysis.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Level represents the severity level of a log entry.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// Entry represents a single log entry.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	File      string            `json:"file,omitempty"`
	Line      int               `json:"line,omitempty"`
	Function  string            `json:"function,omitempty"`
	Error     string            `json:"error,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// Logger handles structured logging with file persistence.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	level    Level
	filePath string
}

// New creates a new logger that writes to the specified file path.
func New(filePath string, level Level) (*Logger, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:     file,
		level:    level,
		filePath: filePath,
	}, nil
}

// log writes a log entry with the given level and message.
func (l *Logger) log(level Level, msg string, err error, fields map[string]string) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Timestamp: time.Now().UTC(),
		Level:     levelNames[level],
		Message:   msg,
		Fields:    fields,
	}

	// Get caller information
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.File = filepath.Base(file)
		entry.Line = line
		if fn := runtime.FuncForPC(pc); fn != nil {
			entry.Function = fn.Name()
		}
	}

	if err != nil {
		entry.Error = err.Error()
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintf(l.file, "%s\n", data)

	// For fatal errors, also print to stderr
	if level == LevelFatal {
		fmt.Fprintf(os.Stderr, "[FATAL] %s: %s\n", msg, err)
	}
}

// Debug logs a debug-level message.
func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg, nil, nil)
}

// DebugWithFields logs a debug-level message with additional fields.
func (l *Logger) DebugWithFields(msg string, fields map[string]string) {
	l.log(LevelDebug, msg, nil, fields)
}

// Info logs an info-level message.
func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg, nil, nil)
}

// InfoWithFields logs an info-level message with additional fields.
func (l *Logger) InfoWithFields(msg string, fields map[string]string) {
	l.log(LevelInfo, msg, nil, fields)
}

// Warn logs a warning-level message.
func (l *Logger) Warn(msg string) {
	l.log(LevelWarn, msg, nil, nil)
}

// WarnWithError logs a warning-level message with an error.
func (l *Logger) WarnWithError(msg string, err error) {
	l.log(LevelWarn, msg, err, nil)
}

// Error logs an error-level message.
func (l *Logger) Error(msg string, err error) {
	l.log(LevelError, msg, err, nil)
}

// ErrorWithFields logs an error-level message with additional fields.
func (l *Logger) ErrorWithFields(msg string, err error, fields map[string]string) {
	l.log(LevelError, msg, err, fields)
}

// Fatal logs a fatal-level message and does NOT exit the program.
// The caller is responsible for handling fatal errors appropriately.
func (l *Logger) Fatal(msg string, err error) {
	l.log(LevelFatal, msg, err, nil)
}

// Close closes the log file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// SetLevel changes the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// ReadLogs reads all log entries from the log file.
func (l *Logger) ReadLogs() ([]Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	var entries []Entry
	lines := string(data)
	decoder := json.NewDecoder(nil)

	for i := 0; i < len(lines); {
		start := i
		for i < len(lines) && lines[i] != '\n' {
			i++
		}
		if i > start {
			line := lines[start:i]
			var entry Entry
			if err := json.Unmarshal([]byte(line), &entry); err == nil {
				entries = append(entries, entry)
			}
		}
		i++ // skip newline
	}

	_ = decoder // suppress unused warning
	return entries, nil
}
