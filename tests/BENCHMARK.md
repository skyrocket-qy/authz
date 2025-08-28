# BENCHMARK

## Env

Mac 15.6.0
i7-9750H CPU @ 2.60GHz
32 GB 2667 MHz DDR4

### Docker container spec

```yaml
mem_limit: 8g
cpus: 4 
```

### Setting

```yaml
roles     = 10000
resources = 1000
users     = 100000
permission = "read"
total tuples = 110000
```

## Result

Test duration: 10s

### 1vu

```yaml
http_req_duration: 
avg=927.48µs min=616µs    med=810µs    max=27.82ms p(90)=1.22ms p(95)=1.48ms
rps: 1001.709445/s
```

### 50vus

```yaml
http_req_duration: 
avg=12.08ms min=674µs med=9.53ms max=147.71ms p(90)=24.81ms p(95)=31.29ms
rps: 4100.409691/s
```

### 100vus

```yaml
http_req_duration: 
avg=23.67ms min=758µs    med=21.03ms max=152.81ms p(90)=40.08ms p(95)=49.2ms
rps: 4198.033303/s
```

### 200vus

```yaml
http_req_duration: 
avg=48.68ms min=656µs    med=46.13ms max=213.08ms p(90)=74.04ms p(95)=86.54ms
rps: 4086.3177/s
```

### 500vus

```yaml
http_req_duration: 
avg=119.86ms min=684µs    med=114.1ms  max=579.27ms p(90)=177.32ms p(95)=204.34ms
rps: 4116.146999/s
```
