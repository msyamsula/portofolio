SHELL := /bin/bash

setup-env:
	kubectl create configmap prometheus-config --from-file=config.yaml -n obs

runkube:
	kubectl apply -f deployment.yaml -n obs

run:
	docker-compose -f docker-compose-prometheus.yaml up -d

restart:
	kubectl rollout restart deployment prometheus-deployment -n obs