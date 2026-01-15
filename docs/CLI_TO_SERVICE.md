# Using CodexDB HTTP Service from Command Line

## Overview

This guide shows how to interact with the CodexDB HTTP service using standard command-line tools like `curl`. This allows you to perform the same operations as `codex-cli`, but against a remote service instead of a local database file.

## Prerequisites

- CodexDB HTTP service running (see [KUBERNETES_DEPLOYMENT.md](KUBERNETES_DEPLOYMENT.md) or [DOCKER_BUILD.md](DOCKER_BUILD.md))
- Service URL (default: `http://localhost:11111`)
- API key configured on the service

## Environment Setup

Set these environment variables for convenience:

```bash
export CODEX_SERVER_URL="http://localhost:11111"
export CODEX_API_KEY="your-api-key-here"
```

## Command Equivalents

### Set a Value

**codex-cli:**
```bash
codex-cli --file mydb.db set mykey '"hello world"'
```

**HTTP Service (curl):**
```bash
curl -X PUT "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  -H "X-API-Key: $CODEX_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"value": "hello world"}'
```

**HTTP Service (httpie):**
```bash
http PUT "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  X-API-Key:$CODEX_API_KEY \
  value="hello world"
```

### Get a Value

**codex-cli:**
```bash
codex-cli --file mydb.db get mykey
```

**HTTP Service (curl):**
```bash
curl -s "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  -H "X-API-Key: $CODEX_API_KEY" | jq '.value'
```

**HTTP Service (httpie):**
```bash
http GET "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  X-API-Key:$CODEX_API_KEY
```

### Delete a Key

**codex-cli:**
```bash
codex-cli --file mydb.db delete mykey
```

**HTTP Service (curl):**
```bash
curl -X DELETE "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  -H "X-API-Key: $CODEX_API_KEY"
```

**HTTP Service (httpie):**
```bash
http DELETE "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  X-API-Key:$CODEX_API_KEY
```

### Check if Key Exists

**codex-cli:**
```bash
codex-cli --file mydb.db has mykey
```

**HTTP Service (curl):**
```bash
curl -I "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  -H "X-API-Key: $CODEX_API_KEY" | grep "HTTP/" | awk '{print ($2 == 200) ? "true" : "false"}'
```

**HTTP Service (httpie):**
```bash
http HEAD "$CODEX_SERVER_URL/api/v1/keys/mykey" \
  X-API-Key:$CODEX_API_KEY
```

### List All Keys

**codex-cli:**
```bash
codex-cli --file mydb.db keys
```

**HTTP Service (curl):**
```bash
curl -s "$CODEX_SERVER_URL/api/v1/keys" \
  -H "X-API-Key: $CODEX_API_KEY" | jq -r '.keys[]'
```

**HTTP Service (httpie):**
```bash
http GET "$CODEX_SERVER_URL/api/v1/keys" \
  X-API-Key:$CODEX_API_KEY
```

### Clear All Keys

**codex-cli:**
```bash
codex-cli --file mydb.db clear
```

**HTTP Service (curl):**
```bash
curl -X DELETE "$CODEX_SERVER_URL/api/v1/keys" \
  -H "X-API-Key: $CODEX_API_KEY"
```

**HTTP Service (httpie):**
```bash
http DELETE "$CODEX_SERVER_URL/api/v1/keys" \
  X-API-Key:$CODEX_API_KEY
```

## Shell Aliases

Create convenient aliases for common operations:

```bash
# Add to ~/.bashrc or ~/.zshrc

# Set CODEX environment variables
export CODEX_SERVER_URL="http://localhost:11111"
export CODEX_API_KEY="your-api-key"

# Define helper function
codex() {
    local cmd="$1"
    shift

    case "$cmd" in
        set)
            local key="$1"
            local value="$2"
            curl -X PUT "$CODEX_SERVER_URL/api/v1/keys/$key" \
                -H "X-API-Key: $CODEX_API_KEY" \
                -H "Content-Type: application/json" \
                -d "{\"value\": $value}"
            ;;
        get)
            local key="$1"
            curl -s "$CODEX_SERVER_URL/api/v1/keys/$key" \
                -H "X-API-Key: $CODEX_API_KEY" | jq '.value'
            ;;
        delete)
            local key="$1"
            curl -X DELETE "$CODEX_SERVER_URL/api/v1/keys/$key" \
                -H "X-API-Key: $CODEX_API_KEY"
            ;;
        has)
            local key="$1"
            local status=$(curl -s -o /dev/null -w "%{http_code}" \
                -H "X-API-Key: $CODEX_API_KEY" \
                "$CODEX_SERVER_URL/api/v1/keys/$key")
            [ "$status" = "200" ] && echo "true" || echo "false"
            ;;
        keys)
            curl -s "$CODEX_SERVER_URL/api/v1/keys" \
                -H "X-API-Key: $CODEX_API_KEY" | jq -r '.keys[]'
            ;;
        clear)
            curl -X DELETE "$CODEX_SERVER_URL/api/v1/keys" \
                -H "X-API-Key: $CODEX_API_KEY"
            ;;
        *)
            echo "Usage: codex {set|get|delete|has|keys|clear} [args]"
            return 1
            ;;
    esac
}
```

**Usage:**
```bash
# After sourcing the aliases
codex set mykey '"hello world"'
codex get mykey
codex has mykey
codex keys
codex delete mykey
codex clear
```

## Python Script

Create a Python wrapper script for more complex operations:

```python
#!/usr/bin/env python3
"""
codex_client.py - Python client for CodexDB HTTP service
Usage: ./codex_client.py <command> [args]
"""

import os
import sys
import json
import requests

class CodexClient:
    def __init__(self, server_url=None, api_key=None):
        self.server_url = server_url or os.getenv('CODEX_SERVER_URL', 'http://localhost:11111')
        self.api_key = api_key or os.getenv('CODEX_API_KEY')

        if not self.api_key:
            raise ValueError("CODEX_API_KEY environment variable or api_key parameter required")

        self.headers = {
            'X-API-Key': self.api_key,
            'Content-Type': 'application/json'
        }

    def set(self, key, value):
        """Set a value for a key"""
        url = f"{self.server_url}/api/v1/keys/{key}"
        response = requests.put(url, headers=self.headers, json={'value': value})
        response.raise_for_status()
        return response.json()

    def get(self, key):
        """Get a value by key"""
        url = f"{self.server_url}/api/v1/keys/{key}"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()['value']

    def delete(self, key):
        """Delete a key"""
        url = f"{self.server_url}/api/v1/keys/{key}"
        response = requests.delete(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

    def has(self, key):
        """Check if a key exists"""
        url = f"{self.server_url}/api/v1/keys/{key}"
        response = requests.head(url, headers=self.headers)
        return response.status_code == 200

    def keys(self):
        """List all keys"""
        url = f"{self.server_url}/api/v1/keys"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()['keys']

    def clear(self):
        """Clear all keys"""
        url = f"{self.server_url}/api/v1/keys"
        response = requests.delete(url, headers=self.headers)
        response.raise_for_status()
        return response.json()

def main():
    if len(sys.argv) < 2:
        print("Usage: codex_client.py <command> [args]")
        print("Commands: set, get, delete, has, keys, clear")
        sys.exit(1)

    client = CodexClient()
    command = sys.argv[1]

    try:
        if command == 'set':
            if len(sys.argv) != 4:
                print("Usage: codex_client.py set <key> <json_value>")
                sys.exit(1)
            key = sys.argv[2]
            value = json.loads(sys.argv[3])
            result = client.set(key, value)
            print(json.dumps(result, indent=2))

        elif command == 'get':
            if len(sys.argv) != 3:
                print("Usage: codex_client.py get <key>")
                sys.exit(1)
            key = sys.argv[2]
            value = client.get(key)
            print(json.dumps(value, indent=2))

        elif command == 'delete':
            if len(sys.argv) != 3:
                print("Usage: codex_client.py delete <key>")
                sys.exit(1)
            key = sys.argv[2]
            result = client.delete(key)
            print(json.dumps(result, indent=2))

        elif command == 'has':
            if len(sys.argv) != 3:
                print("Usage: codex_client.py has <key>")
                sys.exit(1)
            key = sys.argv[2]
            exists = client.has(key)
            print("true" if exists else "false")

        elif command == 'keys':
            keys = client.keys()
            for key in keys:
                print(key)

        elif command == 'clear':
            result = client.clear()
            print(json.dumps(result, indent=2))

        else:
            print(f"Unknown command: {command}")
            sys.exit(1)

    except requests.exceptions.HTTPError as e:
        print(f"HTTP Error: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
```

**Installation:**
```bash
chmod +x codex_client.py
pip install requests
```

**Usage:**
```bash
export CODEX_SERVER_URL="http://localhost:11111"
export CODEX_API_KEY="your-api-key"

./codex_client.py set mykey '"hello world"'
./codex_client.py get mykey
./codex_client.py has mykey
./codex_client.py keys
./codex_client.py delete mykey
./codex_client.py clear
```

## Go CLI Wrapper

Create a lightweight Go wrapper that uses the HTTP service:

```go
// File: cmd/codex-http-cli/main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	serverURL := os.Getenv("CODEX_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:11111"
	}

	apiKey := os.Getenv("CODEX_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: CODEX_API_KEY environment variable not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: codex-http-cli <command> [args]")
		fmt.Fprintln(os.Stderr, "Commands: set, get, delete, has, keys, clear")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	if err := executeCommand(serverURL, apiKey, command, args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func executeCommand(serverURL, apiKey, command string, args []string) error {
	client := &http.Client{}

	switch command {
	case "set":
		if len(args) < 2 {
			return fmt.Errorf("usage: set <key> <json_value>")
		}
		key := args[0]
		valueStr := strings.Join(args[1:], " ")

		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return fmt.Errorf("invalid JSON value: %v", err)
		}

		reqBody, _ := json.Marshal(map[string]interface{}{"value": value})
		req, _ := http.NewRequest("PUT", serverURL+"/api/v1/keys/"+key, bytes.NewReader(reqBody))
		req.Header.Set("X-API-Key", apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		fmt.Println("OK")

	case "get":
		if len(args) != 1 {
			return fmt.Errorf("usage: get <key>")
		}
		key := args[0]

		req, _ := http.NewRequest("GET", serverURL+"/api/v1/keys/"+key, nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		jsonVal, _ := json.MarshalIndent(result["value"], "", "  ")
		fmt.Println(string(jsonVal))

	case "delete":
		if len(args) != 1 {
			return fmt.Errorf("usage: delete <key>")
		}
		key := args[0]

		req, _ := http.NewRequest("DELETE", serverURL+"/api/v1/keys/"+key, nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		fmt.Println("OK")

	case "has":
		if len(args) != 1 {
			return fmt.Errorf("usage: has <key>")
		}
		key := args[0]

		req, _ := http.NewRequest("HEAD", serverURL+"/api/v1/keys/"+key, nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("true")
		} else {
			fmt.Println("false")
		}

	case "keys":
		req, _ := http.NewRequest("GET", serverURL+"/api/v1/keys", nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}

		var result map[string][]string
		json.NewDecoder(resp.Body).Decode(&result)
		fmt.Println(strings.Join(result["keys"], "\n"))

	case "clear":
		req, _ := http.NewRequest("DELETE", serverURL+"/api/v1/keys", nil)
		req.Header.Set("X-API-Key", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed: %s", resp.Status)
		}
		fmt.Println("OK")

	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}
```

**Build and use:**
```bash
go build -o codex-http-cli cmd/codex-http-cli/main.go

export CODEX_SERVER_URL="http://localhost:11111"
export CODEX_API_KEY="your-api-key"

./codex-http-cli set mykey '"hello world"'
./codex-http-cli get mykey
```

## RESP Protocol (High-Performance Alternative)

For high-throughput scenarios, use the RESP protocol server on port 212 instead of HTTP. RESP provides 3-5x better performance:

**Setup:**
```bash
export CODEX_RESP_SERVER="localhost:212"
export CODEX_API_KEY="your-api-key-here"
```

**Set a value (RESP):**

```bash
{
  echo -ne "*3\r\n\$7\r\nCDX.SET\r\n\$5\r\nmykey\r\n\$11\r\nhello world\r\n"
  sleep 0.1
} | nc $CODEX_RESP_SERVER

# Response: +OK
```

**Get a value (RESP):**

```bash
{
  echo -ne "*2\r\n\$7\r\nCDX.GET\r\n\$5\r\nmykey\r\n"
  sleep 0.1
} | nc $CODEX_RESP_SERVER

# Response: $11\r\nhello world\r\n
```

**Using Go client:**

```go
// See RESP_PROTOCOL.md for complete Go client example
conn, _ := net.Dial("tcp", "localhost:212")
cmd := "*3\r\n$7\r\nCDX.SET\r\n$5\r\nmykey\r\n$11\r\nhello world\r\n"
conn.Write([]byte(cmd))
```

**Using Python client:**

```python
import socket

conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn.connect(('localhost', 212))

# SET command
cmd = b"*3\r\n$7\r\nCDX.SET\r\n$5\r\nmykey\r\n$11\r\nhello world\r\n"
conn.sendall(cmd)

response = conn.recv(1024)
print(response)  # b'+OK\r\n'
```

For detailed RESP documentation, see [RESP Protocol Guide](./RESP_PROTOCOL.md).

## Comparison: CLI vs HTTP vs RESP

| Feature | codex-cli (Direct) | HTTP Service | RESP Server |
| --- | --- | --- | --- |
| **Access Mode** | Direct database file | Remote HTTP API | Remote RESP Protocol |
| **Concurrency** | Single process | Multiple clients | Multiple clients |
| **Deployment** | Local file system | Kubernetes/Docker | Kubernetes/Docker |
| **Authentication** | File permissions | API key | API key |
| **Network** | Local only | Network accessible | Network accessible |
| **Performance** | Baseline (0.1-0.5ms) | Good (1-5ms) | Excellent (0.3-1ms) |
| **Throughput** | 2000-10000 ops/sec | 200-1000 ops/sec | 1000-3000 ops/sec |
| **Best For** | Development | Web services | High-throughput services |

## When to Use Each Approach

### Use codex-cli (Direct Access) When:
- Working with local databases
- Single-user/single-process access
- No network access needed
- Development and testing
- Batch processing scripts

### Use HTTP Service When:
- Multiple clients need access
- Running in Kubernetes/Docker
- Need network accessibility
- Production deployments
- Centralized data management
- Need authentication and access control

## See Also

- [HTTP API Reference](HTTP_API.md) - Complete API documentation
- [Kubernetes Deployment](KUBERNETES_DEPLOYMENT.md) - Deploy service to K8s
- [Docker Build Guide](DOCKER_BUILD.md) - Run service in Docker
- [Service Deployment Summary](SERVICE_DEPLOYMENT_SUMMARY.md) - Quick reference
