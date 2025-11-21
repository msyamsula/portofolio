SHELL := /bin/bash
include .env

run:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t user .

rundocker:
	docker run --rm -p ${PORT}:${PORT} \
	--network test \
	--env-file .env \
	--name user \
	user