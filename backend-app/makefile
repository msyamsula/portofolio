SHELL := /bin/bash
include .env

create-aws-stack:
	aws cloudformation deploy --template-file eks-cluster.yaml \
	--stack-name eks \
	--parameter-overrides \
		 EksClusterName='${EKS_CLUSTER_NAME}' \
		 EksMaxNodes='${EKS_MAX_NODES}' \
		 EksDesiredNodes='${EKS_DESIRE_NODES}' \
	--capabilities CAPABILITY_NAMED_IAM

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name eks


