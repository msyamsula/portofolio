build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t user-http .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag user-http syamsuldocker/user-http:0.0.4
	docker push syamsuldocker/user-http:0.0.4
	docker image rm syamsuldocker/user-http:0.0.3

deploy:
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml

alias:
	alias k="kubectl --kubeconfig=../../../kubeconfig.yaml"

# docker run -it --rm --network my-network --name user-http -p 5000:5000 ubuntu  /bin/bash