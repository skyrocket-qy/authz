https://github.com/grafana/helm-charts/tree/main/charts/loki-distributed
```
kubectl create ns loki
helm -n loki install loki grafana/loki-distributed
```