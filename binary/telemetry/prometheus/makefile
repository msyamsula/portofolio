run:
	docker-compose -f nsqd-one-exporter/docker-compose.yaml up -d
	docker-compose -f nsqd-two-exporter/docker-compose.yaml up -d
	docker-compose -f postgres-exporter/docker-compose.yaml up -d

	docker-compose up -d
down:
	docker-compose -f nsqd-one-exporter/docker-compose.yaml down
	docker-compose -f nsqd-two-exporter/docker-compose.yaml down
	docker-compose -f postgres-exporter/docker-compose.yaml down

	docker-compose down
deploy:
	kubectl --kubeconfig=../../../kubeconfig.yaml delete configmap prometheus-configmap
	kubectl --kubeconfig=../../../kubeconfig.yaml create configmap prometheus-configmap --from-file=./config.yaml
	kubectl --kubeconfig=../../../kubeconfig.yaml delete -f deployment.yaml
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml

