INFRA_DIR ?= backend-app/infrastructure/instance/local
HTTP_DIR ?= backend-app/binary/http
PG_AGENT_DIR ?= backend-app/binary/pg-agent
CV_DIR ?= frontend-app/cv

.PHONY: infra-start infra-stop swagger up stop start-backend stop-backend start-frontend stop-frontend start-all stop-all test generate pg-agent pg-agent-build

start-all: start-backend start-frontend

stop-all: stop-frontend stop-backend

restart: stop-all start-all

start-backend: infra-start up

stop-backend: stop infra-stop

start-frontend:
	$(MAKE) -C $(CV_DIR) start

stop-frontend:
	$(MAKE) -C $(CV_DIR) stop

infra-start:
	$(MAKE) -C $(INFRA_DIR) start

infra-stop:
	$(MAKE) -C $(INFRA_DIR) stop

swagger:
	$(MAKE) -C $(HTTP_DIR) swagger

up:
	$(MAKE) -C $(HTTP_DIR) up

stop:
	$(MAKE) -C $(HTTP_DIR) stop

test:
	go test -v $$(go list ./backend-app/... | grep -v -e /mock -e /binary) -count=1 -coverprofile=coverage.out
	grep -v -e '/mock/' -e '/binary/' coverage.out > coverage.filtered.out
	go tool cover -func=coverage.filtered.out
	go tool cover -html=coverage.filtered.out -o coverage.html
	open coverage.html

generate:
	go generate ./backend-app/...

pg-agent: pg-agent-build
	@echo "Starting PostgreSQL CLI Agent..."
	@cd $(PG_AGENT_DIR) && go run main.go

pg-agent-build:
	@echo "Building pg-agent..."
	@go build -o bin/pg-agent $(PG_AGENT_DIR)/main.go

pg-agent-run:
	@echo "Running PostgreSQL CLI Agent..."
	@./bin/pg-agent
