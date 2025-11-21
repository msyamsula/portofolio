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

runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap user-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml
	kubectl apply -f canary.yaml

deploy-canary:
	kubectl set image deployment/user-canary-deployment user-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/user:${VERSION}

deploy:
	kubectl set image deployment/user-deployment user=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/user:${VERSION}