SHELL := /bin/bash
include .env

local:
	set -a; source .env; set +a; go run main.go

build:
	docker build --platform linux/arm64 -t url-shortener .

run:
	docker run -it --name url-shortener --network=test --rm \
	-p 10000:10000 \
	--env-file .env \
	url-shortener

setup-env:
	kubectl create configmap url-shortener-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -

kube:
	kubectl apply -f deployment.yaml

deploy:
	kubectl set image deployment/url-shortener url-shortener=docker.io/library/url-shortener:${VERSION} -n default
	kubectl set image deployment/url-shortener-canary url-shortener-canary=docker.io/library/url-shortener:${VERSION} -n default

delete:
	kubectl delete -f deployment.yaml -n default

load:
	kind load docker-image url-shortener:${VERSION} --name ${CLUSTER}



### aws env
aws-kube:
	aws eks --region ${AWS_REGION} update-kubeconfig --name ${EKS_CLUSTER_NAME}

aws-deploy-canary:
	kubectl set image deployment/url-shortener-canary url-shortener-canary=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:${VERSION}

aws-deploy:
	kubectl set image deployment/url-shortener url-shortener=717420436752.dkr.ecr.ap-southeast-1.amazonaws.com/url-shortener:${VERSION}

