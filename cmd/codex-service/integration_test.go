package main

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"testing"
	"time"

	codex "github.com/evertonmj/codex/app"
	"github.com/evertonmj/codex/pkg/resp"
)

// TestRESPServerIntegration tests the RESP server with actual connections
func TestRESPServerIntegration(t *testing.T) {
	// Create a temporary database for testing
	tmpFile := "/tmp/codex_test_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(tmpFile)

	// Create store
	store, err := codex.NewWithOptions(tmpFile, codex.Options{})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create config
	config := &Config{
		Host:            "127.0.0.1",
		Port:            "11111",
		RESPPort:        "2127", // Use non-standard port for testing
		ShutdownTimeout: 5 * time.Second,
		APIKeys:         []string{},
	}

	// Create server
	auth := NewAuthenticator(config.APIKeys)
	server := NewRESPServer(store, config, auth)

	// Start server
	err = server.Start()
	if err != nil {
		t.Fatalf("Failed to start RESP server: %v", err)
	}

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Test PING command
	t.Run("PING", func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:2127")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// Send PING command
		_, err = conn.Write([]byte("*1\r\n$8\r\nCDX.PING\r\n"))
		if err != nil {
			t.Fatalf("Failed to write: %v", err)
		}

		// Read response
		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}

		if response != "+PONG\r\n" {
			t.Errorf("Expected +PONG\\r\\n, got: %q", response)
		}
	})

	// Test SET and GET
	t.Run("SET and GET", func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:2127")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// SET command
		_, err = conn.Write([]byte("*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"))
		if err != nil {
			t.Fatalf("Failed to write SET: %v", err)
		}

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read SET response: %v", err)
		}

		if response != "+OK\r\n" {
			t.Errorf("Expected +OK\\r\\n, got: %q", response)
		}

		// GET command
		_, err = conn.Write([]byte("*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n"))
		if err != nil {
			t.Fatalf("Failed to write GET: %v", err)
		}

		// Read bulk string header
		header, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read bulk header: %v", err)
		}

		// Value is JSON encoded, so it's "value" (7 bytes including quotes)
		if header != "$7\r\n" {
			t.Errorf("Expected bulk string header '$7\\r\\n', got: %q", header)
		}

		// Read bulk data (7 bytes for JSON encoded string)
		data := make([]byte, 7)
		_, err = reader.Read(data)
		if err != nil {
			t.Fatalf("Failed to read bulk data: %v", err)
		}

		if string(data) != "\"value\"" {
			t.Errorf("Expected JSON encoded value, got: %q", string(data))
		}
	})

	// Test HAS command
	t.Run("HAS", func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:2127")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// HAS existing key
		_, err = conn.Write([]byte("*2\r\n$7\r\nCDX.HAS\r\n$3\r\nkey\r\n"))
		if err != nil {
			t.Fatalf("Failed to write HAS: %v", err)
		}

		reader := bufio.NewReader(conn)
		response, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read HAS response: %v", err)
		}

		if response != ":1\r\n" {
			t.Errorf("Expected :1\\r\\n, got: %q", response)
		}

		// HAS non-existing key
		_, err = conn.Write([]byte("*2\r\n$7\r\nCDX.HAS\r\n$7\r\nnoexist\r\n"))
		if err != nil {
			t.Fatalf("Failed to write HAS: %v", err)
		}

		response, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read HAS response: %v", err)
		}

		if response != ":0\r\n" {
			t.Errorf("Expected :0\\r\\n, got: %q", response)
		}
	})

	// Test KEYS command
	t.Run("KEYS", func(t *testing.T) {
		conn, err := net.Dial("tcp", "127.0.0.1:2127")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		// KEYS command
		_, err = conn.Write([]byte("*1\r\n$8\r\nCDX.KEYS\r\n"))
		if err != nil {
			t.Fatalf("Failed to write KEYS: %v", err)
		}

		reader := bufio.NewReader(conn)
		// Should return array
		header, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read KEYS response: %v", err)
		}

		// Should be an array header like "*1\r\n"
		if !bytes.Contains([]byte(header), []byte("*")) {
			t.Errorf("Expected array header, got: %q", header)
		}
	})

	// Shutdown server
	err = server.Shutdown()
	if err != nil {
		t.Errorf("Failed to shutdown server: %v", err)
	}
}

// TestRESPParser tests the RESP parser with real protocol messages
func TestRESPParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantCmd string
		wantLen int
	}{
		{
			name:    "SET command",
			input:   "*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			wantCmd: "CDX.SET",
			wantLen: 3,
		},
		{
			name:    "GET command",
			input:   "*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n",
			wantCmd: "CDX.GET",
			wantLen: 2,
		},
		{
			name:    "PING command",
			input:   "*1\r\n$8\r\nCDX.PING\r\n",
			wantCmd: "CDX.PING",
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewBufferString(tt.input))
			parser := resp.NewParser(reader)
			val, err := parser.Parse()

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			cmd, args, err := resp.ParseCommand(val)
			if err != nil {
				t.Fatalf("ParseCommand() error = %v", err)
			}

			if cmd != tt.wantCmd {
				t.Errorf("Expected cmd %q, got %q", tt.wantCmd, cmd)
			}

			// args count is total elements minus 1 (command name)
			if len(args) != tt.wantLen-1 {
				t.Errorf("Expected %d args, got %d", tt.wantLen-1, len(args))
			}
		})
	}
}
