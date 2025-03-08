build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t graph-http .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag graph-http syamsuldocker/graph-http:0.0.13
	docker push syamsuldocker/graph-http:0.0.13
	docker image rm syamsuldocker/graph-http:0.0.12

deploy:
	make down
	make push
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name graph-http -p 5000:5000 ubuntu  /bin/bash