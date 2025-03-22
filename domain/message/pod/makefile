build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t message-consumer .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag message-consumer syamsuldocker/message-consumer:0.0.1
	docker push syamsuldocker/message-consumer:0.0.1
	# docker image rm syamsuldocker/message-consumer:0.0.0

deploy:
	make down
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name message-consumer -p 5000:5000 ubuntu  /bin/bash