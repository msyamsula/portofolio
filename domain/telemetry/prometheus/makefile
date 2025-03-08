run:
	docker-compose up -d
down:
	docker-compose down
deploy:
	kubectl --kubeconfig=../../../kubeconfig.yaml delete configmap prometheus-configmap
	kubectl --kubeconfig=../../../kubeconfig.yaml create configmap prometheus-configmap --from-file=./config.yaml
	kubectl --kubeconfig=../../../kubeconfig.yaml delete -f deployment.yaml
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml

