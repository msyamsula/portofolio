.PHONY: env up down test build

env:
	@if [ -f .env ]; then \
		while IFS= read -r line; do \
			case "$$line" in \
				""|\#*) continue ;; \
			esac; \
			key=$${line%%=*}; \
			if ! grep -q "^$${key}=" .env; then \
				printf "%s\n" "$$line" >> .env; \
				echo "added $${key} to .env"; \
			fi; \
		done < .env.example; \
		echo ".env is up to date"; \
	else \
		cp .env.example .env; \
		echo "created .env from .env.example"; \
	fi

up: env
	docker compose up --build

down:
	docker compose down

test:
	go test ./...

build:
	go build ./cmd/api ./cmd/job-orchestrator
