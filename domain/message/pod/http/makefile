build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t message-http .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag message-http syamsuldocker/message-http:0.0.2
	docker push syamsuldocker/message-http:0.0.2
	docker image rm syamsuldocker/message-http:0.0.1

deploy:
	make push
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

alias:
	alias k="kubectl --kubeconfig=../../../../kubeconfig.yaml"

# docker run -it --rm --network my-network --name message-http -p 5000:5000 ubuntu  /bin/bash