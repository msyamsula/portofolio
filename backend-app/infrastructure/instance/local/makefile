GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.PHONY: start stop status clean

ENV_FILE ?= .env
NETWORK_NAME ?= test

start:
	@echo "$(GREEN)Starting local cluster...$(NC)"
	@docker network create $(NETWORK_NAME) 2>/dev/null || echo "Network '$(NETWORK_NAME)' already exists"
	@docker compose --env-file $(ENV_FILE) up -d
	@echo "$(GREEN)All services started$(NC)"

stop:
	@echo "$(YELLOW)Stopping local cluster...$(NC)"
	@docker compose --env-file $(ENV_FILE) down
	@echo "$(GREEN)All services stopped$(NC)"

status:
	@docker compose --env-file $(ENV_FILE) ps

clean: stop
	@echo "$(YELLOW)Removing network '$(NETWORK_NAME)'...$(NC)"
	@docker network rm $(NETWORK_NAME) 2>/dev/null || echo "Network '$(NETWORK_NAME)' already removed"
