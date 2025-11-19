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
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EksClusterName}
	kubectl create configmap url-shortener-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml

create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name graph-stack

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name graph-stack