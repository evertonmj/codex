// Package native implements the JSON protocol shared by foreign-language bindings.
package native

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	codex "github.com/evertonmj/codex/app"
)

type Request struct {
	Op            string          `json:"op"`
	Handle        string          `json:"handle"`
	File          string          `json:"file"`
	Key           string          `json:"key"`
	Value         json.RawMessage `json:"value,omitempty"`
	EncryptionKey string          `json:"encryptionKey"`
	Ledger        bool            `json:"ledger"`
}
type Response struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
	Code   string      `json:"code,omitempty"`
}

// Engine serializes calls, including close, so no handle is used after release.
// Handles are decimal strings to preserve precision in JavaScript.
type Engine struct {
	mu     sync.Mutex
	next   uint64
	stores map[string]*codex.Store
}

func (e *Engine) Call(input string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var req Request
	var result interface{}
	err := json.Unmarshal([]byte(input), &req)
	if err == nil {
		result, err = e.execute(req)
	}
	response := Response{OK: err == nil, Result: result}
	if err != nil {
		response.Result = nil
		response.Error = err.Error()
		response.Code = "CODEX_ERROR"
		if errors.Is(err, codex.ErrNotFound) {
			response.Code = "NOT_FOUND"
		}
		if errors.Is(err, codex.ErrLocked) {
			response.Code = "LOCKED"
		}
		if errors.Is(err, codex.ErrInvalidKey) {
			response.Code = "INVALID_KEY"
		}
	}
	// Every result comes from validated JSON or primitive values.
	output, _ := json.Marshal(response)
	return string(output)
}
func (e *Engine) execute(req Request) (interface{}, error) {
	if req.Op == "open" {
		if req.File == "" {
			return nil, errors.New("file is required")
		}
		opts := codex.Options{LedgerMode: req.Ledger}
		if req.EncryptionKey != "" {
			opts.EncryptionKey = []byte(req.EncryptionKey)
		}
		store, err := codex.NewWithOptions(req.File, opts)
		if err != nil {
			return nil, err
		}
		if e.stores == nil {
			e.stores = make(map[string]*codex.Store)
		}
		e.next++
		handle := strconv.FormatUint(e.next, 10)
		e.stores[handle] = store
		return handle, nil
	}
	store, ok := e.stores[req.Handle]
	if !ok {
		return nil, errors.New("invalid or closed database handle")
	}
	switch req.Op {
	case "set":
		if len(req.Value) == 0 {
			return nil, errors.New("value is required")
		}
		return true, store.Set(req.Key, req.Value)
	case "get":
		var value json.RawMessage
		err := store.Get(req.Key, &value)
		return value, err
	case "delete":
		return true, store.Delete(req.Key)
	case "has":
		return store.Has(req.Key), nil
	case "keys":
		keys := store.Keys()
		sort.Strings(keys)
		return keys, nil
	case "clear":
		return true, store.Clear()
	case "close":
		delete(e.stores, req.Handle)
		return true, store.Close()
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Op)
	}
}
