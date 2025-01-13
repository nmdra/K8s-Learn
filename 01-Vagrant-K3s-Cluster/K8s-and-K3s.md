# Pod Scheduling

By default, **Kubernetes does not schedule pods on master nodes** unless explicitly configured to do so. This behavior is enforced using the **`node-role.kubernetes.io/master`** or **`node-role.kubernetes.io/control-plane`** taints applied to the control plane nodes.

## Key Details:

- **Control Plane Nodes**: In a Kubernetes cluster, the control plane (master) nodes are reserved for managing the cluster's control plane components (e.g., `kube-apiserver`, `etcd`, `kube-controller-manager`, and `kube-scheduler`) to ensure stability and performance.
- **Taints**: Kubernetes applies a taint to control plane nodes to prevent pods from being scheduled on them:

## Overriding This Behavior:

To allow pods to run on control plane nodes:

1. **Remove the Taint**: Remove the taint using the `kubectl taint` command:
```bash
kubectl taint nodes <node-name> node-role.kubernetes.io/master:NoSchedule-
```

2. **Add a Toleration**: Modify the pod's deployment YAML to include a toleration, so it can tolerate the master node's taint:

```yaml
  spec:
    tolerations:
    - key: "node-role.kubernetes.io/master"
        operator: "Exists"
          effect: "NoSchedule"
```

3. **Set `nodeSelector` or `affinity`**: To specifically schedule pods on the master node, use a `nodeSelector` or `nodeAffinity`:

```yaml
spec:
    nodeSelector:
    node-role.kubernetes.io/master: ""
```

## On Lightweight Clusters (e.g., K3s):

- In K3s, master nodes are also used as worker nodes by default because it is designed to work with minimal resources. Unless you configure dedicated worker nodes or taints manually, pods can run on the K3s master node.

> [!IMPORTANT]
> On K3s by default, all servers are also agents.

### Replicate default K8s behavior

> [!WARNING]
> System pods (like `kube-system` components) may need to run on the control plane node. You can add tolerations to their deployments.

```bash
kubectl taint nodes <master-node-name> node-role.kubernetes.io/control-plane:NoSchedule
```


[Architecture K3s](https://docs.k3s.io/architecture)
