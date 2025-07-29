✅ Update Model
DB is the source of truth

Each permission update:

Writes to DB

Publishes an event to Kafka with a version number

Each authz replica:

Listens to Kafka, applies delta

Retries if failed

Falls back to rebuild from DB if retry fails 3 times

Use distributed lock to make sure not all srv updated at the same time

