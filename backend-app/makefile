SHELL := /bin/bash
include .env

create:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name ${STACK_NAME} \
	--parameter-overrides \
		EksClusterName='${EKS_CLUSTER_NAME}' \
		EksMaxNodes='${EKS_MAX_NODES}' \
		EksDesiredNodes='${EKS_DESIRE_NODES}' \
		SnsTopic='${SNS_TOPIC}' \
		DbPassword='${POSTGRES_PASSWORD}' \
		DbUsername='${POSTGRES_USER}' \
		VpcCidr='${VPC_CIDR}' \
		DynamoTableName='${DYNAMO_TABLE}' \
		SnsPersistenceTopic='${SNS_PERSISTENCE_TOPIC}' \
		SqsPersistenceName='${SQS_PERSISTENCE_NAME}' \
	--capabilities CAPABILITY_NAMED_IAM

destroy:
	aws cloudformation delete-stack --stack-name ${STACK_NAME}


