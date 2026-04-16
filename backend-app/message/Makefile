.PHONY: run down build test tidy migrate-up migrate-down

# run: build all images and start the full stack.
# Services: postgres, kafka (KRaft), migrate, create-topics, api, worker, reconciler.
run:
	docker compose up -d --build

# down: stop all containers and remove volumes (clean slate).
down:
	docker compose down -v

# build: compile all three binaries locally (requires local Go toolchain).
build:
	go build -o bin/api        ./cmd/api
	go build -o bin/worker     ./cmd/worker
	go build -o bin/reconciler ./cmd/reconciler

# test: run all unit tests.
test:
	go test ./...

# tidy: regenerate go.mod / go.sum after dependency changes.
tidy:
	go mod tidy

# migrate-up / migrate-down: apply or roll back schema without restarting the stack.
migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate sh -c \
	  "psql '$$DATABASE_URL' -f /migration/down.sql"
