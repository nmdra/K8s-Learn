# Kubernetes Admission Policies

CEL-based, in-process alternatives to admission webhooks. No external service, no TLS, or network round-trips.

## Folders

- **[Validating Policies](./validating/)** — Replica limit validation (static and parameterized).
- **[Mutating Policies](./mutating/)** — Auto-inject labels and priority classes.
- **[Manifest-Based Policies](./manifest-based/)** — Load policies from disk at API server startup (alpha, v1.36).

---

## Run & Test

### 1. Validating — Static Replica Limit (≤ 5)

```bash
# Setup
kubectl label namespace default environment=test
kubectl apply -f validating/demo-policy.yaml -f validating/demo-binding.yaml

# Test — should be DENIED
kubectl create deployment bad-deploy --image=nginx --replicas=10

# Test — should SUCCEED
kubectl create deployment good-deploy --image=nginx --replicas=3

# Cleanup
kubectl delete deployment good-deploy --ignore-not-found
kubectl delete -f validating/demo-binding.yaml -f validating/demo-policy.yaml
```

### 2. Validating — Parameterized Replica Limit

```bash
# Setup — CRD, policy, param, and binding
kubectl label namespace default environment=production
kubectl apply -f validating/replicalimit-crd.yaml
kubectl apply -f validating/replicalimit-policy.yaml
kubectl apply -f validating/prod-param.yaml -f validating/prod-binding.yaml

# Test — should be DENIED (max 10 for production)
kubectl create deployment over-limit --image=nginx --replicas=15

# Test — should SUCCEED
kubectl create deployment within-limit --image=nginx --replicas=8

# Cleanup
kubectl delete deployment within-limit --ignore-not-found
kubectl delete -f validating/prod-binding.yaml -f validating/prod-param.yaml
kubectl delete -f validating/replicalimit-policy.yaml -f validating/replicalimit-crd.yaml
```

### 3. Mutating — Auto-inject `environment: dev` Label

```bash
# Setup
kubectl apply -f mutating/default-labels-policy.yaml -f mutating/default-labels-binding.yaml

# Test — pod should get environment=dev label
kubectl run test-pod --image=nginx
kubectl get pod test-pod -o jsonpath='{.metadata.labels.environment}'  # → dev

# Cleanup
kubectl delete pod test-pod
kubectl delete -f mutating/default-labels-binding.yaml -f mutating/default-labels-policy.yaml
```

### 4. Mutating — Inject Priority Class from ConfigMap

```bash
# Setup
kubectl apply -f mutating/priority-config.yaml
kubectl apply -f mutating/inject-priority-policy.yaml -f mutating/inject-priority-binding.yaml

# Test — pod should get priorityClassName=medium-priority
kubectl run priority-test --image=nginx
kubectl get pod priority-test -o jsonpath='{.spec.priorityClassName}'  # → medium-priority

# Cleanup
kubectl delete pod priority-test
kubectl delete -f mutating/inject-priority-binding.yaml -f mutating/inject-priority-policy.yaml
kubectl delete -f mutating/priority-config.yaml
```

---

## Manifest-Based Admission Control (Alpha, v1.36)

Loads admission policies from **static files on disk** instead of from the Kubernetes API. Policies are active from API server startup, survive etcd unavailability, and **cannot be deleted** via the API.

> **Feature gate**: `ManifestBasedAdmissionControlConfig` (alpha, disabled by default)

### Why?

| Problem | How manifest-based solves it |
|---------|------------------------------|
| **Bootstrap gap** — policies don't exist until someone creates them | Policies are loaded from disk before the API serves requests |
| **Self-protection gap** — privileged users can delete admission configs | Manifest-based policies aren't API objects, they can't be modified/deleted |
| **etcd dependency** — if etcd is down, policies may not load | Policies are read from local files, independent of etcd |

### Key Rules

- Names **must** end with `.static.k8s.io` (e.g., `deny-privileged.static.k8s.io`)
- `spec.paramKind` / `spec.paramRef` are **not allowed** (no API resource references)
- Webhooks **must** use `clientConfig.url` (not `clientConfig.service`)
- Each plugin gets its own `staticManifestsDir` directory
- Files must be `.yaml`, `.yml`, or `.json`

### Setup with Kind

```bash
# 1. Create the policy files directory
mkdir -p admission/validating-policies

# 2. Copy policy files
cp manifest-based/deny-privileged-policy.yaml admission/validating-policies/
cp manifest-based/admission-config.yaml admission/

# 3. Create Kind cluster with the feature gate and volume mounts
kind create cluster --config manifest-based/kind-config.yaml

# 4. Test — should be DENIED (privileged container)
kubectl run bad-pod --image=nginx --overrides='{"spec":{"containers":[{"name":"nginx","image":"nginx","securityContext":{"privileged":true}}]}}'

# 5. Test — should SUCCEED (non-privileged container)
kubectl run good-pod --image=nginx
```

### Files

| File | Description |
|------|-------------|
| `admission-config.yaml` | AdmissionConfiguration pointing to static manifest directories |
| `deny-privileged-policy.yaml` | Denies privileged containers (except in `kube-system`) |
| `protect-admission-resources.yaml` | Prevents deletion/modification of labeled admission resources |
| `kind-config.yaml` | Kind cluster config with feature gate and volume mounts |

### References

- [Blog: Admission Policies That Can't Be Deleted](https://kubernetes.io/blog/2026/05/04/kubernetes-v1-36-manifest-based-admission-control/)
- [Docs: Manifest-Based Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/manifest-admission-control/)
- [Feature Gate: ManifestBasedAdmissionControlConfig](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/#ManifestBasedAdmissionControlConfig)
