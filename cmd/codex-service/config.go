package main

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds the service configuration
type Config struct {
	// Server
	Port            string
	Host            string
	ShutdownTimeout time.Duration

	// RESP Protocol
	RESPPort string

	// Database
	DBPath           string
	LedgerMode       bool
	NumBackups       int
	Compression      string
	CompressionLevel int

	// Security
	EncryptionKey []byte
	APIKeys       []string

	// Monitoring
	LogLevel string
}

// getDefaultDBPath returns the default database path
// For Kubernetes: /data/codex.db
// For local: ~/.codex_data/codex.db
func getDefaultDBPath() string {
	// Check if running in Kubernetes (CODEX_KUBERNETES env var)
	if getBoolEnv("CODEX_KUBERNETES", false) {
		return "/data/codex.db"
	}

	// For local execution, use home directory
	currentUser, err := user.Current()
	if err != nil {
		// Fallback if we can't get the current user
		homeDir := os.Getenv("HOME")
		if homeDir == "" {
			homeDir = os.TempDir()
		}
		return filepath.Join(homeDir, ".codex_data", "codex.db")
	}

	dbDir := filepath.Join(currentUser.HomeDir, ".codex_data")
	return filepath.Join(dbDir, "codex.db")
}

// ensureDBDirectory creates the database directory if it doesn't exist
func ensureDBDirectory(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

// getAvailablePort returns the first available port from the list of preferred ports
// Tries ports in order: 11111, 922, 1987
func getAvailablePort() string {
	preferredPorts := []string{"11111", "922", "1987"}

	for _, port := range preferredPorts {
		listener, err := net.Listen("tcp", net.JoinHostPort("", port))
		if err == nil {
			listener.Close()
			return port
		}
	}

	// Fallback to 8080 if none of the preferred ports are available
	return "8080"
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Determine the port - use explicit env var if set, otherwise auto-detect
	port := os.Getenv("CODEX_PORT")
	if port == "" {
		port = getAvailablePort()
	}

	config := &Config{
		// Server defaults
		Port:            port,
		Host:            getEnv("CODEX_HOST", "0.0.0.0"),
		ShutdownTimeout: getDurationEnv("CODEX_SHUTDOWN_TIMEOUT", 30*time.Second),
		RESPPort:        getEnv("CODEX_RESP_PORT", "212"),

		// Database defaults
		DBPath:           getEnv("CODEX_DB_PATH", getDefaultDBPath()),
		LedgerMode:       getBoolEnv("CODEX_LEDGER_MODE", false),
		NumBackups:       getIntEnv("CODEX_NUM_BACKUPS", 5),
		Compression:      getEnv("CODEX_COMPRESSION", "none"),
		CompressionLevel: getIntEnv("CODEX_COMPRESSION_LEVEL", 6),

		// Monitoring defaults
		LogLevel: getEnv("CODEX_LOG_LEVEL", "info"),
	}

	// Ensure the database directory exists
	if err := ensureDBDirectory(config.DBPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create database directory: %v\n", err)
	}

	// Load encryption key if provided (hex-encoded)
	if keyHex := os.Getenv("CODEX_ENCRYPTION_KEY"); keyHex != "" {
		config.EncryptionKey = hexDecode(keyHex)
	}

	// Load API keys (comma-separated)
	if keysStr := os.Getenv("CODEX_API_KEYS"); keysStr != "" {
		config.APIKeys = strings.Split(keysStr, ",")
		for i := range config.APIKeys {
			config.APIKeys[i] = strings.TrimSpace(config.APIKeys[i])
		}
	}

	return config
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func hexDecode(s string) []byte {
	// Simple hex decoder - converts "0123456789abcdef" to bytes
	result := make([]byte, 0, len(s)/2)
	for i := 0; i < len(s)-1; i += 2 {
		var b byte
		fmt.Sscanf(s[i:i+2], "%02x", &b)
		result = append(result, b)
	}
	return result
}
