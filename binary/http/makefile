build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t backend .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag backend syamsuldocker/backend:0.0.8
	docker push syamsuldocker/backend:0.0.8

deploy:
#this one is not pushed, manually configure
# kubectl --kubeconfig=../../kubeconfig.yaml apply -f secret.yaml
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml

alias:
	alias k="kubectl --kubeconfig=../../kubeconfig.yaml"

# docker run -it --rm --network my-network --name backend -p 5000:5000 ubuntu  /bin/bash