docker exec -it postgres psql -U debezium -d mydb

-- Example table
CREATE TABLE customers (
  id SERIAL PRIMARY KEY,
  first_name TEXT,
  last_name TEXT
);

-- Publication for Debezium
CREATE PUBLICATION my_publication FOR ALL TABLES;

curl -X PUT http://localhost:8083/connectors/postgres-connector/config \
  -H "Content-Type: application/json" \
  -d '{
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "tasks.max": "1",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "debezium",
    "database.password": "dbz",
    "database.dbname": "mydb",
    "database.server.name": "pgserver1",
    "publication.name": "my_publication",
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


INSERT INTO customers (first_name, last_name) VALUES ('John', 'Doe');

docker exec -it kafka /opt/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 \
  --topic pg.public.customers \
  --from-beginning
