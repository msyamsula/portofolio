SHELL := /bin/bash
include ./websocket-server/src/.env

create-aws-stack:
	aws cloudformation deploy --template-file tech-stack.yaml \
	--stack-name messaging \
	--parameter-overrides \
		SnsTopic='${SNS_TOPIC}' \

destroy-aws-stack:
	aws cloudformation delete-stack --stack-name messaging


