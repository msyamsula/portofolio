SHELL := /bin/bash
include .env

tidy:
	rm -rf node_modules package-lock.json
	npm install

run:
	set -a; source .env; set +a; node dist/main.js

build:
	npx tsc

buildrun:
	make build & make run

builddocker:
	docker build --platform linux/arm64 -t websocket .

rundocker:
	docker run --name websocket \
	--rm -p ${PORT}:${PORT} --network test \
	--env-file .env \
	websocket


runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap websocket-env --from-env-file=./src/.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml

deploy:
	kubectl set image deployment/websocket websocket=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/websocket:${VERSION}
