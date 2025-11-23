SHELL := /bin/bash
include .env

create:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name ${STACK_NAME} \
	--parameter-overrides \
		EksClusterName='${EKS_CLUSTER_NAME}' \
		EksMaxNodes='${EKS_MAX_NODES}' \
		EksDesiredNodes='${EKS_DESIRE_NODES}' \
		SnsTopic='${SOCKET_SNS_TOPIC}' \
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

use-aws-kubernetes:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}

destroy:
	aws cloudformation delete-stack --stack-name ${STACK_NAME}

tunnel-postgres:
	ssh -i ${KEY_NAME}.pem \
		-L 1${POSTGRES_PORT}:${POSTGRES_HOST}:${POSTGRES_PORT} \
		ec2-user@${BASTION_IP}

tunnel-redis:
	ssh -i ${KEY_NAME}.pem \
		-L 1${REDIS_PORT}:${REDIS_HOST}:${REDIS_PORT} \
		ec2-user@${BASTION_IP}

token:
	aws sts get-caller-identity

jaeger:
	docker-compose -f jaeger.yaml up -d


