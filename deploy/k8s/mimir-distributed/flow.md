https://grafana.com/docs/helm-charts/mimir-distributed/latest/get-started-helm-charts/
```zsh
kubectl create ns mimir
helm -n mimir install mimir grafana/mimir-distributed
```