# CodexDB Data Storage Configuration

## Overview

CodexDB HTTP Service uses an intelligent path resolution system that automatically determines the best location for storing database files based on the execution environment.

## Storage Locations

### Local Execution (Default)
When running the service locally (development, testing):

```
~/.codex_data/codex.db
```

**Example paths:**
- macOS/Linux: `/Users/username/.codex_data/codex.db`
- Linux: `/home/username/.codex_data/codex.db`
- Windows: `C:\Users\username\.codex_data\codex.db`

**Advantages:**
- ✅ Stores data in user's home directory
- ✅ No root filesystem permissions needed
- ✅ Easy to backup and manage
- ✅ Automatic directory creation
- ✅ Safe for development and testing

### Kubernetes Execution
When running in Kubernetes (production):

```
/data/codex.db
```

**Inside Container:**
- Path: `/data/codex.db`
- Mounted to: PersistentVolumeClaim
- Access mode: ReadWriteOnce
- Managed by: Kubernetes StatefulSet

**Advantages:**
- ✅ Persistent across pod restarts
- ✅ Easy to backup with volume snapshots
- ✅ Automatic volume management
- ✅ Complies with container best practices
- ✅ Pod-specific isolation

## Auto-Detection

The service automatically detects the environment and selects the appropriate path:

### Detection Logic

```
1. Check CODEX_KUBERNETES environment variable
   ├─ If set to "true" → Use /data/codex.db
   └─ If not set → Use ~/.codex_data/codex.db

2. If CODEX_DB_PATH is explicitly set
   └─ Use the provided path (overrides auto-detection)

3. Create directory if it doesn't exist
   └─ Permissions: 0755 (rwxr-xr-x)
```

## Configuration

### Override the Default Path

You can explicitly set the database path using the `CODEX_DB_PATH` environment variable:

**Local execution:**
```bash
# Use custom location
CODEX_DB_PATH=$HOME/my_codex_data/db.db make run-service

# Use /tmp (temporary)
CODEX_DB_PATH=/tmp/codex.db make run-service
```

**Kubernetes deployment:**
```bash
# Keep using /data (recommended for K8s)
kubectl set env statefulset codex CODEX_DB_PATH=/data/codex.db

# Or modify the ConfigMap
kubectl edit configmap codex-config
```

### Environment Variables

| Variable | Value | Effect |
|----------|-------|--------|
| `CODEX_KUBERNETES` | `true` | Use `/data/codex.db` |
| `CODEX_KUBERNETES` | `false` (default) | Use `~/.codex_data/codex.db` |
| `CODEX_DB_PATH` | Any path | Override all defaults |

## Local Development Setup

### Starting the Service

**Option 1: Using Make (Recommended)**
```bash
make run-service
```

Database will be created at: `~/.codex_data/codex.db`

**Option 2: Direct Execution**
```bash
export CODEX_API_KEYS=dev-key
./bin/codex-service
```

Database will be created at: `~/.codex_data/codex.db`

**Option 3: Custom Location**
```bash
export CODEX_API_KEYS=dev-key
export CODEX_DB_PATH=$HOME/my_db/codex.db
./bin/codex-service
```

Database will be created at: `$HOME/my_db/codex.db`

### Verifying Local Storage

```bash
# Check if directory exists
ls -la ~/.codex_data/

# View database file
file ~/.codex_data/codex.db

# Check database size
du -h ~/.codex_data/codex.db
```

### Cleaning Up Local Data

```bash
# Remove the database
rm -f ~/.codex_data/codex.db*

# Remove the entire directory
rm -rf ~/.codex_data/
```

## Kubernetes Deployment Setup

### Default Configuration

The Kubernetes manifests and Helm chart automatically:

1. ✅ Set `CODEX_KUBERNETES=true` in ConfigMap
2. ✅ Mount PersistentVolume at `/data`
3. ✅ Create database at `/data/codex.db`
4. ✅ Persist data across pod restarts

### Deployment Example

**Using Helm:**
```bash
helm install codex deployments/helm/codex \
  --set security.apiKeys[0]=prod-key
```

**Using raw YAML:**
```bash
# Apply manifests in order
kubectl apply -f deployments/kubernetes/configmap.yaml
kubectl create secret generic codex-secrets \
  --from-literal=CODEX_API_KEYS="prod-key"
kubectl apply -f deployments/kubernetes/statefulset.yaml
kubectl apply -f deployments/kubernetes/service.yaml
```

### Verify Kubernetes Storage

```bash
# Check the pod
kubectl get pods -l app=codex

# Check the PVC
kubectl get pvc

# Access the pod and check the file
kubectl exec codex-0 -- ls -la /data/
kubectl exec codex-0 -- du -h /data/codex.db

# Backup the database
kubectl cp codex-0:/data/codex.db ./codex-backup.db
```

## Data Persistence

### Local Persistence

Data stored in `~/.codex_data/codex.db` persists until:
- You manually delete the file/directory
- The home directory is deleted
- The system storage is wiped

**Backup strategy:**
```bash
# Manual backup
cp ~/.codex_data/codex.db ~/.codex_data/codex.db.backup.$(date +%Y%m%d)

# Script-based backup
#!/bin/bash
SOURCE="$HOME/.codex_data"
DEST="$HOME/backups/codex"
mkdir -p "$DEST"
tar czf "$DEST/codex-$(date +%Y%m%d-%H%M%S).tar.gz" "$SOURCE"
```

### Kubernetes Persistence

Data stored in `/data/codex.db` persists across:
- ✅ Pod restarts
- ✅ Container crashes
- ✅ Deployments
- ✅ Node maintenance (via pod rescheduling)

Data is lost when:
- ❌ PersistentVolumeClaim is deleted
- ❌ PersistentVolume is deleted
- ❌ Storage backend fails

**Backup strategy:**
```bash
# Using kubectl cp
kubectl cp codex-0:/data/codex.db ./backups/codex-$(date +%Y%m%d).db

# Using Kubernetes VolumeSnapshot
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: codex-backup-daily
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: codex-data-codex-0
```

## Directory Structure

### Local Directories Created

```
~/.codex_data/
├── codex.db              # Main database file
├── codex.db.lock         # Lock file (temporary)
├── codex.db.bak.1        # Latest backup
├── codex.db.bak.2        # Previous backups
└── codex.db.bak.N        # Old backups (configurable count)
```

### Kubernetes Directories

```
/data/
└── codex.db              # Main database file in container
```

Backed by PersistentVolumeClaim:
```
<storage-system>/
└── codex-data-codex-0/  # PVC for pod 0
    └── codex.db
```

## Multi-Instance Considerations

### Local Machine (Single Instance)

If running multiple instances locally:
```bash
# Instance 1
CODEX_DB_PATH=$HOME/.codex_data/instance1.db ./bin/codex-service

# Instance 2
CODEX_DB_PATH=$HOME/.codex_data/instance2.db ./bin/codex-service
```

Each instance gets its own database file.

### Kubernetes (Multiple Pods)

Each StatefulSet pod automatically gets:
- ✅ Its own PersistentVolumeClaim
- ✅ Its own `/data/codex.db` file
- ✅ Independent database storage
- ✅ No cross-pod interference

```bash
# Scale to 3 replicas
kubectl scale statefulset codex --replicas=3

# Each pod (codex-0, codex-1, codex-2) gets own PVC
kubectl get pvc
# codex-data-codex-0  Bound
# codex-data-codex-1  Bound
# codex-data-codex-2  Bound
```

## Troubleshooting

### "mkdir: permission denied"

**Symptom:** Service fails with "mkdir: permission denied" error

**Solution:**
- Ensure home directory is writable
- Check disk space: `df -h $HOME`
- Verify permissions: `ls -ld $HOME`

```bash
# Fix permissions
chmod 755 $HOME
```

### "read-only file system"

**Symptom:** In Kubernetes: "mkdir /data: read-only file system"

**Solution:**
- Ensure PersistentVolume is writable
- Check pod securityContext
- Verify volume mount permissions

```bash
# Check volume mount
kubectl exec codex-0 -- mount | grep data

# Check securityContext
kubectl get pod codex-0 -o yaml | grep -A 10 securityContext
```

### "No space left on device"

**Symptom:** Service fails with "no space left on device"

**Solutions:**
- Check disk space: `df -h`
- Reduce number of backups: `CODEX_NUM_BACKUPS=2`
- Expand PersistentVolume capacity

```bash
# Local: Remove old backups
rm ~/.codex_data/codex.db.bak.*

# Kubernetes: Expand PVC
kubectl patch pvc codex-data-codex-0 -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'
```

### "Permission denied" for database file

**Symptom:** Cannot read/write database file

**Solution:**
- Check file permissions
- Ensure process owner has access

```bash
# Check local file permissions
ls -la ~/.codex_data/codex.db

# Fix permissions (if needed)
chmod 600 ~/.codex_data/codex.db
```

## Security Considerations

### Local Permissions

The database file is created with:
- **Owner:** Current user
- **Permissions:** 0600 (rw-------)
- **Accessible by:** Owner only

```bash
# Verify secure permissions
ls -la ~/.codex_data/codex.db
# -rw-------  1 user  staff  12345 Jan 15 12:00 codex.db
```

### Kubernetes Security

- ✅ Pod runs as non-root user (uid=1000)
- ✅ Volume mounted with restricted permissions
- ✅ SecurityContext enforces restrictions
- ✅ Network policies can limit access

### Backup Security

When backing up database files:

```bash
# Local: Preserve permissions
cp -p ~/.codex_data/codex.db /backup/codex.db

# Kubernetes: Secure transfer
kubectl cp codex-0:/data/codex.db ./codex.db
chmod 600 ./codex.db  # Restrict access
```

## Summary

| Aspect | Local | Kubernetes |
|--------|-------|-----------|
| **Default Path** | `~/.codex_data/codex.db` | `/data/codex.db` |
| **Auto-Detection** | ✅ Yes | ✅ Yes (via env var) |
| **Persistence** | Manual backup needed | Automatic via PVC |
| **Multi-Instance** | Each gets own file | Each pod gets own PVC |
| **Permissions** | User home directory | Container filesystem |
| **Scaling** | Limited by disk space | Limited by storage size |

The intelligent path resolution system ensures CodexDB works seamlessly in both local development and production Kubernetes environments! 🎯
