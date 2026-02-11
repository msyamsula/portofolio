SHELL := /bin/bash
include .env

GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.PHONY: start stop status logs

start:
	@echo "$(GREEN)Starting local PostgreSQL...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)PostgreSQL started successfully$(NC)"

stop:
	@echo "$(YELLOW)Stopping local PostgreSQL...$(NC)"
	docker-compose down
	@echo "$(GREEN)PostgreSQL stopped successfully$(NC)"

logs:
	docker-compose logs -f

status:
	@docker-compose ps
