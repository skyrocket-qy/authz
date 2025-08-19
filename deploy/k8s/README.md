kubectl create namespace authz


kafka
kubectl create -f 'https://strimzi.io/install/latest?namespace=authz' -n authz
kubectl apply -f https://strimzi.io/examples/latest/kafka/kafka-single-node.yaml -n authz

wait until ready
kubectl wait authz/my-cluster --for=condition=Ready --timeout=300s -n authz 

kubectl -n authz run kafka-producer -ti --image=quay.io/strimzi/kafka:0.47.0-kafka-4.0.0 --rm=true --restart=Never -- bin/kafka-console-producer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic my-topic

> Hello Strimzi!

kubectl -n authz run kafka-consumer -ti --image=quay.io/strimzi/kafka:0.47.0-kafka-4.0.0 --rm=true --restart=Never -- bin/kafka-console-consumer.sh --bootstrap-server my-cluster-kafka-bootstrap:9092 --topic my-topic --from-beginning

delete
kubectl -n authz delete $(kubectl get strimzi -o name -n authz)
kubectl delete pvc -l strimzi.io/name=my-cluster-kafka -n authz
kubectl -n authz delete -f 'https://strimzi.io/install/latest?namespace=authz'


connecter
```

```
