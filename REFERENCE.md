# Kubebuilder Quick Reference

## Project Setup Commands

```bash
# Initialize new project
kubebuilder init --domain example.com --repo github.com/user/project-name

# Create new API and controller
kubebuilder create api --group mygroup --version v1alpha1 --kind MyKind --resource --controller

# Create webhook
kubebuilder create webhook --group mygroup --version v1alpha1 --kind MyKind --defaulting --programmatic-validation
```

## Development Workflow

```bash
# 1. Edit your types (api/v1alpha1/*_types.go)
vim api/v1alpha1/myresource_types.go

# 2. Generate code and manifests
make manifests generate

# 3. Install CRDs
make install

# 4. Run operator locally
make run

# 5. Test with sample
kubectl apply -f config/samples/
```

## Common Make Targets

| Command | Description |
|---------|-------------|
| `make manifests` | Generate CRDs, RBAC, webhooks |
| `make generate` | Generate DeepCopy code |
| `make install` | Install CRDs to cluster |
| `make uninstall` | Remove CRDs from cluster |
| `make run` | Run operator locally |
| `make docker-build IMG=<img>` | Build container image |
| `make docker-push IMG=<img>` | Push container image |
| `make deploy IMG=<img>` | Deploy to cluster |
| `make undeploy` | Remove from cluster |
| `make test` | Run tests |

## Kubebuilder Markers

### CRD Markers (in types.go)

```go
// Validation
// +kubebuilder:validation:Required
// +kubebuilder:validation:Optional
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=100
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=50
// +kubebuilder:validation:Pattern=`^[a-z]+$`
// +kubebuilder:validation:Enum=value1;value2;value3

// Defaults
// +kubebuilder:default=defaultValue
// +kubebuilder:default=1

// Print columns (kubectl get output)
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Subresources
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas

// Resource options
// +kubebuilder:resource:shortName=res;rs
// +kubebuilder:resource:scope=Cluster
```

### RBAC Markers (in controller.go)

```go
// Basic RBAC
// +kubebuilder:rbac:groups=mygroup,resources=myresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mygroup,resources=myresources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mygroup,resources=myresources/finalizers,verbs=update

// Core resources
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Apps resources
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Cluster-scoped
// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
```

## Reconciliation Patterns

### Basic Reconcile Structure

```go
func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch the resource
    resource := &myapi.MyKind{}
    if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil  // Deleted - OK
        }
        return ctrl.Result{}, err  // Real error - requeue
    }

    // 2. Reconcile child resources
    // ... your logic ...

    // 3. Update status
    if err := r.Status().Update(ctx, resource); err != nil {
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}
```

### Create or Update Pattern

```go
func (r *MyReconciler) createOrUpdate(ctx context.Context, obj client.Object, owner *myapi.MyKind) error {
    // Set owner reference
    if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
        return err
    }

    // Try to get existing
    key := types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
    found := obj.DeepCopyObject().(client.Object)
    
    err := r.Get(ctx, key, found)
    if err != nil && errors.IsNotFound(err) {
        // Create
        return r.Create(ctx, obj)
    } else if err != nil {
        return err
    }

    // Update
    obj.SetResourceVersion(found.GetResourceVersion())
    return r.Update(ctx, obj)
}
```

### Finalizer Pattern

```go
const finalizerName = "mygroup.example.com/finalizer"

func (r *MyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    resource := &myapi.MyKind{}
    if err := r.Get(ctx, req.NamespacedName, resource); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Handle deletion
    if resource.ObjectMeta.DeletionTimestamp.IsZero() {
        // Not being deleted - add finalizer
        if !controllerutil.ContainsFinalizer(resource, finalizerName) {
            controllerutil.AddFinalizer(resource, finalizerName)
            return ctrl.Result{}, r.Update(ctx, resource)
        }
    } else {
        // Being deleted - run cleanup
        if controllerutil.ContainsFinalizer(resource, finalizerName) {
            if err := r.cleanup(ctx, resource); err != nil {
                return ctrl.Result{}, err
            }
            
            controllerutil.RemoveFinalizer(resource, finalizerName)
            return ctrl.Result{}, r.Update(ctx, resource)
        }
        return ctrl.Result{}, nil
    }

    // Normal reconciliation
    // ...
}
```

### Status Update Pattern

```go
func (r *MyReconciler) updateStatus(ctx context.Context, resource *myapi.MyKind) error {
    // Update status fields
    resource.Status.Phase = "Ready"
    resource.Status.ObservedGeneration = resource.Generation
    
    // Update or append condition
    condition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        ObservedGeneration: resource.Generation,
        LastTransitionTime: metav1.Now(),
        Reason:             "ReconciliationSucceeded",
        Message:            "Resource is ready",
    }
    
    meta.SetStatusCondition(&resource.Status.Conditions, condition)
    
    return r.Status().Update(ctx, resource)
}
```

## Client Usage

```go
// Get a resource
obj := &corev1.Pod{}
err := r.Get(ctx, types.NamespacedName{Name: "name", Namespace: "ns"}, obj)

// List resources
list := &corev1.PodList{}
err := r.List(ctx, list, client.InNamespace("ns"), client.MatchingLabels{"app": "myapp"})

// Create a resource
obj := &corev1.Pod{...}
err := r.Create(ctx, obj)

// Update a resource
err := r.Update(ctx, obj)

// Delete a resource
err := r.Delete(ctx, obj)

// Patch a resource
patch := client.MergeFrom(oldObj.DeepCopy())
// ... modify obj ...
err := r.Patch(ctx, obj, patch)

// Update status
err := r.Status().Update(ctx, obj)
```

## SetupWithManager Patterns

```go
// Watch primary resource
ctrl.NewControllerManagedBy(mgr).
    For(&myapi.MyKind{}).
    Complete(r)

// Watch owned resources
ctrl.NewControllerManagedBy(mgr).
    For(&myapi.MyKind{}).
    Owns(&appsv1.Deployment{}).
    Owns(&corev1.Service{}).
    Complete(r)

// Watch arbitrary resources
ctrl.NewControllerManagedBy(mgr).
    For(&myapi.MyKind{}).
    Watches(
        &source.Kind{Type: &corev1.ConfigMap{}},
        handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
    ).
    Complete(r)

// With predicates (filters)
ctrl.NewControllerManagedBy(mgr).
    For(&myapi.MyKind{}).
    WithEventFilter(predicate.GenerationChangedPredicate{}).
    Complete(r)
```

## Requeue Behaviors

```go
// Success - no requeue
return ctrl.Result{}, nil

// Requeue immediately
return ctrl.Result{Requeue: true}, nil

// Requeue after delay
return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil

// Error - exponential backoff requeue
return ctrl.Result{}, fmt.Errorf("something failed")

// Ignore not found errors
return ctrl.Result{}, client.IgnoreNotFound(err)
```

## Testing

### Unit Test Template

```go
func TestMyReconciler(t *testing.T) {
    scheme := runtime.NewScheme()
    _ = myapi.AddToScheme(scheme)
    _ = clientgoscheme.AddToScheme(scheme)
    
    client := fake.NewClientBuilder().
        WithScheme(scheme).
        WithObjects(/* initial objects */).
        Build()
    
    r := &MyReconciler{
        Client: client,
        Scheme: scheme,
    }
    
    req := ctrl.Request{
        NamespacedName: types.NamespacedName{
            Name:      "test",
            Namespace: "default",
        },
    }
    
    result, err := r.Reconcile(context.Background(), req)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    
    // Assert result
    // Assert resources created/updated
}
```

## Common Imports

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // Kubernetes core types
    corev1 "k8s.io/api/core/v1"
    appsv1 "k8s.io/api/apps/v1"
    
    // Kubernetes errors
    "k8s.io/apimachinery/pkg/api/errors"
    
    // Kubernetes meta types
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    
    // Controller runtime
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
    "sigs.k8s.io/controller-runtime/pkg/log"
    
    // Your API
    myapi "github.com/user/project/api/v1alpha1"
)
```

## Debugging Tips

```bash
# View operator logs
kubectl logs -n <namespace> deployment/<operator-name> -f

# Describe CRD
kubectl describe crd <resource-plural>.<group>

# View generated RBAC
kubectl describe clusterrole <operator-name>-manager-role

# Check operator pod
kubectl get pods -n <operator-namespace>
kubectl describe pod -n <operator-namespace> <pod-name>

# Check webhooks (if using)
kubectl get validatingwebhookconfigurations
kubectl get mutatingwebhookconfigurations

# View events
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

## Useful Commands

```bash
# Explain CRD fields
kubectl explain myresource.spec
kubectl explain myresource.status

# Get with custom columns
kubectl get myresource -o custom-columns=NAME:.metadata.name,REPLICAS:.spec.replicas,STATUS:.status.phase

# Watch resources
kubectl get myresource -w

# Patch a resource
kubectl patch myresource myname --type=merge -p '{"spec":{"field":"value"}}'

# Edit resource
kubectl edit myresource myname
```
