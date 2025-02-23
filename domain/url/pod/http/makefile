build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t url-http .

run:
	docker-compose up -d

down:
	docker-compose down

push:
	docker tag url syamsuldocker/url
	docker push syamsuldocker/url

# docker run -it --rm --network my-network --name url-http -p 5000:5000 ubuntu  /bin/bash