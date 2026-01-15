# CodexDB Docker Build & Deployment Guide

## Overview

This guide covers building Docker images for CodexDB HTTP Service and running containers locally.

## Prerequisites

- **Docker** (version 20.10+) - Download from [docker.com](https://docker.com)
- **Make** (optional) - For using build targets
- CodexDB source code

## Building the Docker Image

### Option 1: Using Make (Recommended)

The easiest way to build the Docker image:

```bash
# Build the Docker image
make docker-build

# Build and run in one command
make docker-run
```

### Option 2: Manual Docker Build

```bash
# Build from the Dockerfile
docker build -t codex-service:latest -f deployments/docker/Dockerfile .

# Build with a custom tag
docker build -t codex-service:v1.0.0 -f deployments/docker/Dockerfile .

# Build with custom registry
docker build -t myregistry.azurecr.io/codex-service:latest -f deployments/docker/Dockerfile .
```

### Build Options

```bash
# Build with build args
docker build --build-arg BUILDKIT_INLINE_CACHE=1 \
  -t codex-service:latest \
  -f deployments/docker/Dockerfile \
  .

# Build with progress output
docker build --progress=plain \
  -t codex-service:latest \
  -f deployments/docker/Dockerfile \
  .
```

## Running the Container

### Basic Run

```bash
# Run with default settings
docker run -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  codex-service:latest
```

### With Configuration

```bash
# Run with custom configuration
docker run -p 8080:8080 \
  -e CODEX_API_KEYS=key1,key2,key3 \
  -e CODEX_DB_PATH=/data/codex.db \
  -e CODEX_COMPRESSION=gzip \
  -e CODEX_LOG_LEVEL=debug \
  -v codex-data:/data \
  codex-service:latest
```

### With Volume Mounting

```bash
# Run with persistent storage on host
docker run -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  -v $(pwd)/data:/data \
  codex-service:latest
```

### As a Background Service

```bash
# Run as daemon
docker run -d \
  --name codex \
  -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  -v codex-data:/data \
  codex-service:latest

# View logs
docker logs -f codex

# Stop container
docker stop codex

# Restart container
docker restart codex

# Remove container
docker rm codex
```

## Testing the Container

### Health Check

```bash
# Check if service is running
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy"}
```

### Test API Operations

```bash
# Set a value
curl -X PUT http://localhost:8080/api/v1/keys/test \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-key" \
  -d '{"value": "hello world"}'

# Get the value
curl http://localhost:8080/api/v1/keys/test \
  -H "X-API-Key: test-key"

# List all keys
curl http://localhost:8080/api/v1/keys \
  -H "X-API-Key: test-key"

# Expected response:
# {"keys":["test"]}
```

### Container Shell Access

```bash
# Access container shell
docker exec -it codex /bin/sh

# Inside the container:
ls -la /data
cat /data/codex.db

# Exit
exit
```

## Docker Compose (Development)

Create a `docker-compose.yml` for local development:

```yaml
version: '3.8'

services:
  codex:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile
    container_name: codex-service
    ports:
      - "8080:8080"
    environment:
      CODEX_PORT: "8080"
      CODEX_HOST: "0.0.0.0"
      CODEX_API_KEYS: "dev-key-1,dev-key-2"
      CODEX_DB_PATH: "/data/codex.db"
      CODEX_LEDGER_MODE: "false"
      CODEX_COMPRESSION: "gzip"
      CODEX_LOG_LEVEL: "info"
    volumes:
      - codex-data:/data
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    restart: unless-stopped

volumes:
  codex-data:
    driver: local
```

Run with Docker Compose:

```bash
# Start services
docker-compose up

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f codex

# Stop services
docker-compose down

# Remove volumes
docker-compose down -v
```

## Image Optimization

### Image Size

The multi-stage build is already optimized:

```bash
# Check image size
docker images | grep codex-service

# Expected size: ~15-20MB
```

To further reduce size:

```dockerfile
# Use distroless instead of alpine (even smaller)
FROM gcr.io/distroless/base-debian11

# Copy only the binary
COPY --from=builder /build/codex-service /app/codex-service
```

### Layer Caching

The Dockerfile is organized to maximize layer caching:

1. Copy go.mod/go.sum (slow to change)
2. Download dependencies (cached)
3. Copy source code (changes frequently)
4. Build binary

This means only the last steps run when code changes.

### Build Time

```bash
# Benchmark build time
time docker build -t codex-service:latest -f deployments/docker/Dockerfile .

# Typical build time: 30-60 seconds (first build)
# Subsequent builds: 5-10 seconds (with caching)
```

## Registry Operations

### Docker Hub

```bash
# Login to Docker Hub
docker login

# Tag image
docker tag codex-service:latest myusername/codex-service:latest

# Push to Docker Hub
docker push myusername/codex-service:latest

# Pull from Docker Hub
docker pull myusername/codex-service:latest
```

### Azure Container Registry (ACR)

```bash
# Login to ACR
az acr login --name myregistry

# Tag image
docker tag codex-service:latest myregistry.azurecr.io/codex-service:v1.0.0

# Push to ACR
docker push myregistry.azurecr.io/codex-service:v1.0.0

# Pull from ACR
docker pull myregistry.azurecr.io/codex-service:v1.0.0
```

### AWS Elastic Container Registry (ECR)

```bash
# Get login token
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Tag image
docker tag codex-service:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/codex-service:latest

# Push to ECR
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/codex-service:latest
```

### Google Container Registry (GCR)

```bash
# Authenticate
gcloud auth configure-docker

# Tag image
docker tag codex-service:latest gcr.io/my-project/codex-service:latest

# Push to GCR
docker push gcr.io/my-project/codex-service:latest
```

## Container Security

### Run as Non-Root

The container automatically runs as user `codex` (uid 1000):

```bash
# Verify non-root
docker run --rm codex-service:latest id

# Expected output:
# uid=1000(codex) gid=1000(codex) groups=1000(codex)
```

### Read-Only Root Filesystem

For production, consider making the root filesystem read-only:

```bash
docker run --read-only \
  --tmpfs /tmp \
  -v codex-data:/data \
  -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  codex-service:latest
```

### Security Scanning

Scan the image for vulnerabilities:

```bash
# Using Trivy
trivy image codex-service:latest

# Using Docker Scout
docker scout cves codex-service:latest
```

## Troubleshooting

### Container fails to start

```bash
# Check logs
docker logs codex

# Run with interactive terminal
docker run -it \
  -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  codex-service:latest

# Should show startup logs
```

### Cannot connect to service

```bash
# Check if container is running
docker ps

# Check port mapping
docker port codex

# Test from host
curl http://localhost:8080/health

# Test from container
docker exec codex wget -q -O - http://localhost:8080/health
```

### Database issues

```bash
# Check volume
docker inspect codex | grep -A 10 Mounts

# Inspect volume
docker volume inspect codex-data

# Backup volume
docker run --rm -v codex-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/codex-backup.tar.gz -C /data .

# Restore volume
docker run --rm -v codex-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/codex-backup.tar.gz -C /data
```

### High memory usage

```bash
# Check memory usage
docker stats codex

# Limit memory
docker run --memory=512m \
  -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  codex-service:latest
```

## Production Deployment

### Environment-Specific Configuration

Development:
```bash
docker run -e CODEX_LOG_LEVEL=debug \
  -e CODEX_COMPRESSION=none \
  codex-service:latest
```

Production:
```bash
docker run -e CODEX_LOG_LEVEL=info \
  -e CODEX_COMPRESSION=gzip \
  --memory=2g \
  --cpus=2 \
  codex-service:latest
```

### Backup Strategy

```bash
# Create backup
docker exec codex tar czf /tmp/backup.tar.gz /data/
docker cp codex:/tmp/backup.tar.gz ./codex-backup-$(date +%Y%m%d).tar.gz

# Restore from backup
docker cp ./codex-backup-20240115.tar.gz codex:/tmp/
docker exec codex tar xzf /tmp/codex-backup-20240115.tar.gz -C /
```

### Health Monitoring

```bash
# Set up health check (in docker-compose or docker run)
docker run --health-cmd='curl -f http://localhost:8080/health || exit 1' \
  --health-interval=30s \
  --health-timeout=5s \
  --health-retries=3 \
  -p 8080:8080 \
  -e CODEX_API_KEYS=test-key \
  codex-service:latest
```

## Multi-Architecture Images (ARM, x86_64)

Build for multiple architectures:

```bash
# Install buildx (if not already installed)
docker buildx create --name builder --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64,linux/arm/v7 \
  -t myregistry/codex-service:latest \
  --push \
  -f deployments/docker/Dockerfile \
  .
```

## Performance Tips

1. **Use volume mounts** for better performance than named volumes on Mac/Windows
2. **Enable compression** to reduce network overhead
3. **Use health checks** for better container orchestration
4. **Limit resources** to prevent one container from consuming all resources
5. **Use read-only mounts** where possible for security

## Example Deployment Script

```bash
#!/bin/bash
set -e

# Build image
echo "Building Docker image..."
docker build -t codex-service:latest -f deployments/docker/Dockerfile .

# Create volume
echo "Creating volume..."
docker volume create codex-data || true

# Run container
echo "Starting container..."
docker run -d \
  --name codex \
  --restart always \
  -p 8080:8080 \
  -e CODEX_API_KEYS="prod-key-1,prod-key-2" \
  -e CODEX_COMPRESSION=gzip \
  -e CODEX_LOG_LEVEL=info \
  -v codex-data:/data \
  --health-cmd='curl -f http://localhost:8080/health || exit 1' \
  --health-interval=30s \
  --health-timeout=5s \
  --health-retries=3 \
  codex-service:latest

# Wait for startup
echo "Waiting for startup..."
sleep 5

# Test service
echo "Testing service..."
curl -X GET http://localhost:8080/health \
  -H "X-API-Key: prod-key-1" || exit 1

echo "✓ Deployment successful!"
```

Save as `deploy.sh` and run:

```bash
chmod +x deploy.sh
./deploy.sh
```

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Best Practices for Python/Go Docker Images](https://docs.docker.com/develop/dev-best-practices/)
- [Docker Security](https://docs.docker.com/engine/security/)
