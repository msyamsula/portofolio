build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t chat-server .

run:
	docker-compose up

down:
	docker-compose down

push:
	docker tag chat-server syamsuldocker/chat-server:0.0.4
	docker push syamsuldocker/chat-server:0.0.4
	docker image rm syamsuldocker/chat-server:0.0.3

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name chat-server -p 5000:5000 ubuntu  /bin/bash