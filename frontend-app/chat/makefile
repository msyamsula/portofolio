build:
	docker build -t chat-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag chat-ui syamsuldocker/chat-ui:0.0.35
	docker push syamsuldocker/chat-ui:0.0.35

deploy:
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml
