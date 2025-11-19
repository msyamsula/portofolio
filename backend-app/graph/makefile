SHELL := /bin/bash

include .env

rungo:
	set -a; source .env; set +a; go run main.go
	
build:
	docker build --platform linux/arm64 -t graph .

run:
	docker run -it --name graph --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	graph

runkube:
	aws eks --region ap-southeast-1 update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap graph-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml
	kubectl apply -f canary.yaml

deploy-canary:
	kubectl set image deployment/graph-canary-deployment graph-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/graph:${VERSION}

deploy:
	kubectl set image deployment/graph-deployment graph=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/graph:${VERSION}

create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name graph-stack

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name graph-stack