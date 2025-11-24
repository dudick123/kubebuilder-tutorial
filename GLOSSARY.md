# Kubebuilder and Operator Glossary

## Core Concepts

### Operator
A software extension to Kubernetes that uses custom resources to manage applications and their components. Operators encode operational knowledge (how to deploy, scale, backup, recover, etc.) into software.

**Example:** A PostgreSQL operator knows how to set up replication, perform backups, and handle failovers.

### Custom Resource (CR)
An instance of a Custom Resource Definition. It's an object stored in the Kubernetes API that extends Kubernetes beyond built-in resources.

**Example:**
```yaml
apiVersion: webapp.example.com/v1alpha1
kind: GuestBook
metadata:
  name: my-guestbook
```

### Custom Resource Definition (CRD)
A schema that defines a new resource type in Kubernetes. It extends the Kubernetes API with your own types.

**Example:** The `GuestBook` CRD defines what fields a GuestBook resource can have.

### Controller
A control loop that watches the state of your cluster through the API server and makes changes to bring the current state closer to the desired state.

### Reconciliation
The process of bringing the actual state of the system in line with the desired state defined in custom resources. The reconcile function is called whenever a watched resource changes.

### Reconcile Loop
The continuous cycle of: Watch → Detect Change → Read Desired State → Read Actual State → Make Changes → Update Status → Repeat

## Kubebuilder Terms

### Kubebuilder
A framework and SDK for building Kubernetes APIs using CRDs. It scaffolds projects and generates boilerplate code.

### Manager
The central component that coordinates all controllers, handles leader election, serves metrics, and manages the controller runtime lifecycle.

### Scheme
A registry that maps Go types to Kubernetes GroupVersionKinds (GVK). It's how controller-runtime knows how to serialize/deserialize objects.

**Example:** Maps `&corev1.Pod{}` to `v1/Pod`

### GroupVersionKind (GVK)
- **Group:** API group (e.g., `apps`, `batch`, your custom group)
- **Version:** API version (e.g., `v1`, `v1beta1`, `v1alpha1`)
- **Kind:** Resource type (e.g., `Deployment`, `Pod`, `GuestBook`)

**Example:** `apps/v1/Deployment`

### Markers
Special comments that start with `+` used to generate code and configuration.

**Examples:**
```go
// +kubebuilder:validation:Minimum=1
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list
```

## Resource Structure

### Spec
The desired state of a resource, as defined by the user. Users fill out the spec when they create/update a resource.

**Example:**
```yaml
spec:
  replicas: 3
  welcomeMessage: "Hello!"
```

### Status
The observed state of a resource, updated by the controller. Users should never edit status.

**Example:**
```yaml
status:
  availableReplicas: 3
  conditions:
    - type: Ready
      status: "True"
```

### Metadata
Standard Kubernetes metadata like name, namespace, labels, annotations. Inherited from `metav1.ObjectMeta`.

## Patterns

### Owner Reference
A field in a resource's metadata that links it to a parent resource. Used for garbage collection.

**Effect:** When parent is deleted, owned children are automatically deleted.

### Garbage Collection
Kubernetes' automatic cleanup of child resources when their owner is deleted, based on owner references.

### Finalizer
A key in a resource's metadata that prevents deletion until the controller removes it. Used to run cleanup logic before deletion.

**Example:** Backup data before deleting a database.

### Condition
A status field that describes a particular aspect of a resource's state.

**Example:**
```go
Condition{
    Type:   "Ready",
    Status: "True",
    Reason: "AllPodsReady",
}
```

### Idempotency
The property that a function can be called multiple times with the same result. Reconcile functions must be idempotent.

**Example:** Creating a resource that already exists should succeed (or be a no-op), not error.

## Controller Runtime

### Client
Interface for reading and writing Kubernetes resources. Provides `Get`, `List`, `Create`, `Update`, `Delete`, `Patch`.

### Watcher
Monitors Kubernetes resources for changes and triggers reconciliation.

### Queue
Holds reconciliation requests. Controller runtime manages the queue, adding items when resources change.

### Workqueue
The actual implementation of the queue. Supports rate limiting, retries, and deduplication.

### Predicates
Filters that determine whether an event should trigger reconciliation.

**Example:** Only reconcile when generation changes (not just metadata updates).

## RBAC

### Role-Based Access Control (RBAC)
Kubernetes' authorization system. Controllers need permissions to read and write resources.

### ClusterRole
Cluster-wide permissions. Required for cluster-scoped resources or cross-namespace operations.

### Role
Namespace-scoped permissions. Used for operations within a single namespace.

### ServiceAccount
Identity for pods. Your operator runs as a ServiceAccount with specific permissions.

### RoleBinding / ClusterRoleBinding
Links a Role/ClusterRole to a ServiceAccount, granting permissions.

## Testing

### EnvTest
Test framework that runs a real Kubernetes API server for integration testing.

### Fake Client
In-memory client for unit testing. Doesn't require a real cluster.

### Mock
Simulated object that mimics real behavior for testing.

## Deployment

### Kustomize
Tool for customizing Kubernetes configurations. Kubebuilder uses it for managing deployment manifests.

### Webhook
HTTP callbacks that intercept API requests. Used for validation, mutation, and conversion.

### Leader Election
Mechanism to ensure only one controller instance is active when running multiple replicas for high availability.

## Advanced Concepts

### Admission Webhook
Intercepts requests to create/update resources for validation or mutation.

**Types:**
- **Validating:** Rejects invalid requests
- **Mutating:** Modifies requests (e.g., inject sidecar, set defaults)

### Conversion Webhook
Converts between different versions of your API (e.g., v1alpha1 ↔ v1beta1).

### Subresource
A special resource endpoint like `/status` or `/scale`. Separate from the main resource.

### Watch
Long-running connection to the API server that streams changes to resources.

### Informer
Caches resources locally and watches for changes. Reduces API server load.

### Lister
Provides read-only access to cached resources from an Informer.

## Common Abbreviations

- **CR:** Custom Resource
- **CRD:** Custom Resource Definition
- **GVK:** GroupVersionKind
- **RBAC:** Role-Based Access Control
- **API:** Application Programming Interface
- **HA:** High Availability
- **SA:** ServiceAccount

## Error Types

### IsNotFound
Error indicating a resource doesn't exist. Often not a real error in reconciliation.

```go
if errors.IsNotFound(err) {
    // Resource was deleted - that's OK
    return ctrl.Result{}, nil
}
```

### IsConflict
Error indicating a resource was modified by another client. Usually requires retry.

### IsInvalid
Error indicating request validation failed.

## Return Values

### ctrl.Result{}
No requeue, reconciliation complete successfully.

### ctrl.Result{Requeue: true}
Requeue immediately.

### ctrl.Result{RequeueAfter: duration}
Requeue after specified time.

### (ctrl.Result{}, error)
Error occurred, will requeue with exponential backoff.

## Common Mistakes

### Editing Status in Update()
❌ Wrong: `r.Update(ctx, resource)`
✅ Correct: `r.Status().Update(ctx, resource)`

### Not Checking IsNotFound
❌ Wrong: Return error when resource not found during reconcile
✅ Correct: Return `nil` error - deletion is normal

### Forgetting Owner References
❌ Wrong: Create child resources without owner references
✅ Correct: Use `ctrl.SetControllerReference()`

### Non-Idempotent Reconcile
❌ Wrong: Logic that fails or has side effects when called multiple times
✅ Correct: Check if resource exists before creating

### Infinite Reconciliation Loops
❌ Wrong: Updating main resource in Reconcile (triggers another reconcile)
✅ Correct: Only update status, or use proper conditions to avoid loops

## Useful Commands

### Generation vs ResourceVersion
- **Generation:** Increments when spec changes (meaningful changes)
- **ResourceVersion:** Changes on any modification (including status/metadata)

Use Generation in status.observedGeneration to track which spec you've reconciled.

---

## Need More Definitions?

This glossary covers the most common terms you'll encounter. For deeper dives:

- [Kubernetes Glossary](https://kubernetes.io/docs/reference/glossary/)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller Runtime GoDoc](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
