build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t chat-server .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag chat-server syamsuldocker/chat-server:0.0.8
	docker push syamsuldocker/chat-server:0.0.8
	docker image rm syamsuldocker/chat-server:0.0.7

deploy:
	make down
	make push
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name chat-server -p 5000:5000 ubuntu  /bin/bash