SHELL := /bin/bash
include .env

rungo:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t url-shortener .

run:
	docker run -it --name url-shortener --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	url-shortener

runkube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}
	kubectl create configmap url-shortener-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml
	kubectl apply -f canary.yaml

deploy-canary:
	kubectl set image deployment/url-shortener-canary-deployment url-shortener-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:${VERSION}

deploy:
	kubectl set image deployment/url-shortener-deployment url-shortener=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:${VERSION}


create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name url-shortener-stack \
	--parameter-overrides DbPassword='${POSTGRES_PASSWORD}' \
		 VpcCidr='${VPC_CIDR}' \
		 DbUsername='${POSTGRES_USER}' \
		 EksClusterName='${EKS_CLUSTER_NAME}' \
		 EksMaxNodes='${EKS_MAX_NODES}' \
		 EksDesiredNodes='${EKS_DESIRE_NODES}' \
		 DynamoTableName='${DYNAMO_TABLE}' \
	--capabilities CAPABILITY_NAMED_IAM

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name url-shortener-stack


