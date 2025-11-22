SHELL := /bin/bash
include .env

rungo:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t message .

run:
	docker run -it --name message --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	message

runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap message-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml

deploy-canary:
	kubectl set image deployment/message-canary-deployment message-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/message:${VERSION}

deploy:
	kubectl set image deployment/message-deployment message=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/message:${VERSION}

