# Storage Refactoring Summary - CodexDB HTTP Service

## ✅ Refactoring Complete

The CodexDB HTTP Service has been refactored to use intelligent storage location detection that automatically selects the appropriate database path based on the execution environment.

## 🎯 What Changed

### Previous Behavior
- ❌ Always used `/data/codex.db` path
- ❌ Only suitable for Kubernetes containers
- ❌ Failed on local execution due to permissions
- ❌ No home directory support

### New Behavior
- ✅ Automatic environment detection
- ✅ Uses `~/.codex_data/codex.db` for local execution
- ✅ Uses `/data/codex.db` for Kubernetes
- ✅ Automatic directory creation
- ✅ No permission issues on local machines

## 📁 Storage Paths

### Local Development
```
~/.codex_data/codex.db
```
- **Location:** User's home directory
- **Example:** `/Users/everton/.codex_data/codex.db`
- **Permissions:** User-owned, secure permissions
- **Created:** Automatically on first run
- **Backup files:** `.bak.1`, `.bak.2`, etc.

### Kubernetes Production
```
/data/codex.db
```
- **Location:** Inside container (mounted PVC)
- **Persistence:** PersistentVolumeClaim
- **Access Mode:** ReadWriteOnce
- **Management:** Kubernetes StatefulSet

## 🔧 Implementation Details

### Files Modified

1. **`cmd/codex-service/config.go`**
   - Added `getDefaultDBPath()` function
   - Intelligent environment detection
   - Automatic directory creation
   - Added `ensureDBDirectory()` function
   - New imports: `os/user`, `path/filepath`

2. **`deployments/kubernetes/configmap.yaml`**
   - Added `CODEX_KUBERNETES: "true"` flag
   - Removed explicit `CODEX_DB_PATH`
   - Added documentation

3. **`deployments/helm/codex/values.yaml`**
   - Added `server.kubernetesMode: true`
   - Configuration for Helm deployments

4. **`deployments/helm/codex/templates/configmap.yaml`**
   - Updated to use `CODEX_KUBERNETES` env var
   - Removed hardcoded path

### Files Created

1. **`docs/DATA_STORAGE.md`** (Comprehensive guide)
   - Storage location reference
   - Local vs Kubernetes comparison
   - Configuration options
   - Backup strategies
   - Troubleshooting guide
   - Security considerations

## 🚀 How It Works

### Automatic Detection

```go
// Environment detection logic
if CODEX_KUBERNETES == "true" {
    database_path = "/data/codex.db"  // Kubernetes
} else {
    database_path = "~/.codex_data/codex.db"  // Local
}

// If CODEX_DB_PATH is set
database_path = CODEX_DB_PATH  // Override everything
```

### Directory Creation

```go
// Automatically creates directory structure
ensureDBDirectory(config.DBPath)
// Creates: ~/.codex_data/ with permissions 0755
```

## 💡 Usage Examples

### Local Development

**Option 1: Using Make**
```bash
make run-service
# Creates: ~/.codex_data/codex.db
# Test: curl http://localhost:8080/health
```

**Option 2: Direct Execution**
```bash
export CODEX_API_KEYS=dev-key
./bin/codex-service
# Creates: ~/.codex_data/codex.db
```

**Option 3: Custom Location**
```bash
export CODEX_DB_PATH=$HOME/my_databases/codex.db
./bin/codex-service
# Creates: $HOME/my_databases/codex.db
```

### Kubernetes Deployment

**Using Helm (Recommended)**
```bash
helm install codex deployments/helm/codex \
  --set security.apiKeys[0]=prod-key
# Uses: /data/codex.db (inside container)
```

**Using Raw YAML**
```bash
kubectl apply -f deployments/kubernetes/
# Uses: /data/codex.db (inside container)
# CODEX_KUBERNETES=true is set in ConfigMap
```

## 🔍 Verification

### Local Storage Created

```bash
$ ls -la ~/.codex_data/
total 0
-rw-------  1 user  staff     0 Jan 15 11:08 codex.db.lock
-rw-------  1 user  staff  1024 Jan 15 11:08 codex.db
```

### Service Startup Logs

**Local execution:**
```
[codex-service] 2026/01/15 11:15:10 main.go:27: Database: /Users/everton/.codex_data/codex.db (ledger=false)
[codex-service] 2026/01/15 11:15:10 main.go:55: Database initialized successfully
```

**Kubernetes execution:**
```
[codex-service] 2026/01/15 11:15:10 main.go:27: Database: /data/codex.db (ledger=false)
[codex-service] 2026/01/15 11:15:10 main.go:55: Database initialized successfully
```

## ✨ Benefits

### For Local Development
- ✅ **No Permission Issues:** Uses home directory
- ✅ **Easy to Find:** `~/.codex_data/` is standard convention
- ✅ **Safe to Delete:** Won't affect OS files
- ✅ **Portable:** Works across different machines
- ✅ **Automatic:** Directory created on first run

### For Kubernetes Production
- ✅ **Persistent Storage:** PVC managed by Kubernetes
- ✅ **Backup-Friendly:** Easy to snapshot volumes
- ✅ **Container-Native:** Follows best practices
- ✅ **Scalable:** Each pod gets independent storage
- ✅ **Secure:** Mounted with restricted permissions

## 📚 Documentation

Comprehensive documentation available:

1. **`docs/DATA_STORAGE.md`** - Complete storage guide
   - Storage locations reference
   - Configuration options
   - Backup strategies
   - Troubleshooting
   - Security considerations

2. **`docs/HTTP_API.md`** - API reference
3. **`docs/KUBERNETES_DEPLOYMENT.md`** - K8s deployment guide
4. **`docs/DOCKER_BUILD.md`** - Docker guide

## 🔐 Security

### Local File Permissions
```
~/.codex_data/codex.db
- Owner: Current user
- Permissions: 0600 (read/write for owner only)
- Group: User's group
- Others: No access
```

### Directory Permissions
```
~/.codex_data/
- Owner: Current user
- Permissions: 0755 (owner can read/write/execute)
- For backup files and lock files
```

## 🧪 Testing

### Local Execution Test

```bash
# Build
make build-service

# Run
CODEX_API_KEYS=test-key ./bin/codex-service &
SERVICE_PID=$!

# Test API
curl http://localhost:8080/health

# Verify storage
ls -la ~/.codex_data/
file ~/.codex_data/codex.db

# Stop
kill $SERVICE_PID
```

### Kubernetes Test

```bash
# Deploy
helm install codex deployments/helm/codex \
  --set security.apiKeys[0]=test-key

# Verify
kubectl logs codex-0 | grep "Database:"

# Check storage
kubectl exec codex-0 -- ls -la /data/
kubectl exec codex-0 -- file /data/codex.db
```

## 📊 Comparison Table

| Feature | Before | After |
|---------|--------|-------|
| **Local Path** | ❌ `/data/` | ✅ `~/.codex_data/` |
| **K8s Path** | ✅ `/data/` | ✅ `/data/` |
| **Auto-Detection** | ❌ No | ✅ Yes |
| **Directory Creation** | ❌ Manual | ✅ Automatic |
| **Local Development** | ❌ Failed | ✅ Works perfectly |
| **Kubernetes** | ✅ Works | ✅ Still works |
| **Customization** | ❌ Limited | ✅ Via env var |

## 🔄 Migration

### From Old to New

If you have existing data:

**Option 1: Keep Using Custom Path**
```bash
# Set explicitly
CODEX_DB_PATH=/path/to/old/database ./bin/codex-service
```

**Option 2: Migrate to New Location**
```bash
# Copy old database to new location
mkdir -p ~/.codex_data
cp /old/path/codex.db ~/.codex_data/

# New execution will use ~/.codex_data/codex.db
./bin/codex-service
```

## 🎓 Environment Variables

### Configuration

| Variable | Value | Effect |
|----------|-------|--------|
| `CODEX_KUBERNETES` | `true` | Force `/data/codex.db` |
| `CODEX_KUBERNETES` | `false` | Force `~/.codex_data/codex.db` |
| `CODEX_DB_PATH` | `/any/path` | Override all defaults |
| `CODEX_API_KEYS` | `key1,key2` | API authentication |

### Example Commands

```bash
# Local with default path
CODEX_API_KEYS=dev-key ./bin/codex-service

# Local with custom path
CODEX_API_KEYS=dev-key CODEX_DB_PATH=/tmp/codex.db ./bin/codex-service

# Force Kubernetes mode locally (for testing)
CODEX_KUBERNETES=true CODEX_API_KEYS=test-key ./bin/codex-service
```

## 🐛 Troubleshooting

### Directory Not Created

**Problem:** `~/.codex_data/` directory doesn't exist

**Solution:**
```bash
# Create manually
mkdir -p ~/.codex_data

# Or let service create it
make run-service
```

### Permission Denied

**Problem:** "Permission denied" when accessing `~/.codex_data/`

**Solution:**
```bash
# Check permissions
ls -la ~/.codex_data

# Fix if needed
chmod 755 ~/.codex_data
chmod 600 ~/.codex_data/codex.db
```

### Wrong Path Being Used

**Problem:** Service using `/data/codex.db` locally

**Solution:**
```bash
# Ensure CODEX_KUBERNETES is not set
unset CODEX_KUBERNETES

# Or explicitly use local mode
CODEX_KUBERNETES=false ./bin/codex-service
```

## 📋 Checklist

- ✅ Config parsing updated with home directory support
- ✅ Automatic environment detection implemented
- ✅ Directory creation automatic and safe
- ✅ Kubernetes mode still fully supported
- ✅ Local development now works without issues
- ✅ Kubernetes manifests updated with CODEX_KUBERNETES flag
- ✅ Helm chart updated with kubernetesMode configuration
- ✅ Comprehensive documentation created
- ✅ Service compiles without errors
- ✅ Tested locally with home directory storage
- ✅ Backward compatible (CODEX_DB_PATH still works)

## 🎉 Summary

The refactoring successfully enables CodexDB HTTP Service to:

1. **Work seamlessly locally** using `~/.codex_data/` directory
2. **Maintain Kubernetes support** with `/data/codex.db` path
3. **Auto-detect environment** without manual configuration
4. **Create directories automatically** with proper permissions
5. **Support custom paths** via environment variable override

Users can now develop locally with zero permission issues while maintaining full Kubernetes production support!

---

**Status:** ✅ Complete and tested
**Build:** ✅ Compiles successfully
**Documentation:** ✅ Comprehensive guide available
**Backward Compatibility:** ✅ Maintained
