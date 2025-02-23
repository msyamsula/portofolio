build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go
	docker build -t chat-websocket .

run:
	docker-compose up

down:
	docker-compose down

push:
	docker tag chat-websocket syamsuldocker/chat-websocket:0.0.0
	docker push syamsuldocker/chat-websocket:0.0.0

# docker run -it --rm --network my-network --name chat-websocket -p 5000:5000 ubuntu  /bin/bash