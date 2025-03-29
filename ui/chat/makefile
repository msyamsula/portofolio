build:
	docker build -t chat-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag chat-ui syamsuldocker/chat-ui:0.0.24
	docker push syamsuldocker/chat-ui:0.0.24

deploy:
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml
