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
spec:
  targetNamespaces:
    - debezium-example
# Create kafka
kubectl edit operatorgroup global-operators -n operators

kubectl apply -f debezium_2.yaml

# Wait untail ready

kubectl wait kafka/debezium-cluster --for=condition=Ready --timeout=300s -n example
```
