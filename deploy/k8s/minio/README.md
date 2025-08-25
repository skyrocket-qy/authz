https://github.com/minio/operator/blob/master/README.md
```
kubectl create namespace minio-tenant

kubectl kustomize github.com/minio/operator\?ref=v7.1.1 | kubectl apply -f -
kubectl get pods -n minio-operator

kubectl apply -k github.com/minio/operator/examples/kustomization/base
```