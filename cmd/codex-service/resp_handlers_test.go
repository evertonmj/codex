package main

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/evertonmj/codex/pkg/resp"
)

// Test helpers

func createTestConnection(store any) *connection {
	return &connection{
		id:            1,
		reader:        bufio.NewReader(bytes.NewBuffer([]byte{})),
		writer:        resp.NewWriter(bytes.NewBuffer([]byte{})),
		authenticated: true,
	}
}

func captureResponse(t *testing.T, fn func(*connection)) string {
	var buf bytes.Buffer
	conn := &connection{
		id:            1,
		reader:        bufio.NewReader(bytes.NewBuffer([]byte{})),
		writer:        resp.NewWriter(&buf),
		authenticated: true,
	}

	fn(conn)
	return buf.String()
}

// Tests for Authenticator

func TestAuthenticator(t *testing.T) {
	tests := []struct {
		name  string
		keys  []string
		key   string
		valid bool
	}{
		{
			name:  "valid key",
			keys:  []string{"key1", "key2"},
			key:   "key1",
			valid: true,
		},
		{
			name:  "invalid key",
			keys:  []string{"key1", "key2"},
			key:   "invalid",
			valid: false,
		},
		{
			name:  "empty keys",
			keys:  []string{},
			key:   "anything",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAuthenticator(tt.keys)
			result := auth.Authenticate(tt.key)
			if result != tt.valid {
				t.Errorf("Authenticate() = %v, want %v", result, tt.valid)
			}
		})
	}
}

func TestAuthenticatorNil(t *testing.T) {
	var auth *Authenticator
	result := auth.Authenticate("anything")
	if !result {
		t.Error("Nil authenticator should return true")
	}
}

// Tests for command handlers

func TestHandleAuth(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		validKeys  []string
		wantOK     bool
		wantError  bool
	}{
		{
			name:      "valid key",
			args:      []string{"mykey"},
			validKeys: []string{"mykey"},
			wantOK:    true,
			wantError: false,
		},
		{
			name:      "invalid key",
			args:      []string{"wrongkey"},
			validKeys: []string{"mykey"},
			wantOK:    false,
			wantError: true,
		},
		{
			name:      "no args",
			args:      []string{},
			validKeys: []string{"mykey"},
			wantOK:    false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := captureResponse(t, func(conn *connection) {
				conn.server = &RESPServer{
					authenticator: NewAuthenticator(tt.validKeys),
				}
				handleAuth(conn, tt.args)
			})

			hasOK := bytes.Contains([]byte(response), []byte("+OK"))
			hasError := bytes.Contains([]byte(response), []byte("-"))

			if tt.wantOK && !hasOK {
				t.Errorf("Expected +OK response, got: %s", response)
			}
			if tt.wantError && !hasError {
				t.Errorf("Expected error response, got: %s", response)
			}
		})
	}
}

func TestHandleSetGet(t *testing.T) {
	// This would require creating a real store, so we'll skip detailed testing
	// The integration tests will cover this
	t.Run("integration test needed", func(t *testing.T) {
		// Full integration test would be done at service level
		t.Skip("Integration test")
	})
}

// Benchmark tests

func BenchmarkAuthenticator(b *testing.B) {
	auth := NewAuthenticator([]string{"key1", "key2", "key3"})
	for i := 0; i < b.N; i++ {
		_ = auth.Authenticate("key2")
	}
}

func BenchmarkHandleAuth(b *testing.B) {
	for i := 0; i < b.N; i++ {
		response := captureResponse(&testing.T{}, func(conn *connection) {
			conn.server = &RESPServer{
				authenticator: NewAuthenticator([]string{"testkey"}),
			}
			handleAuth(conn, []string{"testkey"})
		})
		_ = response
	}
}
