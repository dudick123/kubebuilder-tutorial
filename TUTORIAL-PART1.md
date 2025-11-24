# Kubebuilder Tutorial: Building Your First Kubernetes Operator

## Table of Contents
1. [Introduction](#introduction)
2. [Prerequisites](#prerequisites)
3. [Understanding Operators](#understanding-operators)
4. [Project Setup](#project-setup)
5. [Creating the API](#creating-the-api)
6. [Understanding the Generated Code](#understanding-the-generated-code)
7. [Implementing the Controller](#implementing-the-controller)
8. [Testing Your Operator](#testing-your-operator)
9. [Deploying to Kubernetes](#deploying-to-kubernetes)
10. [Next Steps](#next-steps)

---

## Introduction

### What is Kubebuilder?

Kubebuilder is a framework for building Kubernetes APIs using Custom Resource Definitions (CRDs). It generates the scaffolding and boilerplate code needed to create Kubernetes operators, letting you focus on your business logic.

### What We'll Build

We'll create a simple **GuestBook** operator that manages a guestbook application. This will teach you:
- How to define Custom Resources (CRDs)
- How to write a reconciliation loop
- How to manage Kubernetes resources from an operator
- How operators respond to changes

### Why Learn This?

Understanding Kubebuilder helps you:
- Automate complex Kubernetes workflows
- Create self-healing systems
- Build platform abstractions for your teams
- Implement GitOps patterns effectively

---

## Prerequisites

Before starting, ensure you have:

### Required Tools
```bash
# Go 1.21 or later
go version

# Docker (for building images)
docker version

# kubectl configured with a cluster
kubectl version

# Kubebuilder 3.x or later
kubebuilder version
```

### Installing Kubebuilder

**On macOS:**
```bash
brew install kubebuilder
```

**On Linux:**
```bash
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/
```

### Kubernetes Cluster

You can use:
- **minikube**: `minikube start`
- **kind**: `kind create cluster`
- **Docker Desktop**: Enable Kubernetes
- **Any cloud K8s cluster** (EKS, AKS, GKE)

---

## Understanding Operators

### What is an Operator?

An operator is a Kubernetes controller that:
1. **Watches** for changes to custom resources
2. **Reconciles** the actual state with desired state
3. **Manages** related Kubernetes resources

### The Reconciliation Loop

```
User creates/updates CR
         ↓
Controller gets notified
         ↓
Controller reads desired state
         ↓
Controller reads actual state
         ↓
Controller reconciles (create/update/delete resources)
         ↓
Controller updates status
         ↓
(Loop continues as changes occur)
```

### Example Flow

**User creates this:**
```yaml
apiVersion: webapp.example.com/v1alpha1
kind: GuestBook
metadata:
  name: my-guestbook
spec:
  replicas: 3
  message: "Welcome!"
```

**Operator does this:**
1. Creates a Deployment with 3 replicas
2. Creates a Service
3. Creates a ConfigMap with the message
4. Updates the GuestBook status

---

## Project Setup

### Step 1: Create Project Directory

```bash
# Create a directory for your project
mkdir guestbook-operator
cd guestbook-operator
```

### Step 2: Initialize the Kubebuilder Project

```bash
# Initialize with your domain and repo
kubebuilder init \
  --domain example.com \
  --repo github.com/yourusername/guestbook-operator
```

**What this does:**
- Creates a Go module
- Sets up project structure
- Creates Makefile with common commands
- Creates Dockerfile for building operator image
- Sets up basic configuration files

**Key files created:**
```
guestbook-operator/
├── Dockerfile           # For building operator image
├── Makefile            # Build, test, deploy commands
├── PROJECT             # Project metadata
├── go.mod              # Go dependencies
├── go.sum              # Dependency checksums
├── cmd/
│   └── main.go         # Entry point
└── config/             # Kubernetes manifests
    ├── default/        # Default kustomize configs
    ├── manager/        # Deployment for operator
    ├── rbac/          # RBAC rules
    └── ...
```

### Step 3: Understanding main.go

Open `cmd/main.go`. This is your operator's entry point:

```go
package main

import (
    // Standard imports
    "flag"
    "os"
    
    // Kubernetes client libraries
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    
    // Your imports will be added here
)

func main() {
    var metricsAddr string
    var enableLeaderElection bool
    
    // Parse command-line flags
    flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", 
                   "The address the metric endpoint binds to.")
    flag.BoolVar(&enableLeaderElection, "leader-elect", false,
                 "Enable leader election for controller manager.")
    flag.Parse()

    // Create the controller manager
    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        Scheme:             scheme,
        MetricsBindAddress: metricsAddr,
        LeaderElection:     enableLeaderElection,
    })
    
    // Controllers will be registered here
    
    // Start the manager
    mgr.Start(ctrl.SetupSignalHandler())
}
```

**Key concepts:**
- **Manager**: Coordinates all controllers
- **Scheme**: Registry of Go types to Kubernetes API objects
- **Leader Election**: Only one operator instance is active (for HA)

---

## Creating the API

### Step 4: Generate API and Controller

```bash
# Create the GuestBook API
kubebuilder create api \
  --group webapp \
  --version v1alpha1 \
  --kind GuestBook \
  --resource \
  --controller
```

**What this does:**
- Creates the CRD Go types in `api/v1alpha1/`
- Creates the controller in `internal/controller/`
- Updates `cmd/main.go` to register the controller
- Generates RBAC permissions
- Creates sample manifests

**Output structure:**
```
api/v1alpha1/
├── guestbook_types.go      # Your CRD definition
└── groupversion_info.go    # API group metadata

internal/controller/
└── guestbook_controller.go # Your controller logic

config/
├── crd/
│   └── bases/              # Generated CRD YAML
├── samples/
│   └── webapp_v1alpha1_guestbook.yaml  # Sample CR
└── rbac/
    └── role.yaml           # Generated RBAC
```

---

## Understanding the Generated Code

### Step 5: Examine guestbook_types.go

Open `api/v1alpha1/guestbook_types.go`:

```go
package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE! Add your desired spec and status fields

// GuestBookSpec defines the desired state of GuestBook
type GuestBookSpec struct {
    // INSERT ADDITIONAL SPEC FIELDS
    // Important: Run "make" to regenerate code after modifying this file
    
    // Foo is an example field of GuestBook. Edit guestbook_types.go to remove/update
    Foo string `json:"foo,omitempty"`
}

// GuestBookStatus defines the observed state of GuestBook
type GuestBookStatus struct {
    // INSERT ADDITIONAL STATUS FIELD
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GuestBook is the Schema for the guestbooks API
type GuestBook struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   GuestBookSpec   `json:"spec,omitempty"`
    Status GuestBookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GuestBookList contains a list of GuestBook
type GuestBookList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []GuestBook `json:"items"`
}

func init() {
    SchemeBuilder.Register(&GuestBook{}, &GuestBookList{})
}
```

**Understanding the structure:**

1. **Spec** (GuestBookSpec): What the user wants
   - This is the desired state
   - User fills this out in the YAML

2. **Status** (GuestBookStatus): What actually exists
   - Controller updates this
   - Reflects the current state
   - Users should NOT edit this

3. **Metadata**: Standard K8s metadata
   - Name, namespace, labels, annotations
   - Provided by `metav1.ObjectMeta`

4. **Markers** (lines starting with `+kubebuilder:`):
   - Code generation directives
   - CRD generation instructions
   - `+kubebuilder:object:root=true` - Top-level API type
   - `+kubebuilder:subresource:status` - Enable status updates

### Step 6: Examine guestbook_controller.go

Open `internal/controller/guestbook_controller.go`:

```go
package controller

import (
    "context"
    
    "k8s.io/apimachinery/pkg/runtime"
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

// Reconcile is part of the main kubernetes reconciliation loop
func (r *GuestBookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    _ = log.FromContext(ctx)

    // TODO(user): Your logic here

    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestBookReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&webappv1alpha1.GuestBook{}).
        Complete(r)
}
```

**Understanding the controller:**

1. **GuestBookReconciler struct**:
   - `client.Client`: For reading/writing K8s resources
   - `Scheme`: Type information for serialization

2. **Reconcile function**:
   - Called when a GuestBook resource changes
   - Also called periodically
   - Should be idempotent (safe to call multiple times)
   - Returns `ctrl.Result` for requeue behavior

3. **SetupWithManager**:
   - Registers this controller with the manager
   - Specifies which resources to watch (`For()`)
   - Can watch owned resources (`Owns()`)

---

## Implementing the Controller

Now let's implement actual functionality!

### Step 7: Define the GuestBook Spec

Edit `api/v1alpha1/guestbook_types.go`:

```go
// GuestBookSpec defines the desired state of GuestBook
type GuestBookSpec struct {
    // Replicas is the number of guestbook instances
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=10
    // +kubebuilder:default=1
    Replicas int32 `json:"replicas,omitempty"`
    
    // WelcomeMessage is displayed on the guestbook page
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:default="Welcome to our Guestbook!"
    WelcomeMessage string `json:"welcomeMessage,omitempty"`
}

// GuestBookStatus defines the observed state of GuestBook
type GuestBookStatus struct {
    // AvailableReplicas is the number of running replicas
    AvailableReplicas int32 `json:"availableReplicas"`
    
    // URL is the service endpoint
    URL string `json:"url,omitempty"`
    
    // Conditions represent the latest observations of the GuestBook state
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

**Understanding the markers:**
- `+kubebuilder:validation:Minimum=1` - CRD validation
- `+kubebuilder:default=1` - Default value
- These become OpenAPI validation in the CRD

### Step 8: Add Printcolumns for kubectl

Add these markers above the `GuestBook` type:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableReplicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// GuestBook is the Schema for the guestbooks API
type GuestBook struct {
    // ... rest of the code
}
```

This makes `kubectl get guestbook` show useful columns!

### Step 9: Generate Code and Manifests

```bash
# Generate DeepCopy methods and CRD manifests
make manifests generate
```

**What this does:**
- Runs `controller-gen` to generate:
  - `zz_generated.deepcopy.go` - DeepCopy methods for your types
  - CRD YAML in `config/crd/bases/`
  - RBAC in `config/rbac/`

**Always run this after changing:**
- Spec or Status fields
- Kubebuilder markers
- RBAC requirements

---

*Continued in TUTORIAL-PART2.md...*
