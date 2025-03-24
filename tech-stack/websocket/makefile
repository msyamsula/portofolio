build:
	docker build -t websocket-server .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag websocket-server syamsuldocker/websocket-server:0.0.4
	docker push syamsuldocker/websocket-server:0.0.4

deploy:
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name websocket-server -p 5000:5000 ubuntu  /bin/bash