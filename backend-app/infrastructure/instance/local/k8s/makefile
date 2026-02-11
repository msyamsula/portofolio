SHELL := /bin/bash
include .env

GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.PHONY: start stop status

start:
	@echo "$(GREEN)Starting local Kind cluster: $(CLUSTER)$(NC)"
	kind create cluster --config kind-config.yaml --name ${CLUSTER}
	@echo "$(GREEN)Cluster $(CLUSTER) started successfully$(NC)"

stop:
	@echo "$(YELLOW)Stopping local Kind cluster: $(CLUSTER)$(NC)"
	kind delete cluster --name ${CLUSTER}
	@echo "$(GREEN)Cluster $(CLUSTER) stopped successfully$(NC)"

status:
	@kind get clusters
