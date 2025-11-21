SHELL := /bin/bash
include ./src/.env

tidy:
	rm -rf node_modules package-lock.json
	npm install

run:
	set -a; source src/.env; set +a; node dist/main.js

build:
	npx tsc

buildrun:
	make build & make run

builddocker:
	docker build --platform linux/arm64 -t websocket .

rundocker:
	docker run --name websocket \
	--rm -p ${PORT}:${PORT} --network test \
	--env-file ./src/.env \
	websocket

