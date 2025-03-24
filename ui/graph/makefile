build:
	docker build -t graph-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag graph-ui syamsuldocker/graph-ui:0.0.11
	docker push syamsuldocker/graph-ui:0.0.11

deploy:
	make build
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml
