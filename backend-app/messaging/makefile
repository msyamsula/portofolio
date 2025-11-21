SHELL := /bin/bash
# include .env

create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name messaging

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name eks


