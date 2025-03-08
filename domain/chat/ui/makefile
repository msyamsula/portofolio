build:
	docker build -t chat-ui .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag chat-ui syamsuldocker/chat-ui:0.0.4
	docker push syamsuldocker/chat-ui:0.0.4
	docker image rm syamsuldocker/chat-ui:0.0.3

deploy:
	make down
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml
