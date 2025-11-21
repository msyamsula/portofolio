SHELL := /bin/bash
include .env

run:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t friend .

rundocker:
	docker run --rm -p ${PORT}:${PORT} \
	--network test \
	--env-file .env \
	--name friend \
	friend