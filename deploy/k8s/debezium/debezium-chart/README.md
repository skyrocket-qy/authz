# Debezium Helm Chart

This Helm chart deploys a Debezium instance on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Strimzi Kafka Operator installed

## Installing the Chart

To install the chart with the release name `my-debezium`:

```bash
helm install my-debezium .
```

## Configuration

The following table lists the configurable parameters of the Debezium chart and their default values.

| Parameter | Description | Default |
| --- | --- | --- |
| `nameOverride` | Override the name of the chart | `""` |
| `fullnameOverride` | Override the full name of the chart | `""` |
| `operatorGroup.create` | If `true`, create an `OperatorGroup` resource | `true` |
| `operatorGroup.namespace` | Namespace for the `OperatorGroup` | `operators` |
| `subscription.create` | If `true`, create a `Subscription` resource for the Strimzi operator | `true` |
| `subscription.namespace` | Namespace for the `Subscription` | `operators` |
| `subscription.channel` | Channel for the `Subscription` | `stable` |
| `subscription.name` | Name of the operator to subscribe to | `strimzi-kafka-operator` |
| `subscription.source` | Source of the operator | `operatorhubio-catalog` |
| `subscription.sourceNamespace` | Source namespace of the operator | `olm` |
| `kafka.replicas` | Number of Kafka replicas | `1` |
| `kafka.version` | Kafka version | `4.0.0` |
| `kafka.metadataVersion` | Kafka metadata version | `4.0-IV3` |
| `kafka.storage.size` | Size of the Kafka storage | `20Gi` |
| `kafka.externalListener.type` | Type of the external listener | `nodeport` |
| `kafka.config.offsetsTopicReplicationFactor` | Replication factor for the offsets topic | `1` |
| `kafka.config.transactionStateLogReplicationFactor` | Replication factor for the transaction state log topic | `1` |
| `kafka.config.transactionStateLogMinIsr` | Minimum in-sync replicas for the transaction state log topic | `1` |
| `kafka.config.defaultReplicationFactor` | Default replication factor for topics | `1` |
| `kafka.config.minInsyncReplicas` | Minimum in-sync replicas | `1` |
| `kafkaConnect.version` | Kafka Connect version | `4.0.0` |
| `kafkaConnect.replicas` | Number of Kafka Connect replicas | `1` |
| `kafkaConnect.build.output.image` | Output image for the Kafka Connect build | `10.106.188.150/debezium-connect-postgres:latest` |
| `kafkaConnect.build.plugins[0].artifacts[0].url` | URL for the Debezium connector plugin | `https://repo1.maven.org/maven2/io/debezium/debezium-connector-postgres/3.2.0.Final/debezium-connector-postgres-3.2.0.Final-plugin.tar.gz` |
| `debezium.username` | Debezium database username | `postgres` |
| `debezium.password` | Debezium database password | `password` |
| `debezium.connector.tasksMax` | Maximum number of tasks for the Debezium connector | `1` |
| `debezium.database.hostname` | Database hostname | `postgres` |
| `debezium.database.port` | Database port | `5432` |
| `debezium.database.dbname` | Database name | `postgres` |
| `deim.database.serverName` | Database server name | `pgserver1` |
| `debezium.database.serverId` | Database server ID | `184054` |
| `debezium.database.includeList` | List of databases to include | `authz` |
| `debezium.publication.name` | Name of the publication | `debezium` |
| `debezium.slot.name` | Name of the slot | `debezium_slot` |
| `debezium.topic.prefix` | Prefix for the topics | `pg` |
| `debezium.table.includeList` | List of tables to include | `tuples` |
