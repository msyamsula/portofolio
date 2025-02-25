build:
	docker build -t url-ui .

push:
	docker tag url-ui syamsuldocker/url-ui:0.0.3
	docker push syamsuldocker/url-ui:0.0.3
	docker image rm syamsuldocker/url-ui:0.0.2

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml
