package logger

import "testing"

func TestLoggerOpenFailure(t *testing.T) {
	if _, err := New(t.TempDir(), LevelInfo); err == nil {
		t.Fatal("directory accepted as log file")
	}
}
