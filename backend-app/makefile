SHELL := /bin/bash
include .env

## stack and repo creation
repo:
	aws cloudformation deploy --template-file ecr-repo.yaml \
		--stack-name repo \
		--capabilities CAPABILITY_NAMED_IAM

destroy-repo:
	aws cloudformation delete-stack --stack-name repo

create:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name ${STACK_NAME} \
	--parameter-overrides \
		EksClusterName='${EKS_CLUSTER_NAME}' \
		EksMaxNodes='${EKS_MAX_NODES}' \
		EksDesiredNodes='${EKS_DESIRE_NODES}' \
		SnsTopic='socket-io' \
		DbPassword='${POSTGRES_PASSWORD}' \
		DbUsername='${POSTGRES_USER}' \
		VpcCidr='${VPC_CIDR}' \
		DynamoTableName='${DYNAMO_TABLE}' \
		SnsPersistenceTopic='${SNS_PERSISTENCE_TOPIC}' \
		SqsPersistenceName='${SQS_PERSISTENCE_QUEUE}' \
		KeyName='${KEY_NAME}' \
		PrivateSubnetCidr1='${PRIVATE_SUBNET_CIDR1}' \
		PrivateSubnetCidr2='${PRIVATE_SUBNET_CIDR2}' \
	--capabilities CAPABILITY_NAMED_IAM

destroy:
	aws cloudformation delete-stack --stack-name ${STACK_NAME}

## identity and kubernetes connect
kube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}

token:
	aws sts get-caller-identity

## loud balancer provider in kind
## in cloud provider you do not have to do this
## this install, run, and allow is just for mimicking cloud behavior in local
install-lb-cloud-provider:
	brew install cloud-provider-kind

run-lb-provider:
	sudo cloud-provider-kind

## istio
istio-install:
	istioctl install -f istio.yaml -y

istio-label:
	kubectl label namespace default istio-injection=enabled

istio-check:
	kubectl get pods -n istio-system
	kubectl get namespace --show-labels

istio-uninstall:
	istioctl uninstall --purge -y


## bastion tunneling
tunnel-postgres:
	ssh -i ${KEY_NAME}.pem \
		-L 1${POSTGRES_PORT}:${POSTGRES_HOST}:${POSTGRES_PORT} \
		ec2-user@${BASTION_IP}

tunnel-redis:
	ssh -i ${KEY_NAME}.pem \
		-L 1${REDIS_PORT}:${REDIS_HOST}:${REDIS_PORT} \
		ec2-user@${BASTION_IP}

tunnel-neptune:
	ssh -i ${KEY_NAME}.pem \
		-L 1${NEPTUNE_PORT}:${NEPTUNE_HOST}:${NEPTUNE_PORT} \
		ec2-user@${BASTION_IP}


