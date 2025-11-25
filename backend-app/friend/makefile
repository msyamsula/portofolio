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

runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap friend-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml
	kubectl apply -f canary.yaml

deploy-canary:
	kubectl set image deployment/friend-canary friend-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/friend:${VERSION}

deploy:
	kubectl set image deployment/friend friend=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/friend:${VERSION}