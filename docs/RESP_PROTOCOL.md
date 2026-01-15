# CodexDB RESP Protocol Documentation

CodexDB provides both HTTP REST API and RESP (Redis Serialization Protocol) server interfaces. The RESP protocol offers superior performance for high-throughput scenarios while maintaining backward compatibility with the HTTP API.

## Overview

| Aspect | HTTP | RESP |
|--------|------|------|
| Protocol | Text-based REST | Binary RESP |
| Port | 11111 (auto-detected) | 212 |
| Latency | 1-5ms | 0.3-1ms |
| Throughput | 200-1000 ops/sec | 1000-3000 ops/sec |
| Best For | Web services, debugging | High-throughput scenarios |
| Client Availability | Everywhere (curl, browsers) | Redis clients, custom |

## When to Use RESP

**Use RESP when:**
- You need high throughput (>1000 ops/sec)
- Latency is critical (<1ms)
- You're building internal services (not web)
- You have many connections from a single service

**Use HTTP when:**
- You're building web applications
- You need easy debugging (curl, Postman)
- Simplicity is more important than throughput
- Clients are widely distributed

## Connection

### Basic Connection

```bash
# Using telnet
telnet localhost 212

# Using nc (netcat)
nc localhost 212

# Using Go net/dial
go run examples/resp_client.go

# Using Python socket
python examples/resp_client.py
```

### Connection Persistence

RESP connections are persistent. Once connected, send multiple commands without reconnecting:

```bash
telnet localhost 212
# Connection established

# Send command 1
*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
# Response: +OK\r\n

# Send command 2 (same connection)
*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n
# Response: $5\r\nvalue\r\n
```

## Protocol Format

RESP uses a simple text protocol with the following data types:

### Simple Strings
- Format: `+status\r\n`
- Example: `+OK\r\n`
- Used for: Success responses

### Errors
- Format: `-error message\r\n`
- Example: `-ERR key not found\r\n`
- Used for: Error responses

### Integers
- Format: `:number\r\n`
- Example: `:1\r\n`
- Used for: Counts, boolean results (0 or 1)

### Bulk Strings
- Format: `$length\r\ndata\r\n`
- Example: `$5\r\nhello\r\n`
- Null: `$-1\r\n`
- Used for: Key values, JSON data

### Arrays
- Format: `*count\r\n[elements]\r\n`
- Example: `*2\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n`
- Used for: Multiple values (keys list)

## Commands

All CodexDB commands use the `CDX.` prefix to avoid conflicts with Redis.

### Core Commands

#### CDX.SET
Set a key to a value.

**Format:**
```
*3\r\n
$7\r\nCDX.SET\r\n
$<key_length>\r\n<key>\r\n
$<value_length>\r\n<value>\r\n
```

**Example:**
```
*3\r\n
$7\r\nCDX.SET\r\n
$4\r\nname\r\n
$5\r\nJohn\r\n

Response: +OK\r\n
```

**Telnet:**
```bash
telnet localhost 212
*3
$7
CDX.SET
$4
name
$5
John
# Response: +OK
```

#### CDX.GET
Get the value of a key.

**Format:**
```
*2\r\n
$7\r\nCDX.GET\r\n
$<key_length>\r\n<key>\r\n
```

**Example:**
```
*2\r\n
$7\r\nCDX.GET\r\n
$4\r\nname\r\n

Response: $5\r\nJohn\r\n
```

**Telnet:**
```bash
*2
$7
CDX.GET
$4
name
# Response: $5\r\nJohn\r\n
```

**If key doesn't exist:**
```
Response: $-1\r\n
```

#### CDX.DELETE
Delete a key.

**Format:**
```
*2\r\n
$10\r\nCDX.DELETE\r\n
$<key_length>\r\n<key>\r\n
```

**Response:**
- `:1\r\n` if key was deleted
- `:0\r\n` if key didn't exist

**Example:**
```
*2
$10
CDX.DELETE
$4
name

Response: :1\r\n
```

#### CDX.HAS
Check if a key exists.

**Format:**
```
*2\r\n
$7\r\nCDX.HAS\r\n
$<key_length>\r\n<key>\r\n
```

**Response:**
- `:1\r\n` if key exists
- `:0\r\n` if key doesn't exist

**Example:**
```
*2
$7
CDX.HAS
$4
name

Response: :1\r\n
```

#### CDX.KEYS
List all keys, optionally filtered by pattern.

**Format:**
```
*1\r\n
$8\r\nCDX.KEYS\r\n
```

Or with pattern:
```
*2\r\n
$8\r\nCDX.KEYS\r\n
$<pattern_length>\r\n<pattern>\r\n
```

**Response:** Array of bulk strings

**Example:**
```
*1
$8
CDX.KEYS

Response:
*3\r\n
$4\r\nkey1\r\n
$4\r\nkey2\r\n
$4\r\nkey3\r\n
```

#### CDX.CLEAR
Delete all keys from the database.

**Format:**
```
*1\r\n
$9\r\nCDX.CLEAR\r\n
```

**Response:** `+OK\r\n`

**Example:**
```
*1
$9
CDX.CLEAR

Response: +OK\r\n
```

### Admin Commands

#### CDX.PING
Ping the server (health check).

**Format:**
```
*1\r\n
$8\r\nCDX.PING\r\n
```

Or with message:
```
*2\r\n
$8\r\nCDX.PING\r\n
$<message_length>\r\n<message>\r\n
```

**Response:**
- Without message: `+PONG\r\n`
- With message: `$<length>\r\n<message>\r\n`

**Example:**
```
*1
$8
CDX.PING

Response: +PONG
```

#### CDX.INFO
Get server information and statistics.

**Format:**
```
*1\r\n
$8\r\nCDX.INFO\r\n
```

**Response:** Bulk string with server info

**Example:**
```
*1
$8
CDX.INFO

Response:
$XXX\r\n
codex_version: 1.0.0
storage_engine: file
database_path: /data/codex.db
num_keys: 42
db_size: 4096
connected_clients: 1
...
\r\n
```

#### CDX.AUTH
Authenticate with an API key (if required).

**Format:**
```
*2\r\n
$8\r\nCDX.AUTH\r\n
$<key_length>\r\n<key>\r\n
```

**Response:**
- Success: `+OK\r\n`
- Failure: `-ERR invalid API key\r\n`

**Note:** If API keys are configured, CDX.AUTH must be the first command sent.

**Example:**
```
*2
$8
CDX.AUTH
$32
my-secret-api-key-here

Response: +OK
```

## Client Examples

### Go

```go
package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	// Connect to RESP server
	conn, err := net.Dial("tcp", "localhost:212")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// SET command
	cmd := "*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	fmt.Fprintf(conn, cmd)

	// Read response
	response, _ := reader.ReadString('\n')
	fmt.Println("SET response:", response) // +OK

	// GET command
	cmd = "*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n"
	fmt.Fprintf(conn, cmd)

	// Read response
	response, _ = reader.ReadString('\n')
	fmt.Println("GET response:", response) // $5
	value, _ := reader.ReadString('\n')
	fmt.Println("Value:", value) // value
}
```

### Python

```python
import socket

# Connect to RESP server
conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn.connect(('localhost', 212))

# SET command
cmd = b"*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
conn.sendall(cmd)

# Read response
response = conn.recv(1024)
print("SET response:", response)  # b'+OK\r\n'

# GET command
cmd = b"*2\r\n$7\r\nCDX.GET\r\n$3\r\nkey\r\n"
conn.sendall(cmd)

# Read response
response = conn.recv(1024)
print("GET response:", response)  # b'$5\r\nvalue\r\n'

conn.close()
```

### Node.js

```javascript
const net = require('net');

const socket = net.createConnection({
  host: 'localhost',
  port: 212
});

socket.on('connect', () => {
  console.log('Connected to RESP server');

  // SET command
  socket.write('*3\r\n$7\r\nCDX.SET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n');
});

socket.on('data', (data) => {
  console.log('Response:', data.toString());
});

socket.on('error', (err) => {
  console.error('Error:', err);
});
```

### Shell/Bash

```bash
#!/bin/bash

# Using nc (netcat)
{
  echo -ne "*3\r\n\$7\r\nCDX.SET\r\n\$3\r\nkey\r\n\$5\r\nvalue\r\n"
  sleep 0.1
  echo -ne "*2\r\n\$7\r\nCDX.GET\r\n\$3\r\nkey\r\n"
  sleep 0.1
} | nc localhost 212
```

## Performance Comparison

### Throughput Test

```bash
# Using redis-benchmark (compatible with RESP)
redis-benchmark -h localhost -p 212 -t set,get -n 100000 -c 10
```

Expected results on modern hardware (single connection):

| Operation | HTTP | RESP | Improvement |
|-----------|------|------|------------|
| SET (1KB) | 500 ops/sec | 2000 ops/sec | 4x |
| GET (1KB) | 600 ops/sec | 2200 ops/sec | 3.6x |
| DEL | 700 ops/sec | 3000 ops/sec | 4.3x |
| Batch (10 ops) | 60 ops/sec | 300 ops/sec | 5x |

### Latency Test

```bash
# Measure single operation latency
time redis-benchmark -h localhost -p 212 -t set -n 1 -c 1 -q
```

Expected latency (p50 / p99):
- **HTTP**: 1.2ms / 3.5ms
- **RESP**: 0.3ms / 0.8ms

## Error Handling

All errors are returned as RESP error messages:

```
-ERR <error message>\r\n
```

### Common Errors

```
-ERR invalid command
-ERR wrong number of arguments
-ERR key not found
-ERR authentication required
-ERR invalid API key
-ERR database error
```

Example:

```bash
*2
$7
CDX.GET
# Missing key argument

Response: -ERR wrong number of arguments for 'CDX.GET' command
```

## Kubernetes Deployment

### Service Configuration

The RESP port is exposed in Kubernetes as a named service port:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: codex-service
spec:
  ports:
  - name: http
    port: 80
    targetPort: 11111
  - name: resp
    port: 212
    targetPort: 212
  selector:
    app: codex
```

### Accessing RESP in Kubernetes

From within cluster:
```bash
# Using service DNS
telnet codex-service 212
```

From outside cluster (port-forward):
```bash
kubectl port-forward svc/codex-service 212:212
# Then connect to localhost:212
```

### Environment Configuration

Set RESP port via ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: codex-config
data:
  CODEX_RESP_PORT: "212"
```

## Security Considerations

### Authentication

1. Configure API keys in environment variable:
```bash
export CODEX_API_KEYS="key1,key2,key3"
```

2. On each RESP connection, send authentication before other commands:
```
*2\r\n
$8\r\nCDX.AUTH\r\n
$<key_length>\r\n<your-api-key>\r\n
```

3. All subsequent commands on that connection are authenticated.

### Network Security

- Keep RESP port (212) behind firewall in production
- Use Kubernetes NetworkPolicies to restrict access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codex-resp-restrict
spec:
  podSelector:
    matchLabels:
      app: codex
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          role: codex-client
    ports:
    - protocol: TCP
      port: 212
```

## Migration from HTTP to RESP

### Step 1: Update Connection Code

**Before (HTTP):**
```go
resp, err := http.Post("http://localhost:11111/set",
  "application/json",
  strings.NewReader(`{"key":"k1","value":"v1"}`))
```

**After (RESP):**
```go
conn, err := net.Dial("tcp", "localhost:212")
conn.Write([]byte("*3\r\n$7\r\nCDX.SET\r\n$2\r\nk1\r\n$2\r\nv1\r\n"))
```

### Step 2: Update Command Format

Refer to [Commands](#commands) section for RESP format.

### Step 3: Update Response Parsing

**Before (HTTP):**
```go
var result map[string]interface{}
json.Unmarshal(body, &result)
```

**After (RESP):**
```go
// Parse RESP format manually or use a RESP library
```

### Step 4: Test Thoroughly

1. Unit tests with RESP client
2. Load tests comparing HTTP and RESP
3. Gradual rollout (canary deployment)

## Troubleshooting

### Connection Refused

```
Error: connection refused on localhost:212
```

**Solution:**
1. Check if RESP server is running: `netstat -ln | grep 212`
2. Check CODEX_RESP_PORT environment variable
3. Review service logs: `kubectl logs <pod-name>`

### Authentication Failed

```
-ERR authentication required
```

**Solution:**
1. Check if API keys are configured
2. Verify API key sent in CDX.AUTH command
3. Ensure CDX.AUTH is first command on connection

### Timeout on Commands

**Solution:**
1. Check network connectivity
2. Monitor server CPU/memory usage
3. Check database file permissions
4. Review application logs

## Advanced Usage

### Connection Pooling

For high-throughput applications, maintain multiple connections:

```go
var conns []net.Conn
for i := 0; i < 10; i++ {
    conn, _ := net.Dial("tcp", "localhost:212")
    conns = append(conns, conn)
}

// Round-robin distribute commands across connections
for i, cmd := range commands {
    conns[i % len(conns)].Write([]byte(cmd))
}
```

### Pipelining

Send multiple commands before reading responses:

```
*3\r\nCDX.SET\r\nk1\r\nv1\r\n
*3\r\nCDX.SET\r\nk2\r\nv2\r\n
*2\r\nCDX.GET\r\nk1\r\n
*2\r\nCDX.GET\r\nk2\r\n

# Read 4 responses
```

### Monitoring

Query server info periodically:

```
*1\r\n$8\r\nCDX.INFO\r\n
```

Parse returned statistics for monitoring:
- `num_keys`: Current key count
- `db_size`: Database file size
- `connected_clients`: Current connections

## References

- [RESP Protocol Specification](https://redis.io/docs/reference/protocol-spec/)
- [CodexDB HTTP API](./HTTP_API.md)
- [CodexDB CLI to Service Guide](./CLI_TO_SERVICE.md)
- [Kubernetes Deployment Guide](./KUBERNETES_DEPLOYMENT.md)

## Support

For issues, questions, or feature requests:
- GitHub Issues: https://github.com/evertonmj/codex/issues
- Documentation: https://github.com/evertonmj/codex/docs
