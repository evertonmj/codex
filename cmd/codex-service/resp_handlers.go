package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/evertonmj/codex/pkg/resp"
)

// Authenticator manages API key authentication
type Authenticator struct {
	keys map[string]bool
	mu   sync.RWMutex
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(keys []string) *Authenticator {
	auth := &Authenticator{
		keys: make(map[string]bool),
	}
	for _, key := range keys {
		auth.keys[key] = true
	}
	return auth
}

// Authenticate checks if the key is valid
func (a *Authenticator) Authenticate(key string) bool {
	if a == nil {
		return true
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.keys[key]
}

// Handler functions

// handleSet handles CDX.SET command
func handleSet(conn *connection, args []string) {
	if len(args) < 2 {
		conn.writeError("ERR wrong number of arguments for 'CDX.SET' command")
		return
	}

	key := args[0]
	valueStr := strings.Join(args[1:], " ")

	// Try to parse as JSON
	var value interface{}
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		// If not JSON, treat as plain string
		value = valueStr
	}

	// Set in store
	if err := conn.server.store.Set(key, value); err != nil {
		log.Printf("[conn:%d] SET error: %v", conn.id, err)
		conn.writeError(fmt.Sprintf("ERR %v", err))
		return
	}

	conn.writeSimpleString("OK")
}

// handleGet handles CDX.GET command
func handleGet(conn *connection, args []string) {
	if len(args) != 1 {
		conn.writeError("ERR wrong number of arguments for 'CDX.GET' command")
		return
	}

	key := args[0]
	var value interface{}

	if err := conn.server.store.Get(key, &value); err != nil {
		if err.Error() == "key not found" {
			conn.writeNull()
			return
		}
		log.Printf("[conn:%d] GET error: %v", conn.id, err)
		conn.writeError(fmt.Sprintf("ERR %v", err))
		return
	}

	// Convert value to JSON bytes for RESP
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		log.Printf("[conn:%d] GET marshal error: %v", conn.id, err)
		conn.writeError(fmt.Sprintf("ERR failed to marshal value: %v", err))
		return
	}

	conn.writeBulkString(jsonBytes)
}

// handleDelete handles CDX.DELETE command
func handleDelete(conn *connection, args []string) {
	if len(args) != 1 {
		conn.writeError("ERR wrong number of arguments for 'CDX.DELETE' command")
		return
	}

	key := args[0]
	existed := conn.server.store.Has(key)

	if err := conn.server.store.Delete(key); err != nil {
		log.Printf("[conn:%d] DELETE error: %v", conn.id, err)
		conn.writeError(fmt.Sprintf("ERR %v", err))
		return
	}

	// Return 1 if key existed, 0 if not
	if existed {
		conn.writeInteger(1)
	} else {
		conn.writeInteger(0)
	}
}

// handleHas handles CDX.HAS command
func handleHas(conn *connection, args []string) {
	if len(args) != 1 {
		conn.writeError("ERR wrong number of arguments for 'CDX.HAS' command")
		return
	}

	key := args[0]
	exists := conn.server.store.Has(key)

	if exists {
		conn.writeInteger(1)
	} else {
		conn.writeInteger(0)
	}
}

// handleKeys handles CDX.KEYS command
func handleKeys(conn *connection, args []string) {
	// CDX.KEYS [pattern] - pattern matching not implemented yet, just returns all keys
	keys := conn.server.store.Keys()

	// Create array of bulk strings
	values := make([]resp.Value, len(keys))
	for i, key := range keys {
		values[i] = resp.Value{
			Type: resp.BulkString,
			Bulk: []byte(key),
		}
	}

	if err := conn.writeArray(values); err != nil {
		log.Printf("[conn:%d] KEYS write error: %v", conn.id, err)
	}
}

// handleClear handles CDX.CLEAR command
func handleClear(conn *connection, args []string) {
	if len(args) != 0 {
		conn.writeError("ERR wrong number of arguments for 'CDX.CLEAR' command")
		return
	}

	if err := conn.server.store.Clear(); err != nil {
		log.Printf("[conn:%d] CLEAR error: %v", conn.id, err)
		conn.writeError(fmt.Sprintf("ERR %v", err))
		return
	}

	conn.writeSimpleString("OK")
}

// handlePing handles CDX.PING command
func handlePing(conn *connection, args []string) {
	if len(args) == 0 {
		conn.writeSimpleString("PONG")
	} else {
		// Echo the message back
		msg := strings.Join(args, " ")
		conn.writeBulkStringStr(msg)
	}
}

// handleInfo handles CDX.INFO command
func handleInfo(conn *connection, args []string) {
	if len(args) > 1 {
		conn.writeError("ERR wrong number of arguments for 'CDX.INFO' command")
		return
	}

	// Gather server info
	keys := conn.server.store.Keys()
	info := fmt.Sprintf(
		"# CodexDB RESP Server\r\ncodex_version: 1.0.0\r\nstorage_engine: file\r\ndatabase_path: %s\r\nnum_keys: %d\r\nuptime_seconds: %d\r\n",
		conn.server.store.Path(),
		len(keys),
		int(time.Since(startTime).Seconds()),
	)

	conn.writeBulkStringStr(info)
}

// handleAuth handles CDX.AUTH command
func handleAuth(conn *connection, args []string) {
	if len(args) != 1 {
		conn.writeError("ERR wrong number of arguments for 'CDX.AUTH' command")
		return
	}

	apiKey := args[0]

	// Check if key is valid
	if !conn.server.authenticator.Authenticate(apiKey) {
		conn.writeError("NOAUTH invalid API key")
		return
	}

	conn.authenticated = true
	log.Printf("[conn:%d] Authentication successful", conn.id)
	conn.writeSimpleString("OK")
}
