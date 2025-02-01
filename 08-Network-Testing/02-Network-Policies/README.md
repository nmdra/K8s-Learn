## Commands

```bash
kubectl apply -f app.yaml
kubectl apply -f network-polily-1.yaml
```

### Check Network Policy 

1. `./network-policy-1.yaml`
```bash
kubectl run test-pod --rm -it --image=busybox -- /bin/sh

wget --spider --timeout=2 nginx
```

```bash
kubectl run allowed-pod --rm -it --image=busybox --labels="access=allowed" -- /bin/sh

wget --spider --timeout=2 nginx
```

2. `./network-policy-2.yaml`

```bash
kubectl run test-trusted --rm -it --image=busybox -n trusted -- /bin/sh

wget --spider --timeout=2 nginx.default
```
