build:
	docker build -t url-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag url-ui syamsuldocker/url-ui:0.0.10
	docker push syamsuldocker/url-ui:0.0.10
	docker image rm syamsuldocker/url-ui:0.0.9

deploy:
	make down
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml
