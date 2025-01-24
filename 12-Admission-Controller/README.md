> [!IMPORTANT]
> This example assume you use kind k8s cluster..   
> [kind Cluster](../02-Kind-Cluster-Config/)   
> [Static Pods](../11-Static-Pods/)

---

1. copy manifest file from control plane node
```bash
docker cp multi-control-plane:/etc/kubernetes/manifests/kube-apiserver.yaml .
```

2. Edit manifest file to enable admission controller.

```yaml
spec:
  containers:
  - command:
    - kube-apiserver
    - --client-ca-file=/etc/kubernetes/pki/ca.crt
    - --enable-admission-plugins=NodeRestriction,NamespaceAutoProvision,LimitRanger # Manualy Updated
    - --enable-bootstrap-token-auth=true
```

3. Copy Updated manifest file to control plane node
```bash
docker cp kube-apiserver.yaml multi-control-plane:/etc/kubernetes/manifests/kube-apiserver.yaml
```

4. Check if the Plugins Are Enabled
```bash
kubectl get pod -n kube-system -l component=kube-apiserver -o yaml | grep enable-admission-plugins
```
5. Test the `NamespaceAutoProvision` plugin

```bash
kubectl create deployment test-deployment --image=nginx:alpine -n test-namespace

```
