build:
	docker build -t graph-ui .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag graph-ui syamsuldocker/graph-ui:0.0.8
	docker push syamsuldocker/graph-ui:0.0.8
	docker image rm syamsuldocker/graph-ui:0.0.7

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml
