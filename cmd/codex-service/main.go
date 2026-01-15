package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	codex "github.com/evertonmj/codex/app"
	"github.com/evertonmj/codex/pkg/api"
)

func main() {
	// Load configuration from environment
	config := LoadConfig()

	// Create logger
	logger := log.New(os.Stdout, "[codex-service] ", log.LstdFlags|log.Lshortfile)

	logger.Printf("Starting CodexDB HTTP Service")
	logger.Printf("Server: %s:%s", config.Host, config.Port)
	logger.Printf("Database: %s (ledger=%v)", config.DBPath, config.LedgerMode)

	// Initialize Store
	storeOptions := codex.Options{
		LedgerMode:       config.LedgerMode,
		NumBackups:       config.NumBackups,
		CompressionLevel: config.CompressionLevel,
		EncryptionKey:    config.EncryptionKey,
	}

	// Set compression algorithm
	switch config.Compression {
	case "gzip":
		storeOptions.Compression = codex.GzipCompression
	case "zstd":
		storeOptions.Compression = codex.ZstdCompression
	case "snappy":
		storeOptions.Compression = codex.SnappyCompression
	default:
		storeOptions.Compression = codex.NoCompression
	}

	store, err := codex.NewWithOptions(config.DBPath, storeOptions)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	logger.Printf("Database initialized successfully")

	// Create service
	service := &Service{
		store:   store,
		config:  config,
		logger:  logger,
		apiKeys: make(map[string]struct{}),
	}

	// Hash API keys
	for _, key := range config.APIKeys {
		service.apiKeys[key] = struct{}{}
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Health/readiness endpoints (no auth required)
	mux.HandleFunc("/health", service.HealthHandler)
	mux.HandleFunc("/ready", service.ReadyHandler)

	// API endpoints (with auth middleware)
	mux.HandleFunc("/api/v1/keys", service.withAuth(service.KeysHandler))
	mux.HandleFunc("/api/v1/keys/", service.withAuth(service.KeyValueHandler))

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler: loggingMiddleware(logger, mux),
	}

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start server in a goroutine
	go func() {
		logger.Printf("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	logger.Printf("Shutdown signal received, gracefully stopping server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server shutdown error: %v", err)
	}

	logger.Printf("Server stopped successfully")
}

// Service represents the HTTP service
type Service struct {
	store   *codex.Store
	config  *Config
	logger  *log.Logger
	apiKeys map[string]struct{}
}

// withAuth wraps a handler with API key authentication
func (s *Service) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			sendError(w, api.Unauthorized("Missing X-API-Key header"))
			return
		}

		if _, exists := s.apiKeys[apiKey]; !exists && len(s.apiKeys) > 0 {
			sendError(w, api.Forbidden("Invalid API key"))
			return
		}

		next(w, r)
	}
}

// KeyValueHandler handles /api/v1/keys/{key} requests
func (s *Service) KeyValueHandler(w http.ResponseWriter, r *http.Request) {
	// Extract key from URL path
	key := r.URL.Path[len("/api/v1/keys/"):]
	if key == "" {
		sendError(w, api.BadRequest("Key is required"))
		return
	}

	switch r.Method {
	case http.MethodPut:
		s.handleSet(w, r, key)
	case http.MethodGet:
		s.handleGet(w, r, key)
	case http.MethodDelete:
		s.handleDelete(w, r, key)
	case http.MethodHead:
		s.handleHas(w, r, key)
	default:
		sendError(w, api.BadRequest(fmt.Sprintf("Method %s not allowed", r.Method)))
	}
}

// KeysHandler handles /api/v1/keys requests
func (s *Service) KeysHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListKeys(w, r)
	case http.MethodPost:
		s.handleBatch(w, r)
	case http.MethodDelete:
		s.handleClear(w, r)
	default:
		sendError(w, api.BadRequest(fmt.Sprintf("Method %s not allowed", r.Method)))
	}
}

// handleSet stores a key-value pair
func (s *Service) handleSet(w http.ResponseWriter, r *http.Request, key string) {
	var req api.SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, api.BadRequest("Invalid request body"))
		return
	}

	// Unmarshal the value to get the actual value
	var value interface{}
	if err := json.Unmarshal(req.Value, &value); err != nil {
		sendError(w, api.BadRequest("Invalid JSON value"))
		return
	}

	if err := s.store.Set(key, value); err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	sendJSON(w, http.StatusOK, api.StatusResponse{Status: "ok"})
}

// handleGet retrieves a value by key
func (s *Service) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	var value interface{}
	if err := s.store.Get(key, &value); err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	// Marshal value back to JSON
	data, err := json.Marshal(value)
	if err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	response := api.GetResponse{Value: data}
	sendJSON(w, http.StatusOK, response)
}

// handleDelete removes a key
func (s *Service) handleDelete(w http.ResponseWriter, r *http.Request, key string) {
	if err := s.store.Delete(key); err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	sendJSON(w, http.StatusOK, api.StatusResponse{Status: "ok"})
}

// handleHas checks if a key exists
func (s *Service) handleHas(w http.ResponseWriter, r *http.Request, key string) {
	if s.store.Has(key) {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

// handleListKeys returns all keys
func (s *Service) handleListKeys(w http.ResponseWriter, r *http.Request) {
	keys := s.store.Keys()
	if keys == nil {
		keys = []string{}
	}
	sendJSON(w, http.StatusOK, api.KeysResponse{Keys: keys})
}

// handleClear removes all keys
func (s *Service) handleClear(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Clear(); err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	sendJSON(w, http.StatusOK, api.StatusResponse{Status: "ok"})
}

// handleBatch processes batch operations
func (s *Service) handleBatch(w http.ResponseWriter, r *http.Request) {
	var req api.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, api.BadRequest("Invalid request body"))
		return
	}

	batch := s.store.NewBatch()
	for _, op := range req.Operations {
		switch op.Op {
		case "set":
			var value interface{}
			if err := json.Unmarshal(op.Value, &value); err != nil {
				sendError(w, api.BadRequest("Invalid value in batch operation"))
				return
			}
			batch.Set(op.Key, value)
		case "delete":
			batch.Delete(op.Key)
		default:
			sendError(w, api.BadRequest(fmt.Sprintf("Unknown operation: %s", op.Op)))
			return
		}
	}

	if err := batch.Execute(); err != nil {
		sendError(w, api.ErrorToHTTP(err))
		return
	}

	sendJSON(w, http.StatusOK, api.StatusResponse{Status: "ok"})
}

// HealthHandler returns the health status
func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, api.HealthResponse{Status: "healthy"})
}

// ReadyHandler returns the readiness status
func (s *Service) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, api.ReadyResponse{Status: "ready"})
}

// Helper functions

func sendJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, httpErr api.HTTPError) {
	sendJSON(w, httpErr.StatusCode, api.ErrorResponse{
		Error:  httpErr.Message,
		Code:   httpErr.Code,
		Status: httpErr.StatusCode,
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		logger.Printf("%s %s %s (%.2fms)", r.Method, r.RequestURI, r.RemoteAddr, float64(duration.Microseconds())/1000)
	})
}
