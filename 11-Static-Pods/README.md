### Copy Manifests

```bash
k run static-busybox --image=busybox -n kube-system --dry-run=client -o yaml --command -- sleep 1000 >> busybox-kube-system.yaml
```

```bash
docker cp busybox-kube-system.yaml multi-control-plane:/etc/kubernetes/manifests
```


