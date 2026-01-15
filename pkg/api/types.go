package api

import "encoding/json"

// Request/Response Types

// SetRequest is the request body for PUT /api/v1/keys/{key}
type SetRequest struct {
	Value json.RawMessage `json:"value"`
}

// GetResponse is the response body for GET /api/v1/keys/{key}
type GetResponse struct {
	Value json.RawMessage `json:"value"`
}

// KeysResponse is the response body for GET /api/v1/keys
type KeysResponse struct {
	Keys []string `json:"keys"`
}

// StatusResponse is the response body for successful operations
type StatusResponse struct {
	Status string `json:"status"`
}

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Op    string          `json:"op"`    // "set" or "delete"
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value,omitempty"`
}

// BatchRequest is the request body for POST /api/v1/batch
type BatchRequest struct {
	Operations []BatchOperation `json:"operations"`
}

// BatchGetRequest is the request body for POST /api/v1/batch/get
type BatchGetRequest struct {
	Keys []string `json:"keys"`
}

// BatchGetResponse is the response body for POST /api/v1/batch/get
type BatchGetResponse struct {
	Data map[string]json.RawMessage `json:"data"`
}

// HealthResponse is the response body for GET /health
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadyResponse is the response body for GET /ready
type ReadyResponse struct {
	Status string `json:"status"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error  string `json:"error"`
	Code   string `json:"code"`
	Status int    `json:"status"`
}
