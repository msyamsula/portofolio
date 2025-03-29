build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t consumer .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag consumer syamsuldocker/consumer:0.0.5
	docker push syamsuldocker/consumer:0.0.5

deploy:
	make push
	kubectl --kubeconfig=../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name consumer -p 5000:5000 ubuntu  /bin/bash