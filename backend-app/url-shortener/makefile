SHELL := /bin/bash
include stack.env
include .env

rungo:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/amd64,linux/arm64 -t url-shortener .

run:
	docker run -it --name url-shortener --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	url-shortener

runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EksClusterName}
	kubectl create configmap url-shortener-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml

deploy-canary:
	kubectl set image deployment/url-shortener-canary-deployment url-shortener-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:679441d-21

deploy:
	kubectl set image deployment/url-shortener-deployment url-shortener=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:679441d-21


create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name url-shortener-stack \
	--parameter-overrides DbPassword='${DbPassword}' \
		 VpcCidr='${VpcCidr}' \
		 DbUsername='${DbUsername}' \
		 EksClusterName='${EksClusterName}' \
		 EksMaxNodes='${EksMaxNodes}' \
		 EksDesiredNodes='${EksDesiredNodes}' \
		 DynamoTableName='${DynamoTableName}' \
	--capabilities CAPABILITY_NAMED_IAM

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name url-shortener-stack


