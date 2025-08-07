### Generate configmap using `kubectl`

```bash 

kubectl create configmap app-config --from-file=./demo-app/config.json
```

We can check generated configmap using following command.

```bash
kubectl get configmap app-config -o yaml
```


--- 




