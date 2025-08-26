```zsh
kubectl create ns grafana
helm install grafana grafana/grafana --namespace grafana


# get password instruction
helm get notes grafana -n grafana
#fGrCOT4B2NjFJL9DDCqSW3EhgBSvE0qmWcnHuSgI
```