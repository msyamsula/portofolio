COMPOSE ?= docker compose

.DEFAULT_GOAL := help

.PHONY: help up down restart ps logs config pull

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services in the background
	$(COMPOSE) up -d --remove-orphans

down: ## Stop and remove all services
	$(COMPOSE) down

restart: ## Restart all services
	$(COMPOSE) restart

ps: ## Show service status and health
	$(COMPOSE) ps

logs: ## Follow logs from all services
	$(COMPOSE) logs --follow --tail=100

config: ## Validate and render the Compose configuration
	$(COMPOSE) config

pull: ## Pull service images
	$(COMPOSE) pull

