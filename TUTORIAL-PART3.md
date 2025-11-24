# Kubebuilder Tutorial Part 3: Testing and Deployment

## Testing Your Operator

### Step 12: Generate and Install CRDs

```bash
# Generate CRD manifests and code
make manifests generate

# Install CRDs into your cluster
make install
```

**What `make install` does:**
```bash
kubectl apply -f config/crd/bases/
```

**Verify CRD is installed:**
```bash
kubectl get crd guestbooks.webapp.example.com

# View the CRD details
kubectl describe crd guestbooks.webapp.example.com

# Check what fields are available
kubectl explain guestbook.spec
kubectl explain guestbook.status
```

### Step 13: Run the Operator Locally

```bash
# Run the operator on your local machine
make run
```

**What happens:**
```
2024-11-24T00:00:00.000Z    INFO    controller-runtime.metrics    Metrics server is starting to listen    {"addr": ":8080"}
2024-11-24T00:00:00.000Z    INFO    setup    starting manager
2024-11-24T00:00:00.000Z    INFO    Starting server    {"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
2024-11-24T00:00:00.000Z    INFO    Starting EventSource    {"controller": "guestbook", "source": "kind source: *v1alpha1.GuestBook"}
2024-11-24T00:00:00.000Z    INFO    Starting Controller    {"controller": "guestbook"}
2024-11-24T00:00:00.000Z    INFO    Starting workers    {"controller": "guestbook", "worker count": 1}
```

**Understanding the logs:**
- **Metrics server**: Prometheus metrics on :8080
- **EventSource**: Watching for GuestBook resources
- **Controller**: Your reconciliation loop is running
- **Workers**: Number of concurrent reconciliation threads

**Leave this running** - it's your operator!

### Step 14: Create a Sample GuestBook

In a **new terminal**, create a sample resource:

```bash
# Create a namespace for testing
kubectl create namespace guestbook-test

# Apply the sample
kubectl apply -f - <<EOF
apiVersion: webapp.example.com/v1alpha1
kind: GuestBook
metadata:
  name: my-guestbook
  namespace: guestbook-test
spec:
  replicas: 2
  welcomeMessage: "Hello from Kubebuilder!"
EOF
```

**In the operator terminal, you'll see:**
```
INFO    Reconciling GuestBook    {"name": "my-guestbook"}
INFO    Creating resource    {"kind": "ConfigMap", "name": "my-guestbook-config"}
INFO    Creating resource    {"kind": "Deployment", "name": "my-guestbook"}
INFO    Creating resource    {"kind": "Service", "name": "my-guestbook-service"}
INFO    Reconciliation complete
```

### Step 15: Verify the Resources Were Created

```bash
# Check the GuestBook resource
kubectl get guestbook -n guestbook-test

# Should show:
# NAME            REPLICAS   AVAILABLE   AGE
# my-guestbook    2          2           30s

# View detailed info
kubectl describe guestbook my-guestbook -n guestbook-test

# Check the created resources
kubectl get all -n guestbook-test

# Should show:
# - deployment.apps/my-guestbook
# - service/my-guestbook-service
# - replicaset.apps/my-guestbook-xxx
# - pod/my-guestbook-xxx-yyy (2 of them)

# Check the ConfigMap
kubectl get configmap my-guestbook-config -n guestbook-test -o yaml
```

### Step 16: Test Self-Healing

**Delete the Deployment and watch it get recreated:**

```bash
# Delete the deployment
kubectl delete deployment my-guestbook -n guestbook-test

# Immediately check - it should be recreated!
kubectl get deployment my-guestbook -n guestbook-test
```

**In the operator logs:**
```
INFO    Reconciling GuestBook    {"name": "my-guestbook"}
INFO    Creating resource    {"kind": "Deployment", "name": "my-guestbook"}
INFO    Reconciliation complete
```

**Why this works:**
- We used `Owns(&appsv1.Deployment{})` in SetupWithManager
- When the Deployment is deleted, operator gets notified
- Reconcile runs and recreates the missing resource
- This is **self-healing** behavior!

### Step 17: Test Updates

**Update the GuestBook spec:**

```bash
kubectl patch guestbook my-guestbook -n guestbook-test \
  --type='merge' \
  -p '{"spec":{"replicas":3}}'
```

**Watch the operator update the Deployment:**

```bash
# Check the deployment replicas
kubectl get deployment my-guestbook -n guestbook-test

# Should show 3 replicas now
```

**In the operator logs:**
```
INFO    Reconciling GuestBook    {"name": "my-guestbook"}
INFO    Updating resource    {"kind": "Deployment", "name": "my-guestbook"}
INFO    Reconciliation complete
```

### Step 18: Check the Status

```bash
# View the status subresource
kubectl get guestbook my-guestbook -n guestbook-test -o jsonpath='{.status}' | jq

# Should show:
{
  "availableReplicas": 3,
  "url": "http://my-guestbook-service.guestbook-test.svc.cluster.local",
  "conditions": [
    {
      "type": "Ready",
      "status": "True",
      "reason": "DeploymentReady",
      "message": "3/3 replicas available"
    }
  ]
}
```

### Step 19: Test Deletion

```bash
# Delete the GuestBook
kubectl delete guestbook my-guestbook -n guestbook-test

# Check that all resources are cleaned up
kubectl get all -n guestbook-test
# Should be empty!
```

**Why everything is deleted:**
- We used `SetControllerReference` for all child resources
- Kubernetes garbage collection handles cleanup
- Owner reference creates parent-child relationship

---

## Deploying to Kubernetes

Now let's deploy the operator **inside** the cluster instead of running it locally.

### Step 20: Build the Docker Image

```bash
# Build and tag the image
make docker-build IMG=<your-registry>/guestbook-operator:v0.1.0

# Examples:
# make docker-build IMG=docker.io/myuser/guestbook-operator:v0.1.0
# make docker-build IMG=ghcr.io/myorg/guestbook-operator:v0.1.0
```

**What this does:**
- Builds a multi-stage Docker image
- Stage 1: Compiles the Go binary
- Stage 2: Creates a minimal distroless image with just the binary

**Dockerfile breakdown:**
```dockerfile
# Stage 1: Build
FROM golang:1.21 as builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o manager cmd/main.go

# Stage 2: Runtime
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532
ENTRYPOINT ["/manager"]
```

### Step 21: Push the Image

```bash
# Push to your registry
make docker-push IMG=<your-registry>/guestbook-operator:v0.1.0
```

**Note:** You need to be logged in to your registry:
```bash
# Docker Hub
docker login

# GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin

# Azure Container Registry
az acr login --name myregistry
```

### Step 22: Deploy to Cluster

```bash
# Deploy the operator
make deploy IMG=<your-registry>/guestbook-operator:v0.1.0
```

**What this does:**
```bash
# Applies kustomize configuration from config/default
kubectl apply -k config/default
```

**Resources created:**
```
namespace/guestbook-operator-system
customresourcedefinition.apiextensions.k8s.io/guestbooks.webapp.example.com
serviceaccount/guestbook-operator-controller-manager
role.rbac.authorization.k8s.io/guestbook-operator-leader-election-role
clusterrole.rbac.authorization.k8s.io/guestbook-operator-manager-role
rolebinding.rbac.authorization.k8s.io/guestbook-operator-leader-election-rolebinding
clusterrolebinding.rbac.authorization.k8s.io/guestbook-operator-manager-rolebinding
deployment.apps/guestbook-operator-controller-manager
```

### Step 23: Verify Deployment

```bash
# Check operator pod
kubectl get pods -n guestbook-operator-system

# View operator logs
kubectl logs -n guestbook-operator-system \
  deployment/guestbook-operator-controller-manager \
  -f

# Should see:
# INFO    Starting server    {"kind": "health probe", "addr": "[::]:8081"}
# INFO    Starting server    {"path": "/metrics", "kind": "metrics", "addr": "127.0.0.1:8080"}
# INFO    Starting EventSource    {"controller": "guestbook"}
# INFO    Starting Controller    {"controller": "guestbook"}
```

### Step 24: Test the Deployed Operator

```bash
# Create a GuestBook (operator is now running in-cluster)
kubectl apply -f config/samples/webapp_v1alpha1_guestbook.yaml

# Watch it get created
kubectl get guestbook -A -w

# Check the logs
kubectl logs -n guestbook-operator-system \
  deployment/guestbook-operator-controller-manager \
  -f
```

---

## Understanding the Directory Structure

Let's understand what all these files do:

```
guestbook-operator/
├── cmd/
│   └── main.go                 # Entry point - starts manager
│
├── api/v1alpha1/
│   ├── guestbook_types.go      # CRD definition (Spec & Status)
│   ├── groupversion_info.go    # API group metadata
│   └── zz_generated.deepcopy.go # Generated DeepCopy methods
│
├── internal/controller/
│   └── guestbook_controller.go # Reconciliation logic
│
├── config/
│   ├── crd/
│   │   └── bases/              # Generated CRD YAML
│   ├── default/
│   │   └── kustomization.yaml  # Main kustomize config
│   ├── manager/
│   │   ├── manager.yaml        # Deployment for operator
│   │   └── kustomization.yaml
│   ├── rbac/
│   │   ├── role.yaml           # Generated RBAC
│   │   ├── role_binding.yaml
│   │   └── service_account.yaml
│   ├── samples/
│   │   └── webapp_v1alpha1_guestbook.yaml # Sample CR
│   └── ...
│
├── Dockerfile                   # Multi-stage build
├── Makefile                    # Build, test, deploy commands
├── PROJECT                     # Project metadata
├── go.mod                      # Go dependencies
└── go.sum                      # Dependency checksums
```

---

## Common Makefile Targets

```bash
# Development
make manifests          # Generate CRDs and RBAC
make generate          # Generate DeepCopy code
make fmt               # Format code
make vet               # Run go vet
make test              # Run tests

# Local testing
make install           # Install CRDs
make uninstall         # Remove CRDs
make run               # Run operator locally

# Building
make docker-build      # Build container image
make docker-push       # Push container image

# Deployment
make deploy            # Deploy to cluster
make undeploy          # Remove from cluster

# Cleanup
make clean             # Clean build artifacts
```

---

## Troubleshooting

### Issue: CRD not found

**Error:**
```
Error from server (NotFound): error when creating "config/samples/webapp_v1alpha1_guestbook.yaml": 
the server could not find the requested resource (get guestbooks.webapp.example.com)
```

**Solution:**
```bash
# Install CRDs
make install

# Verify
kubectl get crd guestbooks.webapp.example.com
```

### Issue: Operator not reconciling

**Check:**
1. Is operator running?
   ```bash
   kubectl get pods -n guestbook-operator-system
   kubectl logs -n guestbook-operator-system deployment/guestbook-operator-controller-manager
   ```

2. RBAC permissions correct?
   ```bash
   kubectl describe clusterrole guestbook-operator-manager-role
   ```

3. Resource in correct namespace?
   ```bash
   kubectl get guestbook -A
   ```

### Issue: Resources not being created

**Check operator logs:**
```bash
kubectl logs -n guestbook-operator-system deployment/guestbook-operator-controller-manager -f
```

**Common issues:**
- Missing RBAC permissions (add markers and run `make manifests`)
- Resource name conflicts
- Invalid resource spec

### Issue: Import path errors

**Error:**
```
package github.com/yourusername/guestbook-operator/api/v1alpha1 is not in GOROOT
```

**Solution:**
```bash
# Update module path in go.mod to match your repo
go mod edit -module=github.com/YOUR-USERNAME/guestbook-operator

# Update imports in code
# Replace all instances of the old path with your new path

# Download dependencies
go mod tidy
```

---

## Next Steps and Advanced Topics

### Add Webhooks (Validation/Mutation)

```bash
kubebuilder create webhook \
  --group webapp \
  --version v1alpha1 \
  --kind GuestBook \
  --defaulting \
  --programmatic-validation
```

### Add Metrics

Already included! Prometheus metrics available at `:8080/metrics`

```bash
# Port-forward to see metrics
kubectl port-forward -n guestbook-operator-system \
  deployment/guestbook-operator-controller-manager 8080:8080

# View metrics
curl http://localhost:8080/metrics
```

### Add Unit Tests

Create `internal/controller/guestbook_controller_test.go`:

```go
package controller

import (
    "context"
    "testing"
    
    webappv1alpha1 "github.com/yourusername/guestbook-operator/api/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGuestBookReconciler(t *testing.T) {
    // Create a fake client
    scheme := runtime.NewScheme()
    _ = webappv1alpha1.AddToScheme(scheme)
    
    client := fake.NewClientBuilder().
        WithScheme(scheme).
        Build()
    
    // Create reconciler
    r := &GuestBookReconciler{
        Client: client,
        Scheme: scheme,
    }
    
    // Create a GuestBook
    gb := &webappv1alpha1.GuestBook{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test",
            Namespace: "default",
        },
        Spec: webappv1alpha1.GuestBookSpec{
            Replicas: 2,
            WelcomeMessage: "Test",
        },
    }
    
    _ = client.Create(context.Background(), gb)
    
    // Test reconciliation
    // ... add your tests
}
```

Run tests:
```bash
make test
```

### Use EnvTest for Integration Tests

EnvTest runs a real API server for testing:

```bash
# Install envtest binaries
make envtest

# Run tests with envtest
make test
```

### Add Finalizers (for cleanup logic)

```go
const finalizerName = "webapp.example.com/finalizer"

// In Reconcile function
if guestbook.ObjectMeta.DeletionTimestamp.IsZero() {
    // Not being deleted, add finalizer
    if !controllerutil.ContainsFinalizer(guestbook, finalizerName) {
        controllerutil.AddFinalizer(guestbook, finalizerName)
        return r.Update(ctx, guestbook)
    }
} else {
    // Being deleted
    if controllerutil.ContainsFinalizer(guestbook, finalizerName) {
        // Perform cleanup
        if err := r.cleanup(ctx, guestbook); err != nil {
            return ctrl.Result{}, err
        }
        
        // Remove finalizer
        controllerutil.RemoveFinalizer(guestbook, finalizerName)
        return r.Update(ctx, guestbook)
    }
}
```

---

## Summary

**What you learned:**
1. ✅ How to scaffold a Kubebuilder project
2. ✅ How to define CRDs with Spec and Status
3. ✅ How to implement a reconciliation loop
4. ✅ How to manage child resources with owner references
5. ✅ How to test locally with `make run`
6. ✅ How to deploy to a cluster
7. ✅ How self-healing works with `Owns()`
8. ✅ How to update status subresources
9. ✅ How RBAC markers generate permissions
10. ✅ How to troubleshoot common issues

**Key concepts:**
- **Reconciliation**: Bring actual state to desired state
- **Idempotency**: Safe to call multiple times
- **Owner References**: Automatic garbage collection
- **Status Subresource**: Separate from spec, prevents loops
- **Controller-runtime**: Framework that handles watching, queueing, retries

**Your operator now:**
- ✅ Creates Deployments, Services, and ConfigMaps
- ✅ Self-heals when resources are deleted
- ✅ Updates resources when spec changes
- ✅ Reports status accurately
- ✅ Runs in-cluster or locally

---

*See REFERENCE.md for quick commands and patterns!*
