build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t url-http .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag url-http syamsuldocker/url-http:0.0.17
	docker push syamsuldocker/url-http:0.0.17
	docker image rm syamsuldocker/url-http:0.0.16

deploy:
	make push
	kubectl --kubeconfig=.../../../kubeconfig.yaml apply -f deployment.yaml

alias:
	alias k="kubectl --kubeconfig=../../../kubeconfig.yaml"

# docker run -it --rm --network my-network --name url-http -p 5000:5000 ubuntu  /bin/bash