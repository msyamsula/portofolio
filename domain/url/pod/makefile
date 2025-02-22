build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t url .

run:
	docker-compose up -d

down:
	docker-compose down