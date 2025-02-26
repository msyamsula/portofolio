build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t url-http .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag url-http syamsuldocker/url-http:0.0.15
	docker push syamsuldocker/url-http:0.0.15
	docker image rm syamsuldocker/url-http:0.0.14

deploy:
	kubectl --kubeconfig=../../../../kubeconfig.yaml apply -f deployment.yaml

# docker run -it --rm --network my-network --name url-http -p 5000:5000 ubuntu  /bin/bash