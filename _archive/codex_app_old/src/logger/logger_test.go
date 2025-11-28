package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("creates logger successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "test.log")

		logger, err := New(logPath, LevelInfo)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer logger.Close()

		if logger == nil {
			t.Fatal("expected logger to be created")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "nested", "dir", "test.log")

		logger, err := New(logPath, LevelDebug)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer logger.Close()

		// Check that file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatal("expected log file to exist")
		}
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// Use a path that's definitely invalid on all systems
		logPath := "/\x00/invalid/path/test.log"

		_, err := New(logPath, LevelInfo)
		if err == nil {
			t.Fatal("expected error for invalid path")
		}
	})
}

func TestLogLevels(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		loggerLevel   Level
		logLevel      Level
		shouldLog     bool
		logFunc       func(*Logger)
		expectedLevel string
	}{
		{
			name:          "debug message logged when level is debug",
			loggerLevel:   LevelDebug,
			logLevel:      LevelDebug,
			shouldLog:     true,
			logFunc:       func(l *Logger) { l.Debug("debug message") },
			expectedLevel: "DEBUG",
		},
		{
			name:          "debug message not logged when level is info",
			loggerLevel:   LevelInfo,
			logLevel:      LevelDebug,
			shouldLog:     false,
			logFunc:       func(l *Logger) { l.Debug("debug message") },
			expectedLevel: "DEBUG",
		},
		{
			name:          "info message logged",
			loggerLevel:   LevelInfo,
			logLevel:      LevelInfo,
			shouldLog:     true,
			logFunc:       func(l *Logger) { l.Info("info message") },
			expectedLevel: "INFO",
		},
		{
			name:          "warn message logged",
			loggerLevel:   LevelWarn,
			logLevel:      LevelWarn,
			shouldLog:     true,
			logFunc:       func(l *Logger) { l.Warn("warn message") },
			expectedLevel: "WARN",
		},
		{
			name:          "error message logged",
			loggerLevel:   LevelError,
			logLevel:      LevelError,
			shouldLog:     true,
			logFunc:       func(l *Logger) { l.Error("error message", nil) },
			expectedLevel: "ERROR",
		},
		{
			name:          "fatal message logged",
			loggerLevel:   LevelFatal,
			logLevel:      LevelFatal,
			shouldLog:     true,
			logFunc:       func(l *Logger) { l.Fatal("fatal message", nil) },
			expectedLevel: "FATAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := filepath.Join(tmpDir, tt.name+".log")
			logger, err := New(testPath, tt.loggerLevel)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}
			defer logger.Close()

			tt.logFunc(logger)
			logger.Close()

			data, err := os.ReadFile(testPath)
			if err != nil {
				t.Fatalf("failed to read log file: %v", err)
			}

			content := string(data)
			if tt.shouldLog {
				if content == "" {
					t.Fatal("expected log content, got empty file")
				}
				if !strings.Contains(content, tt.expectedLevel) {
					t.Errorf("expected log level %s in content, got: %s", tt.expectedLevel, content)
				}
			} else {
				if content != "" {
					t.Errorf("expected no log content, got: %s", content)
				}
			}
		})
	}
}

func TestLogWithFields(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelDebug)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	fields := map[string]string{
		"user_id": "12345",
		"action":  "login",
	}

	logger.DebugWithFields("user action", fields)
	logger.InfoWithFields("info with fields", fields)
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "user_id") {
		t.Error("expected fields in log content")
	}
	if !strings.Contains(content, "12345") {
		t.Error("expected field value in log content")
	}
}

func TestLogWithError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelDebug)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	testErr := os.ErrNotExist
	logger.Error("operation failed", testErr)
	logger.WarnWithError("warning with error", testErr)

	fields := map[string]string{"operation": "read"}
	logger.ErrorWithFields("error with fields", testErr, fields)
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "file does not exist") {
		t.Error("expected error message in log content")
	}
	if !strings.Contains(content, "ERROR") {
		t.Error("expected ERROR level in log content")
	}
}

func TestSetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelInfo)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger.GetLevel() != LevelInfo {
		t.Errorf("expected level %v, got %v", LevelInfo, logger.GetLevel())
	}

	logger.SetLevel(LevelDebug)
	if logger.GetLevel() != LevelDebug {
		t.Errorf("expected level %v, got %v", LevelDebug, logger.GetLevel())
	}

	// Now debug messages should be logged
	logger.Debug("debug message")
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(data), "debug message") {
		t.Error("expected debug message after level change")
	}
}

func TestReadLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelInfo)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("first message")
	logger.Warn("second message")
	logger.Error("third message", os.ErrPermission)

	entries, err := logger.ReadLogs()
	if err != nil {
		t.Fatalf("failed to read logs: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Verify first entry
	if entries[0].Message != "first message" {
		t.Errorf("expected 'first message', got '%s'", entries[0].Message)
	}
	if entries[0].Level != "INFO" {
		t.Errorf("expected 'INFO', got '%s'", entries[0].Level)
	}

	// Verify third entry has error
	if entries[2].Error == "" {
		t.Error("expected error in third entry")
	}

	logger.Close()
}

func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelInfo)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer logger.Close()

	done := make(chan bool)
	numGoroutines := 10
	messagesPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("concurrent message")
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	expectedLines := numGoroutines * messagesPerGoroutine

	if len(lines) != expectedLines {
		t.Errorf("expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestLogEntryStructure(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelInfo)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("test message")
	logger.Close()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	// Verify all required fields are present
	if entry.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}
	if entry.Level == "" {
		t.Error("expected level to be set")
	}
	if entry.Message == "" {
		t.Error("expected message to be set")
	}
	if entry.File == "" {
		t.Error("expected file to be set")
	}
	if entry.Line == 0 {
		t.Error("expected line to be set")
	}
	if entry.Function == "" {
		t.Error("expected function to be set")
	}
}

func TestReadLogsFromNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "nonexistent.log")

	logger := &Logger{filePath: logPath, level: LevelInfo}
	_, err := logger.ReadLogs()
	if err == nil {
		t.Fatal("expected error reading non-existent log file")
	}
}

func TestCloseIdempotency(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger, err := New(logPath, LevelInfo)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// First close should succeed
	if err := logger.Close(); err != nil {
		t.Errorf("first close failed: %v", err)
	}

	// Second close should return an error
	if err := logger.Close(); err == nil {
		t.Error("expected error on second close")
	}
}
