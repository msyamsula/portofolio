SHELL := /bin/bash

runkube:
	kubectl apply -f deployment.yaml

run:
	docker-compose -f docker-compose-prometheus.yaml up -d

restart:
	kubectl rollout restart deployment prometheus