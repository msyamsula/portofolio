build:
	docker build -t graph-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag graph-ui syamsuldocker/graph-ui:0.0.10
	docker push syamsuldocker/graph-ui:0.0.10
	docker image rm syamsuldocker/graph-ui:0.0.9

deploy:
	make down
	make build
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml
