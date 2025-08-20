# Flow

```zsh
minikube start --insecure-registry "10.0.0.0/24" --cpus=8 --memory=16g 
minikube addons enable registry

kubectl create ns authz

# Deploying Strimzi Operator
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.20.0/install.sh | bash -s v0.20.0
kubectl create -f https://operatorhub.io/install/strimzi-kafka-operator.yaml

# Creating Secrets, role for the Database
kubectl apply -f debezium_1.yaml

kubectl edit operatorgroup global-operators -n operators
spec:
  targetNamespaces:
    - authz
# Create kafka

kubectl apply -f debezium_3.yaml

# Wait untail ready

kubectl wait kafka/debezium-cluster --for=condition=Ready --timeout=300s -n example

# mysql

kubectl apply -f mysql.yaml

# KafkaConnect
kubectl -n kube-system get svc registry -o jsonpath='{.spec.clusterIP}' # get ip

kubectl apply -f kafka-connect.yaml

# debezium connector
kubectl apply -f debezium_connector.yaml

# watch
kubectl run -n authz -it --rm --image=quay.io/debezium/tooling:1.2  --restart=Never watcher -- kcat -b debezium-cluster-kafka-bootstrap:9092 -C -o beginning -t mysql.inventory.customers
```
