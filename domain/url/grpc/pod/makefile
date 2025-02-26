build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t url-grpc .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag url-grpc syamsuldocker/url-grpc:0.0.2
	docker push syamsuldocker/url-grpc:0.0.2
	docker image rm syamsuldocker/url-grpc:0.0.1