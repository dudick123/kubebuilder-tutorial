# Kubebuilder Tutorial for Beginners

A comprehensive, step-by-step tutorial for learning Kubebuilder and Kubernetes operator development from scratch.

## 📚 Tutorial Structure

This tutorial is divided into three parts:

### [Part 1: Introduction and Setup](TUTORIAL-PART1.md)
- What is Kubebuilder and why use it?
- Understanding operators and the reconciliation loop
- Project setup and initialization
- Understanding the generated code structure
- Creating your first API

**Topics covered:**
- Installing prerequisites
- `kubebuilder init` and project structure
- `kubebuilder create api` command
- Understanding Spec, Status, and Metadata
- Kubebuilder markers and code generation

### [Part 2: Implementing the Controller](TUTORIAL-PART2.md)
- Writing actual reconciliation logic
- Managing child resources (Deployments, Services, ConfigMaps)
- Understanding owner references and garbage collection
- Status updates and conditions
- RBAC and permissions

**Topics covered:**
- The Reconcile function flow
- Create-or-Update pattern
- Setting controller references
- Watching owned resources
- Error handling and requeuing

### [Part 3: Testing and Deployment](TUTORIAL-PART3.md)
- Testing locally with `make run`
- Creating sample resources
- Verifying self-healing behavior
- Building and deploying to cluster
- Troubleshooting common issues

**Topics covered:**
- Local development workflow
- Installing CRDs
- Testing reconciliation
- Building Docker images
- Deploying with `make deploy`
- Viewing logs and debugging

### [Quick Reference](REFERENCE.md)
- Common commands cheat sheet
- Kubebuilder markers reference
- Reconciliation patterns
- Client usage examples
- Testing templates

## 🎯 What You'll Build

A **GuestBook operator** that:
- Manages a simple guestbook web application
- Creates and maintains Deployments, Services, and ConfigMaps
- Handles scaling (1-10 replicas)
- Updates configuration dynamically
- Self-heals when resources are deleted
- Reports status accurately

## 🚀 Quick Start

### Prerequisites
```bash
# Required tools
go version    # 1.21+
docker version
kubectl version
kubebuilder version
```

### Complete Tutorial (Estimated time: 2-3 hours)

**For first-time learners:**
1. Read [TUTORIAL-PART1.md](TUTORIAL-PART1.md) - Understanding the basics
2. Follow [TUTORIAL-PART2.md](TUTORIAL-PART2.md) - Write the controller
3. Practice with [TUTORIAL-PART3.md](TUTORIAL-PART3.md) - Test and deploy
4. Keep [REFERENCE.md](REFERENCE.md) handy for quick lookups

**For experienced Go developers new to Kubebuilder:**
- Start with Part 2 for the controller implementation details
- Reference Part 1 for Kubebuilder-specific concepts
- Jump to Part 3 for testing and deployment

## 📦 Example Files Included

| File | Description |
|------|-------------|
| `guestbook_types.go` | Complete CRD definition |
| `guestbook_controller.go` | Full controller implementation |
| `sample-guestbook.yaml` | Example resource to test with |
| `REFERENCE.md` | Quick reference for common patterns |

## 🎓 Learning Objectives

By the end of this tutorial, you will:

- ✅ Understand what operators are and why they're useful
- ✅ Know how to use Kubebuilder to scaffold projects
- ✅ Be able to define Custom Resource Definitions (CRDs)
- ✅ Implement reconciliation loops correctly
- ✅ Manage child resources with owner references
- ✅ Handle resource updates and deletions
- ✅ Write status conditions properly
- ✅ Test operators locally and in-cluster
- ✅ Debug operator issues effectively
- ✅ Deploy operators to production clusters

## 🔑 Key Concepts Explained

### Operators
Software extensions to Kubernetes that use custom resources to manage applications and their components. Think of them as automated SREs.

### Reconciliation Loop
The core pattern of operators: continuously bringing actual state to match desired state. Called whenever resources change or periodically.

### Custom Resource Definition (CRD)
Extends Kubernetes API with new resource types. Your operator watches these custom resources and manages related standard resources.

### Owner References
Links child resources to parent resources. When parent is deleted, children are automatically cleaned up by Kubernetes garbage collection.

### Status Subresource
Separate from spec, status reflects the observed state. Controllers update status, users update spec. This separation prevents infinite loops.

## 💡 Why This Tutorial?

**Beginner-friendly:**
- Assumes no prior Kubebuilder knowledge
- Explains every concept before using it
- Progressive complexity
- Real, working example

**Comprehensive:**
- Covers full development lifecycle
- Includes testing and debugging
- Production-ready patterns
- Best practices explained

**Practical:**
- Builds a complete, functional operator
- Copy-paste ready code
- Common pitfalls highlighted
- Troubleshooting guide included

## 📖 Tutorial Philosophy

This tutorial follows these principles:

1. **Explain Why, Not Just How** - Understanding concepts, not just following steps
2. **Progressive Disclosure** - Introduce complexity gradually
3. **Working Example** - Build something real you can run
4. **Best Practices** - Learn the right way from the start
5. **Hands-On** - Type the code, see it work, understand it

## 🛠️ What You'll Need

**Knowledge Prerequisites:**
- Basic Go programming (variables, functions, structs)
- Basic Kubernetes concepts (Pods, Deployments, Services)
- Command line comfort
- YAML familiarity

**Don't need:**
- Deep Go expertise (we explain Go patterns)
- Kubernetes internals knowledge
- Prior operator experience
- DevOps background

## 🚦 How to Use This Tutorial

### If you're new to operators:
Start at Part 1 and work through sequentially. Don't skip ahead! Each part builds on previous knowledge.

### If you know operators but not Kubebuilder:
- Skim Part 1 for Kubebuilder-specific info
- Focus on Part 2's code generation and markers
- Review Part 3's development workflow

### If you're stuck:
1. Check the troubleshooting sections in each part
2. Review the REFERENCE.md for patterns
3. Ensure you ran `make manifests generate`
4. Check operator logs
5. Verify RBAC permissions

## 📝 Getting Help

**Common issues:**
- Import path errors → Update go.mod and imports
- CRD not found → Run `make install`
- RBAC errors → Check `+kubebuilder:rbac` markers
- Resource not reconciling → Check operator logs

**Next steps after tutorial:**
- Add webhooks for validation
- Implement finalizers for cleanup logic
- Add unit and integration tests
- Explore advanced patterns in controller-runtime
- Build your own operator for a real use case

## 🎯 Success Criteria

You've completed the tutorial when you can:
- [ ] Scaffold a new Kubebuilder project
- [ ] Define a CRD with validation
- [ ] Implement a reconciliation loop
- [ ] Create and manage child resources
- [ ] Update status conditions
- [ ] Test locally with `make run`
- [ ] Deploy to a cluster
- [ ] Debug operator issues

## 📚 Additional Resources

**Official Documentation:**
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes API Concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)

**Community:**
- [Kubernetes Slack #kubebuilder](https://kubernetes.slack.com/messages/kubebuilder)
- [Controller Runtime GitHub](https://github.com/kubernetes-sigs/controller-runtime)
- [Kubebuilder GitHub](https://github.com/kubernetes-sigs/kubebuilder)

**Examples:**
- [Kubebuilder Examples](https://github.com/kubernetes-sigs/kubebuilder/tree/master/docs/book/src/cronjob-tutorial/testdata/project)
- [Operator SDK Examples](https://github.com/operator-framework/operator-sdk/tree/master/testdata)

---

## Let's Get Started!

Ready to build your first operator? Start with [Part 1: Introduction and Setup](TUTORIAL-PART1.md)!

**Estimated completion time:**
- Part 1: 45 minutes (reading + setup)
- Part 2: 60 minutes (coding)
- Part 3: 45 minutes (testing + deployment)
- **Total: 2.5-3 hours**

Take breaks between parts - this is a lot of new information!

---

## Feedback

Found an error? Have suggestions? This tutorial is designed to help you learn effectively. Your feedback makes it better for the next person!
