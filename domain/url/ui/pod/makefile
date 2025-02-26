build:
	docker build -t url-ui .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag url-ui syamsuldocker/url-ui:0.0.4
	docker push syamsuldocker/url-ui:0.0.4
	docker image rm syamsuldocker/url-ui:0.0.3

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml
