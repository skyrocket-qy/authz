# Flow

```zsh
minikube start --insecure-registry "10.0.0.0/24" --cpus=8 --memory=16g 
minikube addons enable registry

kubectl create ns authz

# Deploying OLM
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.20.0/install.sh | bash -s v0.20.0

sleep 5
kubectl apply -f global-operators.yaml 

# Deploying Strimzi Operator
kubectl apply -f subscription.yaml

# Creating Secrets, role for the Database
kubectl apply -f debezium_conf.yaml

# Create kafka

kubectl apply -f kafka.yaml

# Wait untail ready

kubectl wait kafka/debezium-cluster --for=condition=Ready --timeout=300s -n authz

# db

kubectl apply -f postgres.yaml
TODO: need to wait

# KafkaConnect
kubectl -n kube-system get svc registry -o jsonpath='{.spec.clusterIP}' # get ip

kubectl apply -f kafka-connect.yaml
TODO: need to wait

# debezium connector
kubectl apply -f debezium_connector.yaml

TODO: need to wait
kubectl describe KafkaConnector/debezium-connector-postgres

# watch
kubectl run -n authz -it --rm --image=quay.io/debezium/tooling:1.2  --restart=Never watcher -- kcat -b debezium-cluster-kafka-bootstrap:9092 -C -o beginning -t mysql.inventory.customers
```
