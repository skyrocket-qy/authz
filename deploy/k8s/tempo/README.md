```zsh
https://github.com/grafana/helm-charts/tree/main/charts/tempo-distributed
helm repo add grafana https://grafana.github.io/helm-charts

kubectl create ns grafana-agent-operator
helm install grafana-agent-operator grafana/grafana-agent-operator

kubectl create ns grafana-rollout-operator
helm install  -n grafana-rollout-operator grafana-rollout-operator grafana/rollout-operator

hekubectl create ns tempo
helm -n tempo install tempo grafana/tempo-distributed -f value.yaml 

# use to show all values
helm show values grafana/tempo-distributed

# to delete
helm delete tempo
```