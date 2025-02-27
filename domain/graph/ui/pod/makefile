build:
	docker build -t graph-ui .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag graph-ui syamsuldocker/graph-ui:0.0.3
	docker push syamsuldocker/graph-ui:0.0.3
	docker image rm syamsuldocker/graph-ui:0.0.2

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml
