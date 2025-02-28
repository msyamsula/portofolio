build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t graph-http .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag graph-http syamsuldocker/graph-http:0.0.11
	docker push syamsuldocker/graph-http:0.0.11
	docker image rm syamsuldocker/graph-http:0.0.10

deploy:
	make build
	make push
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name graph-http -p 5000:5000 ubuntu  /bin/bash