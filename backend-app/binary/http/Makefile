ENV_FILE ?= .env

-include $(ENV_FILE)
export

IMAGE ?= portofolio-http:local
CONTAINER ?= portofolio-http
NETWORK ?= test
PORT ?= $(SERVER_PORT)

# Usage (Docker + Swagger)
# 1) Generate Swagger docs:
#    make swagger
# 2) Build image:
#    make build
# 3) Run container:
#    make start
# 4) Open Swagger UI:
#    http://localhost:5000/swagger/index.html
# One-command flow:
#    make up

.PHONY: build start stop swagger up

build:
	docker build -f Dockerfile -t $(IMAGE) ../../..

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g backend-app/binary/http/main.go -o docs -d ../../.. --parseDependency --parseInternal

up: swagger build start

start:
	docker run -d --rm --name $(CONTAINER) --network $(NETWORK) \
		--env-file $(ENV_FILE) \
		-p $(PORT):5000 \
		$(IMAGE)

stop:
	docker stop $(CONTAINER)
