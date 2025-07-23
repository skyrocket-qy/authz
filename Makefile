pg:
	docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:17.2

fbit:
	docker run -d \
	--name=fluentbit \
	--log-driver=none \
	-v /var/log:/var/log:ro \
	-v /var/lib/docker/containers:/var/lib/docker/containers:ro \
	-v $(pwd)/fluent-bit.conf:/fluent-bit/etc/fluent-bit.conf \
	-p 2020:2020 \
	fluent/fluent-bit:2.2

mysql:
	docker run -d \
		--name mysql-container \
		-p 5432:3306 \
		-e MYSQL_ROOT_PASSWORD=admin \
		-e MYSQL_USER=admin \
		-e MYSQL_PASSWORD=admin \
		-e MYSQL_DATABASE=mydb \
		mysql:9.1

redis:
	docker run -d \
	--name my-redis \
	-p 6379:6379 \
	redis \
	redis-server --requirepass password

build-img:
	docker build -t go-server-template .

run-container:
	docker run -d --name go-server-template go-server-template

bk:
	git add .
	git commit -m "backup"
	git push

gen-rest-doc:
	swag init --output ./docs/openapi  -g main.go internal/controller/*

lint:
	golangci-lint run ./...

mgdiff:
	atlas migrate diff --env local

apply:
	atlas migrate apply --env local

hash:
	atlas migrate hash

load-db-to-migrations:
	atlas migrate diff --dev-url "docker://postgres/15/dev" \
	--to "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"


gen-jwt-key:
	openssl rand -base64 64