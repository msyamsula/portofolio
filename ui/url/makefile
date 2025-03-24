build:
	docker build -t url-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag url-ui syamsuldocker/url-ui:0.0.11
	docker push syamsuldocker/url-ui:0.0.11

deploy:
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml
