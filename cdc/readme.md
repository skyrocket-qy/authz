# Connector setup

1. Start service

```zsh
docker compose up -d 
```

2. set db

```zsh
docker exec -it postgres psql -U postgres -d postgres
```

```sql
-- Example table
CREATE TABLE tuples ();

-- Publication for Debezium
CREATE PUBLICATION debezium FOR TABLE tuples;
```

3. Set connector

```zsh
curl -X PUT http://localhost:8083/connectors/postgres-connector/config \
  -H "Content-Type: application/json" \
  -d '{
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "tasks.max": "1",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "password",
    "database.dbname": "postgres",
    "database.server.name": "pgserver1",
    "publication.name": "debezium",
    "slot.name": "debezium_slot",
    "plugin.name": "pgoutput",
    "topic.prefix": "pg",
    "key.converter": "org.apache.kafka.connect.json.JsonConverter",
    "key.converter.schemas.enable": "false",
    "value.converter": "org.apache.kafka.connect.json.JsonConverter",
    "value.converter.schemas.enable": "false",
    "transforms": "unwrap",
    "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
    "transforms.unwrap.drop.tombstones": "true",
    "transforms.unwrap.delete.handling.mode": "drop",
    "transforms.unwrap.add.fields": "op"
  }'
```

4. Test

```sql
INSERT INTO tuples (sbj_ns, sbj_id, relation, obj_ns, obj_id) VALUES ('1', '1', '1', '1', '1');
```

``` zsh
docker exec -it kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 \
  --topic pg.public.tuples \
  --from-beginning
```
