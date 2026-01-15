# CodexDB Kubernetes Deployment Guide

## Overview

This guide covers deploying CodexDB HTTP Service to a Kubernetes cluster. We provide both raw YAML manifests and a production-ready Helm chart.

### Port Auto-Detection

The CodexDB HTTP Service automatically detects and uses the first available port from this list:
- **Port 111** (preferred)
- **Port 1111** (if 111 is unavailable)
- **Port 11111** (if 1111 is unavailable)
- **Port 8080** (fallback)

You can override the port detection by setting the `CODEX_PORT` environment variable in the ConfigMap or deployment manifest.

## Prerequisites

### Required Tools

- **kubectl** (v1.20+) - Kubernetes command-line tool
- **helm** (v3.0+) - Kubernetes package manager (if using Helm)
- **docker** - For building container images
- A **Kubernetes cluster** (1.20+)

### Supported Kubernetes Distributions

- **Docker Desktop** - Includes Kubernetes
- **Minikube** - Local development cluster
- **Kind** - Lightweight testing cluster
- **EKS** (AWS) - Production cluster
- **GKE** (Google Cloud) - Production cluster
- **AKS** (Azure) - Production cluster
- **Any standard Kubernetes cluster**

### Storage Requirements

- **PersistentVolume (PV)** - For database persistence
- **PersistentVolumeClaim (PVC)** - Automatically created by StatefulSet

## Quick Start with Helm (Recommended)

### 1. Prepare the Helm Chart

```bash
# Navigate to the Helm chart directory
cd deployments/helm/codex
```

### 2. Create API Keys

Generate secure API keys for authentication:

```bash
# Generate API keys (base64-encoded random strings)
API_KEY_1=$(openssl rand -base64 32)
API_KEY_2=$(openssl rand -base64 32)

echo "API Key 1: $API_KEY_1"
echo "API Key 2: $API_KEY_2"
```

### 3. Generate Encryption Key (Optional)

For encrypted database:

```bash
# Generate 32-byte hex-encoded key for AES-256
ENCRYPTION_KEY=$(openssl rand -hex 32)

echo "Encryption Key: $ENCRYPTION_KEY"
```

### 4. Create values file for your deployment

Create `values-custom.yaml`:

```yaml
replicaCount: 1

image:
  repository: codex-service
  tag: latest

persistence:
  enabled: true
  size: 10Gi
  storageClass: standard  # Use 'gp2' for AWS, 'standard' for others

database:
  compression: gzip

security:
  apiKeys:
    - "YOUR_API_KEY_1_HERE"
    - "YOUR_API_KEY_2_HERE"
  encryptionKey: "YOUR_ENCRYPTION_KEY_HERE"  # or leave empty for no encryption

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 5. Install with Helm

```bash
# Add the chart repo (or use local path)
helm install codex ./codex -f values-custom.yaml

# Or upgrade if already installed
helm upgrade codex ./codex -f values-custom.yaml
```

### 6. Verify Installation

```bash
# Check the StatefulSet
kubectl get statefulsets

# Check the pods
kubectl get pods -l app.kubernetes.io/name=codex

# Check the service
kubectl get svc

# Check the PVC (storage)
kubectl get pvc

# View the pod logs
kubectl logs codex-0

# Test the service
kubectl port-forward svc/codex-service 8080:80
```

## Deployment with Raw YAML Manifests

### 1. Build the Docker Image

```bash
# Build from the Dockerfile
docker build -t codex-service:latest -f deployments/docker/Dockerfile .

# Tag for your registry
docker tag codex-service:latest myregistry/codex-service:1.0.0

# Push to registry (if using remote registry)
docker push myregistry/codex-service:1.0.0
```

### 2. Create Namespace (Optional)

```bash
kubectl create namespace codex
kubectl config set-context --current --namespace=codex
```

### 3. Create Secrets

```bash
# Generate API keys
API_KEYS="key1,key2,key3"
ENCRYPTION_KEY=$(openssl rand -hex 32)

# Create the secret
kubectl create secret generic codex-secrets \
  --from-literal=CODEX_API_KEYS="$API_KEYS" \
  --from-literal=CODEX_ENCRYPTION_KEY="$ENCRYPTION_KEY"

# Verify
kubectl get secret codex-secrets
kubectl describe secret codex-secrets
```

### 4. Apply ConfigMap

```bash
kubectl apply -f deployments/kubernetes/configmap.yaml
```

### 5. Apply StatefulSet and Service

```bash
# Create the service
kubectl apply -f deployments/kubernetes/service.yaml

# Create the StatefulSet
kubectl apply -f deployments/kubernetes/statefulset.yaml

# Watch the deployment
kubectl get statefulsets -w
kubectl get pods -w
```

### 6. Verify Deployment

```bash
# Check all resources
kubectl get all -l app=codex

# Check the StatefulSet
kubectl describe statefulset codex

# Check the pod
kubectl describe pod codex-0

# Check the PVC
kubectl get pvc

# View logs
kubectl logs codex-0
```

## Configuration Management

### Using ConfigMap

Edit `deployments/kubernetes/configmap.yaml` before applying:

```yaml
data:
  CODEX_PORT: "8080"
  CODEX_LEDGER_MODE: "false"
  CODEX_COMPRESSION: "gzip"
  CODEX_LOG_LEVEL: "info"
```

### Using Secrets

For sensitive data (API keys, encryption keys):

```bash
# Create secret from command line
kubectl create secret generic codex-secrets \
  --from-literal=CODEX_API_KEYS="key1,key2" \
  --from-literal=CODEX_ENCRYPTION_KEY="0123456789abcdef..."

# Or from a file
echo "key1,key2" > api_keys.txt
kubectl create secret generic codex-secrets \
  --from-file=CODEX_API_KEYS=api_keys.txt
```

### Update Configuration

```bash
# Edit ConfigMap
kubectl edit configmap codex-config

# Edit Secret
kubectl edit secret codex-secrets

# Restart pod to apply changes
kubectl delete pod codex-0
```

## Accessing the Service

### Port Forwarding (Development)

```bash
# Forward local port to service
kubectl port-forward svc/codex-service 8080:80

# Access locally
curl http://localhost:8080/health
curl -X GET http://localhost:8080/api/v1/keys \
  -H "X-API-Key: your-key"
```

### Service DNS (From Within Cluster)

Pods within the cluster can access the service at:

```
http://codex-service:80
http://codex-service.default:80
http://codex-service.default.svc.cluster.local:80
```

### LoadBalancer (External Access)

To expose the service externally:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: codex-service
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: codex
```

Then access via the external IP:

```bash
kubectl get svc codex-service
# Use the EXTERNAL-IP
```

### Ingress (Production)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codex-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - codex.example.com
    secretName: codex-tls
  rules:
  - host: codex.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: codex-service
            port:
              number: 80
```

## Monitoring & Troubleshooting

### View Logs

```bash
# Current logs
kubectl logs codex-0

# Follow logs (tail)
kubectl logs -f codex-0

# Previous logs (if pod restarted)
kubectl logs -p codex-0

# Logs from all replicas
kubectl logs -l app=codex --all-containers=true
```

### Pod Status

```bash
# Check pod status
kubectl get pods codex-0

# Detailed pod info
kubectl describe pod codex-0

# Get pod events
kubectl get events --field-selector involvedObject.name=codex-0
```

### Check Health Probes

```bash
# Execute health check from pod
kubectl exec codex-0 -- wget -q -O - http://localhost:8080/health

# Check probe status in pod
kubectl describe pod codex-0 | grep -A 5 "Liveness\|Readiness"
```

### Verify Storage

```bash
# Check PVC
kubectl get pvc

# Check PV
kubectl get pv

# Verify database file exists
kubectl exec codex-0 -- ls -la /data/

# Check disk usage
kubectl exec codex-0 -- du -sh /data/
```

### Database Issues

```bash
# Access pod shell
kubectl exec -it codex-0 -- /bin/sh

# Inside the pod:
ls -la /data/
file /data/codex.db
du -h /data/codex.db
```

## Scaling

### Horizontal Scaling

```bash
# Scale to 3 replicas
kubectl scale statefulset codex --replicas=3

# Watch scaling
kubectl get statefulsets -w

# Each replica gets its own PVC and database
kubectl get pvc
```

**Important:** Each replica has its own independent database. For shared data, implement application-level coordination.

### Vertical Scaling

Increase resource limits:

```yaml
resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

Or with Helm:

```bash
helm upgrade codex ./codex \
  --set resources.requests.cpu=200m \
  --set resources.limits.cpu=1000m
```

## Persistence & Backup

### Volume Snapshots

Create a backup snapshot:

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: codex-snapshot
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: codex-data-codex-0
```

### Backup CronJob

Automated backups to S3:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: codex-backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: amazon/aws-cli:latest
            command:
            - /bin/sh
            - -c
            - |
              tar czf backup.tar.gz /data/
              aws s3 cp backup.tar.gz s3://my-backups/codex-$(date +%Y%m%d).tar.gz
            volumeMounts:
            - name: codex-data
              mountPath: /data
              readOnly: true
          volumes:
          - name: codex-data
            persistentVolumeClaim:
              claimName: codex-data-codex-0
          restartPolicy: OnFailure
```

## Upgrades & Rolling Updates

### Using Helm

```bash
# Update image version
helm upgrade codex ./codex --set image.tag=v1.1.0

# Rollback if needed
helm rollback codex

# Check history
helm history codex
```

### Manual Update

```bash
# Edit StatefulSet to update image
kubectl edit statefulset codex

# Or patch directly
kubectl patch statefulset codex \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"codex-service","image":"codex-service:v1.1.0"}]}}}}'
```

Kubernetes will automatically roll out the update one pod at a time.

## Cleanup

### Delete Release (Helm)

```bash
# Delete all resources
helm uninstall codex

# PVC is retained by default (data persistence)
# To also delete PVC:
kubectl delete pvc --all -l app.kubernetes.io/instance=codex
```

### Delete Resources (Raw YAML)

```bash
kubectl delete statefulset codex
kubectl delete svc codex-service
kubectl delete configmap codex-config
kubectl delete secret codex-secrets

# Delete persistent volume claims
kubectl delete pvc codex-data-codex-0
```

## Production Checklist

- [ ] Use production-grade container registry
- [ ] Set proper resource requests and limits
- [ ] Configure PersistentVolume with high-performance storage
- [ ] Use Kubernetes Secrets for API keys and encryption keys
- [ ] Never commit secrets to git
- [ ] Enable pod security policies
- [ ] Configure network policies to restrict traffic
- [ ] Set up monitoring and logging
- [ ] Implement backup strategy
- [ ] Use Ingress with TLS certificate
- [ ] Configure pod disruption budget
- [ ] Test disaster recovery procedures
- [ ] Monitor disk space usage
- [ ] Plan capacity based on data growth

## Security Hardening

### Pod Security Policy

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: codex-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'MustRunAsNonRoot'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: false
```

### Network Policy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: codex-network-policy
spec:
  podSelector:
    matchLabels:
      app: codex
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector: {}  # Allow from any pod in the cluster
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector: {}
```

## Performance Tuning

### Resource Optimization

For small deployments:
```yaml
resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 100m
    memory: 128Mi
```

For high-traffic deployments:
```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

### Compression

Enable compression for faster network transfers:
```yaml
CODEX_COMPRESSION: "gzip"  # or zstd for max compression
CODEX_COMPRESSION_LEVEL: "6"
```

### Ledger Mode

Use ledger mode for audit trails (slightly slower but O(1) writes):
```yaml
CODEX_LEDGER_MODE: "true"
```

## Troubleshooting Common Issues

### Pod won't start

```bash
# Check events
kubectl describe pod codex-0

# Check logs
kubectl logs codex-0

# Check resource availability
kubectl top nodes
kubectl top pods
```

### Cannot access service

```bash
# Check service exists
kubectl get svc codex-service

# Check pod is running
kubectl get pods

# Test connectivity
kubectl exec -it codex-0 -- wget -q -O - http://localhost:8080/health
```

### Storage issues

```bash
# Check PVC
kubectl get pvc

# Check PV
kubectl get pv

# Check disk space
kubectl exec codex-0 -- df -h /data
```

### High memory usage

```bash
# Check memory usage
kubectl top pod codex-0

# Increase memory limit
kubectl set resources statefulset codex --limits=memory=1Gi
```

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [StatefulSet Guide](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Helm Documentation](https://helm.sh/docs/)
