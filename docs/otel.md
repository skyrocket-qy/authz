# OTEL
## Trace
Tempo
### Current
Level
- General: Async send from app.otel to tempo
- Error: Sync send too sidecar otel collector(with persistent storage), then send to tempo
### Future
Level
- General: Async send from app.otel to tempo
- Error: Sync write to file(with rotate), then a collector parse to tempo
## Metric
Mimir

## Log
Loki