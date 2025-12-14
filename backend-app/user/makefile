SHELL := /bin/bash
include .env

local:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t user .

rundocker:
	docker run --rm -p ${PORT}:${PORT} \
	--network test \
	--env-file .env \
	--name user \
	user

setup-env:
	kubectl create configmap user-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n default

kube:
	kubectl apply -f deployment.yaml -n default

deploy:
	kubectl set image deployment/user user=docker.io/library/user:${VERSION} -n default
	kubectl set image deployment/user-canary user-canary=docker.io/library/user:${VERSION} -n default

delete:
	kubectl delete -f deployment.yaml -n default

load:
	kind load docker-image user:${VERSION} --name ${CLUSTER}


### aws env
aws-kube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${CLUSTER}

aws-deploy-canary:
	kubectl set image deployment/user-canary user-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/user:${VERSION}

aws-deploy:
	kubectl set image deployment/user user=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/user:${VERSION}