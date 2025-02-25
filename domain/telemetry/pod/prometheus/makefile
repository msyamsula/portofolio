run:
	docker-compose up
down:
	docker-compose down
deploy:
# one time only 
# kubectl --kubeconfig=../../../../kubeconfig.yaml create configmap prometheus-configmap --from-file=./config.yaml
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

