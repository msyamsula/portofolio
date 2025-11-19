SHELL := /bin/bash

rungo:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/amd64,linux/arm64 -t url-shortener .

run:
	docker run -it --name url-shortener --network=test --rm \
	-p 10000:10000 \
	-e POSTGRES_PASSWORD=admin \
	-e POSTGRES_USER=admin \
	-e POSTGRES_DB=postgres \
	-e POSTGRES_HOST=postgres \
	-e POSTGRES_PORT=5432 \
	-e REDIS_HOST=redis \
	-e REDIS_PORT=6379 \
	-e REDIS_PASSWORD=admin \
	-e JAEGER_HOST=jaeger \
	-e CALLBACK_URI=http://localhost:10000 url-shortener