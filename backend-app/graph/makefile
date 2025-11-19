SHELL := /bin/bash

rungo:
	set -a; source .env; set +a; go run main.go
	
build:
	docker build --platform linux/amd64,linux/arm64 -t graph .

run:
	docker run -it --name graph --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	graph

runkube:
	kubectl create configmap graph-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml