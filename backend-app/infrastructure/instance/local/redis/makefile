SHELL := /bin/bash
include .env

GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.PHONY: start stop status logs

start:
	@echo "$(GREEN)Starting local Redis...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Redis started successfully$(NC)"

stop:
	@echo "$(YELLOW)Stopping local Redis...$(NC)"
	docker-compose down
	@echo "$(GREEN)Redis stopped successfully$(NC)"

logs:
	docker-compose logs -f

status:
	@docker-compose ps
