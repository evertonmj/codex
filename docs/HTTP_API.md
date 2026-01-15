# CodexDB HTTP Service API Documentation

## Overview

CodexDB HTTP Service exposes all key-value database operations through a RESTful HTTP API. This service is designed to run in Kubernetes and provides:

- **RESTful endpoints** for all database operations
- **API key authentication** for security
- **Health checks** for Kubernetes orchestration
- **JSON request/response bodies** for easy integration
- **Proper HTTP status codes** for error handling

**Note:** For high-performance scenarios, CodexDB also provides a [RESP protocol server](./RESP_PROTOCOL.md) running on port 212 with 3-5x better throughput.

## Base URL

```
http://localhost:<port>/api/v1/
```

**Default Port:** The service auto-detects the best available port in this order:
1. **Port 11111** (preferred)
2. **Port 922** (if 11111 is unavailable)
3. **Port 1987** (if 922 is unavailable)
4. **Port 8080** (fallback)

You can override with the `CODEX_PORT` environment variable:
```bash
CODEX_PORT=9000 ./bin/codex-service
```

## Authentication

All endpoints (except `/health` and `/ready`) require API key authentication.

Include the API key in the request header:

```
X-API-Key: your-api-key-here
```

### Example

```bash
curl -X GET http://localhost:1111111/api/v1/keys \
  -H "X-API-Key: your-api-key-here"
```

**Note:** Replace `11111` with your actual port (see Base URL section above)

## Core Operations

### Set a Key-Value Pair

**Endpoint:** `PUT /api/v1/keys/{key}`

**Authentication:** Required

**Description:** Store a value for a key. If the key exists, it will be overwritten.

**Request Body:**
```json
{
  "value": <any-json-value>
}
```

**Response (Success):** 200 OK
```json
{
  "status": "ok"
}
```

**Response (Error):** 400/401/403/500
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

**Examples:**

Store a string:
```bash
curl -X PUT http://localhost:11111/api/v1/keys/greeting \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{"value": "hello world"}'
```

Store a JSON object:
```bash
curl -X PUT http://localhost:11111/api/v1/keys/user:1 \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{"value": {"name": "Alice", "age": 30}}'
```

Store a number:
```bash
curl -X PUT http://localhost:11111/api/v1/keys/counter \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{"value": 42}'
```

---

### Get a Value

**Endpoint:** `GET /api/v1/keys/{key}`

**Authentication:** Required

**Description:** Retrieve the value stored for a key.

**Response (Success):** 200 OK
```json
{
  "value": <the-stored-value>
}
```

**Response (Key Not Found):** 404 NOT_FOUND
```json
{
  "error": "Key not found",
  "code": "NOT_FOUND",
  "status": 404
}
```

**Examples:**

```bash
# Get a string value
curl http://localhost:8080/api/v1/keys/greeting \
  -H "X-API-Key: test-key"

# Response:
# {"value":"hello world"}

# Get an object value
curl http://localhost:8080/api/v1/keys/user:1 \
  -H "X-API-Key: test-key"

# Response:
# {"value":{"name":"Alice","age":30}}
```

---

### Delete a Key

**Endpoint:** `DELETE /api/v1/keys/{key}`

**Authentication:** Required

**Description:** Remove a key and its value from the database.

**Response (Success):** 200 OK
```json
{
  "status": "ok"
}
```

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/keys/greeting \
  -H "X-API-Key: test-key"
```

---

### Check if Key Exists

**Endpoint:** `HEAD /api/v1/keys/{key}`

**Authentication:** Required

**Description:** Check if a key exists without retrieving its value.

**Response (Key Exists):** 200 OK
(no body)

**Response (Key Not Found):** 404 NOT_FOUND
(no body)

**Example:**

```bash
curl -I http://localhost:8080/api/v1/keys/greeting \
  -H "X-API-Key: test-key"
```

---

### List All Keys

**Endpoint:** `GET /api/v1/keys`

**Authentication:** Required

**Description:** Retrieve all keys in the database.

**Response (Success):** 200 OK
```json
{
  "keys": ["key1", "key2", "key3"]
}
```

**Response (Empty Database):** 200 OK
```json
{
  "keys": []
}
```

**Example:**

```bash
curl http://localhost:8080/api/v1/keys \
  -H "X-API-Key: test-key"
```

---

### Clear All Keys

**Endpoint:** `DELETE /api/v1/keys`

**Authentication:** Required

**Description:** Remove all keys and values from the database.

**Response (Success):** 200 OK
```json
{
  "status": "ok"
}
```

**WARNING:** This operation cannot be undone!

**Example:**

```bash
curl -X DELETE http://localhost:8080/api/v1/keys \
  -H "X-API-Key: test-key"
```

---

## Batch Operations

### Batch Set/Delete

**Endpoint:** `POST /api/v1/batch`

**Authentication:** Required

**Description:** Perform multiple set/delete operations atomically.

**Request Body:**
```json
{
  "operations": [
    {"op": "set", "key": "key1", "value": "value1"},
    {"op": "set", "key": "key2", "value": {"nested": "object"}},
    {"op": "delete", "key": "old-key"}
  ]
}
```

**Response (Success):** 200 OK
```json
{
  "status": "ok"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/batch \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "operations": [
      {"op": "set", "key": "user:1", "value": {"name": "Alice"}},
      {"op": "set", "key": "user:2", "value": {"name": "Bob"}},
      {"op": "delete", "key": "temp:old"}
    ]
  }'
```

---

### Batch Get

**Endpoint:** `POST /api/v1/batch/get`

**Authentication:** Required

**Description:** Retrieve multiple values in a single request.

**Request Body:**
```json
{
  "keys": ["key1", "key2", "key3"]
}
```

**Response (Success):** 200 OK
```json
{
  "data": {
    "key1": "value1",
    "key2": {"nested": "value"},
    "key3": 42
  }
}
```

**Note:** Missing keys are simply omitted from the response.

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/batch/get \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "keys": ["user:1", "user:2", "config"]
  }'
```

---

## Health & Status Endpoints

### Health Check

**Endpoint:** `GET /health`

**Authentication:** Not Required

**Description:** Simple health check endpoint for load balancers and orchestrators.

**Response:** 200 OK
```json
{
  "status": "healthy"
}
```

**Example:**

```bash
curl http://localhost:8080/health
```

---

### Readiness Check

**Endpoint:** `GET /ready`

**Authentication:** Not Required

**Description:** Readiness probe endpoint. Returns success when the service is ready to handle requests.

**Response:** 200 OK
```json
{
  "status": "ready"
}
```

**Example:**

```bash
curl http://localhost:8080/ready
```

---

## Error Handling

The service returns standard HTTP status codes and error responses:

### Error Response Format

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "status": 400
}
```

### Status Code Reference

| HTTP Status | Code | Meaning |
|-------------|------|---------|
| 200 | OK | Request successful |
| 400 | BAD_REQUEST | Invalid request (malformed JSON, missing fields) |
| 401 | UNAUTHORIZED | Missing API key header |
| 403 | FORBIDDEN | Invalid API key |
| 404 | NOT_FOUND | Key not found |
| 500 | INTERNAL_ERROR | Server error |
| 500 | CORRUPTION | Data integrity check failed |
| 500 | INVALID_KEY | Configuration error (invalid encryption key) |
| 503 | SERVICE_UNAVAILABLE | Database locked or unavailable |

### Common Errors

**Missing API Key:**
```json
{
  "error": "Missing X-API-Key header",
  "code": "UNAUTHORIZED",
  "status": 401
}
```

**Invalid API Key:**
```json
{
  "error": "Invalid API key",
  "code": "FORBIDDEN",
  "status": 403
}
```

**Invalid JSON:**
```json
{
  "error": "Invalid request body",
  "code": "BAD_REQUEST",
  "status": 400
}
```

**Key Not Found:**
```json
{
  "error": "Key not found",
  "code": "NOT_FOUND",
  "status": 404
}
```

---

## Environment Variables

Configure the service using environment variables:

### Server Configuration
- `CODEX_PORT` - Service port (default: 8080)
- `CODEX_HOST` - Bind address (default: 0.0.0.0)
- `CODEX_SHUTDOWN_TIMEOUT` - Graceful shutdown timeout (default: 30s)

### Database Configuration
- `CODEX_DB_PATH` - Path to database file (default: /data/codex.db)
- `CODEX_LEDGER_MODE` - Enable ledger mode for audit trail (default: false)
- `CODEX_NUM_BACKUPS` - Number of rotating backups (default: 5)

### Compression
- `CODEX_COMPRESSION` - Compression algorithm: none, gzip, zstd, snappy (default: none)
- `CODEX_COMPRESSION_LEVEL` - Compression level 1-9 (default: 6)

### Security
- `CODEX_API_KEYS` - Comma-separated API keys (required for authentication)
- `CODEX_ENCRYPTION_KEY` - AES encryption key as hex string (32 bytes for AES-256)

### Monitoring
- `CODEX_LOG_LEVEL` - Log level: debug, info, warn, error (default: info)

---

## Examples

### Node.js/JavaScript

```javascript
const apiKey = 'test-key';
const baseUrl = 'http://localhost:8080/api/v1';

// Set a value
const response = await fetch(`${baseUrl}/keys/mykey`, {
  method: 'PUT',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': apiKey,
  },
  body: JSON.stringify({ value: 'my-value' }),
});

// Get a value
const getResponse = await fetch(`${baseUrl}/keys/mykey`, {
  headers: { 'X-API-Key': apiKey },
});
const data = await getResponse.json();
console.log(data.value);
```

### Python

```python
import requests

api_key = 'test-key'
base_url = 'http://localhost:8080/api/v1'
headers = {'X-API-Key': api_key}

# Set a value
response = requests.put(
    f'{base_url}/keys/mykey',
    json={'value': 'my-value'},
    headers=headers
)
print(response.json())

# Get a value
response = requests.get(f'{base_url}/keys/mykey', headers=headers)
print(response.json()['value'])
```

### Go

```go
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	apiKey  = "test-key"
	baseUrl = "http://localhost:8080/api/v1"
)

func setKey(key string, value interface{}) error {
	payload, _ := json.Marshal(map[string]interface{}{"value": value})

	req, _ := http.NewRequest("PUT", baseUrl+"/keys/"+key, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
```

---

## Rate Limiting & Performance

The HTTP service has no built-in rate limiting. Use Kubernetes NetworkPolicy or an Ingress controller to enforce rate limits if needed.

### Performance Tips

1. **Use batch operations** for multiple keys (10-50x faster)
2. **Enable compression** for large values (reduce network bandwidth)
3. **Use connection pooling** in client libraries
4. **Set appropriate timeouts** in client code

---

## Troubleshooting

### Service won't start

Check the logs:
```bash
kubectl logs <pod-name>
```

Verify configuration:
```bash
kubectl describe configmap codex-config
kubectl describe secret codex-secrets
```

### Cannot connect to service

Verify the pod is running:
```bash
kubectl get pods
kubectl describe pod <pod-name>
```

Test connectivity:
```bash
kubectl port-forward <pod-name> 8080:8080
curl http://localhost:8080/health
```

### Authentication errors

Verify the API key is set:
```bash
kubectl get secret codex-secrets -o jsonpath='{.data.CODEX_API_KEYS}' | base64 -d
```

Ensure the header name is exactly `X-API-Key` (case-sensitive).

---

## API Versioning

The current API version is `/api/v1/`. Future breaking changes will use `/api/v2/`, etc.

Clients should be prepared to handle new optional fields in responses.

---

## Rate Limits & Quotas

Currently no limits are enforced at the HTTP layer. Implement through Kubernetes mechanisms:

- **NetworkPolicy** - Limit traffic between pods
- **ResourceQuota** - Limit pod resource usage
- **Ingress Controllers** - Add rate limiting at the edge
- **Service Mesh** - Advanced traffic policies (Istio, Linkerd)
