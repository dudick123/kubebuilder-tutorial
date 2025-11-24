# Kubebuilder Tutorial Part 2: Implementing the Controller

## Implementing the Reconciliation Logic

Now we'll write the actual logic that makes our operator work!

### Step 10: Understanding Reconciliation

The `Reconcile` function is called whenever:
1. A GuestBook resource is created, updated, or deleted
2. A resource owned by GuestBook changes
3. Periodically (for resync)

**Key principles:**
- **Idempotent**: Can be called multiple times with same result
- **Edge-triggered**: Responds to changes
- **Level-triggered**: Brings actual state to desired state

### Step 11: Implement the Controller

Replace the contents of `internal/controller/guestbook_controller.go`:

```go
package controller

import (
    "context"
    "fmt"
    
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    webappv1alpha1 "github.com/yourusername/guestbook-operator/api/v1alpha1"
)

// GuestBookReconciler reconciles a GuestBook object
type GuestBookReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.example.com,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.example.com,resources=guestbooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.example.com,resources=guestbooks/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main reconciliation loop
func (r *GuestBookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch the GuestBook instance
    guestbook := &webappv1alpha1.GuestBook{}
    err := r.Get(ctx, req.NamespacedName, guestbook)
    if err != nil {
        if errors.IsNotFound(err) {
            // Object not found, could have been deleted
            log.Info("GuestBook resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue the request
        log.Error(err, "Failed to get GuestBook")
        return ctrl.Result{}, err
    }

    log.Info("Reconciling GuestBook", "name", guestbook.Name)

    // 2. Create or update the ConfigMap
    configMap := r.configMapForGuestBook(guestbook)
    if err := r.createOrUpdate(ctx, configMap, guestbook); err != nil {
        log.Error(err, "Failed to create/update ConfigMap")
        return ctrl.Result{}, err
    }

    // 3. Create or update the Deployment
    deployment := r.deploymentForGuestBook(guestbook)
    if err := r.createOrUpdate(ctx, deployment, guestbook); err != nil {
        log.Error(err, "Failed to create/update Deployment")
        return ctrl.Result{}, err
    }

    // 4. Create or update the Service
    service := r.serviceForGuestBook(guestbook)
    if err := r.createOrUpdate(ctx, service, guestbook); err != nil {
        log.Error(err, "Failed to create/update Service")
        return ctrl.Result{}, err
    }

    // 5. Update status
    if err := r.updateStatus(ctx, guestbook); err != nil {
        log.Error(err, "Failed to update GuestBook status")
        return ctrl.Result{}, err
    }

    log.Info("Reconciliation complete")
    return ctrl.Result{}, nil
}

// createOrUpdate creates or updates a Kubernetes resource
func (r *GuestBookReconciler) createOrUpdate(ctx context.Context, obj client.Object, owner *webappv1alpha1.GuestBook) error {
    log := log.FromContext(ctx)
    
    // Set GuestBook instance as the owner
    if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
        return err
    }

    // Try to get the existing object
    key := types.NamespacedName{
        Name:      obj.GetName(),
        Namespace: obj.GetNamespace(),
    }
    
    found := obj.DeepCopyObject().(client.Object)
    err := r.Get(ctx, key, found)
    
    if err != nil && errors.IsNotFound(err) {
        // Object doesn't exist, create it
        log.Info("Creating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
        return r.Create(ctx, obj)
    } else if err != nil {
        return err
    }

    // Object exists, update it
    log.Info("Updating resource", "kind", obj.GetObjectKind().GroupVersionKind().Kind, "name", obj.GetName())
    obj.SetResourceVersion(found.GetResourceVersion())
    return r.Update(ctx, obj)
}

// configMapForGuestBook creates a ConfigMap for the welcome message
func (r *GuestBookReconciler) configMapForGuestBook(gb *webappv1alpha1.GuestBook) *corev1.ConfigMap {
    return &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      gb.Name + "-config",
            Namespace: gb.Namespace,
            Labels:    labelsForGuestBook(gb.Name),
        },
        Data: map[string]string{
            "welcome.txt": gb.Spec.WelcomeMessage,
        },
    }
}

// deploymentForGuestBook creates a Deployment for the guestbook
func (r *GuestBookReconciler) deploymentForGuestBook(gb *webappv1alpha1.GuestBook) *appsv1.Deployment {
    replicas := gb.Spec.Replicas
    labels := labelsForGuestBook(gb.Name)

    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      gb.Name,
            Namespace: gb.Namespace,
            Labels:    labels,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: labels,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: labels,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "guestbook",
                            Image: "gcr.io/google-samples/gb-frontend:v4",
                            Ports: []corev1.ContainerPort{
                                {
                                    ContainerPort: 80,
                                    Name:          "http",
                                },
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "config",
                                    MountPath: "/config",
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "config",
                            VolumeSource: corev1.VolumeSource{
                                ConfigMap: &corev1.ConfigMapVolumeSource{
                                    LocalObjectReference: corev1.LocalObjectReference{
                                        Name: gb.Name + "-config",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

// serviceForGuestBook creates a Service for the guestbook
func (r *GuestBookReconciler) serviceForGuestBook(gb *webappv1alpha1.GuestBook) *corev1.Service {
    labels := labelsForGuestBook(gb.Name)

    return &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      gb.Name + "-service",
            Namespace: gb.Namespace,
            Labels:    labels,
        },
        Spec: corev1.ServiceSpec{
            Selector: labels,
            Ports: []corev1.ServicePort{
                {
                    Port:     80,
                    Protocol: corev1.ProtocolTCP,
                },
            },
            Type: corev1.ServiceTypeClusterIP,
        },
    }
}

// updateStatus updates the GuestBook status subresource
func (r *GuestBookReconciler) updateStatus(ctx context.Context, gb *webappv1alpha1.GuestBook) error {
    // Get the Deployment to check available replicas
    deployment := &appsv1.Deployment{}
    err := r.Get(ctx, types.NamespacedName{Name: gb.Name, Namespace: gb.Namespace}, deployment)
    if err != nil {
        return err
    }

    // Update status
    gb.Status.AvailableReplicas = deployment.Status.AvailableReplicas
    gb.Status.URL = fmt.Sprintf("http://%s-service.%s.svc.cluster.local", gb.Name, gb.Namespace)

    // Update condition
    condition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        ObservedGeneration: gb.Generation,
        LastTransitionTime: metav1.Now(),
        Reason:             "DeploymentReady",
        Message:            fmt.Sprintf("%d/%d replicas available", gb.Status.AvailableReplicas, gb.Spec.Replicas),
    }

    if gb.Status.AvailableReplicas < gb.Spec.Replicas {
        condition.Status = metav1.ConditionFalse
        condition.Reason = "DeploymentNotReady"
    }

    // Update or append condition
    updated := false
    for i, c := range gb.Status.Conditions {
        if c.Type == condition.Type {
            gb.Status.Conditions[i] = condition
            updated = true
            break
        }
    }
    if !updated {
        gb.Status.Conditions = append(gb.Status.Conditions, condition)
    }

    return r.Status().Update(ctx, gb)
}

// labelsForGuestBook returns the labels for a GuestBook resource
func labelsForGuestBook(name string) map[string]string {
    return map[string]string{
        "app":        "guestbook",
        "guestbook":  name,
        "managed-by": "guestbook-operator",
    }
}

// SetupWithManager sets up the controller with the Manager
func (r *GuestBookReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&webappv1alpha1.GuestBook{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&corev1.ConfigMap{}).
        Complete(r)
}
```

### Understanding the Code

#### 1. The Reconcile Function Flow

```go
func (r *GuestBookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
```

**Inputs:**
- `ctx`: Context for cancellation and timeouts
- `req`: Contains the name/namespace of the resource that triggered this call

**Outputs:**
- `ctrl.Result`: Controls requeue behavior
  - `ctrl.Result{}` - No requeue
  - `ctrl.Result{Requeue: true}` - Requeue immediately
  - `ctrl.Result{RequeueAfter: 5*time.Minute}` - Requeue after delay
- `error`: If non-nil, controller-runtime will requeue with exponential backoff

#### 2. Fetching the Resource

```go
guestbook := &webappv1alpha1.GuestBook{}
err := r.Get(ctx, req.NamespacedName, guestbook)
if err != nil {
    if errors.IsNotFound(err) {
        // Resource deleted - that's OK
        return ctrl.Result{}, nil
    }
    // Real error - requeue
    return ctrl.Result{}, err
}
```

**Why check IsNotFound?**
- Delete events may trigger reconciliation
- Resource might be deleted while reconciling
- Not an error - just means our work is done

#### 3. Owner References

```go
if err := ctrl.SetControllerReference(owner, obj, r.Scheme); err != nil {
    return err
}
```

**What this does:**
- Links the child resource to the parent
- When parent is deleted, children are automatically deleted (garbage collection)
- Kubernetes handles cleanup for you

#### 4. Create or Update Pattern

```go
found := obj.DeepCopyObject().(client.Object)
err := r.Get(ctx, key, found)

if err != nil && errors.IsNotFound(err) {
    return r.Create(ctx, obj)  // Doesn't exist - create it
}

// Exists - update it
obj.SetResourceVersion(found.GetResourceVersion())
return r.Update(ctx, obj)
```

**Why this pattern?**
- Kubernetes requires ResourceVersion for updates
- Create fails if resource exists
- Update fails if resource doesn't exist
- This handles both cases

#### 5. Status Updates

```go
return r.Status().Update(ctx, gb)
```

**Important:**
- Status is a separate subresource
- Use `.Status().Update()` not `.Update()`
- Prevents infinite reconciliation loops
- Status updates don't trigger reconciliation

#### 6. RBAC Markers

```go
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
```

**What this does:**
- Generates RBAC permissions in `config/rbac/role.yaml`
- Operator needs permission to manage resources
- Run `make manifests` to regenerate after adding

#### 7. Watching Owned Resources

```go
func (r *GuestBookReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&webappv1alpha1.GuestBook{}).    // Watch GuestBook CRs
        Owns(&appsv1.Deployment{}).           // Watch Deployments we own
        Owns(&corev1.Service{}).              // Watch Services we own
        Owns(&corev1.ConfigMap{}).            // Watch ConfigMaps we own
        Complete(r)
}
```

**What this does:**
- `For()` - Primary resource to watch
- `Owns()` - Child resources to watch
- When a child changes, reconcile is called for the parent
- Enables self-healing: if someone deletes the Deployment, operator recreates it

---

*Continued in TUTORIAL-PART3.md...*
