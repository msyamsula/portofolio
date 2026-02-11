GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

.PHONY: start stop status

start:
	@echo "$(GREEN)Starting all local infrastructure...$(NC)"
	@docker network create test 2>/dev/null || echo "Network 'test' already exists"
	@$(MAKE) -C k8s start
	@$(MAKE) -C postgres start
	@$(MAKE) -C redis start
	@echo "$(GREEN)All services started$(NC)"

stop:
	@echo "$(YELLOW)Stopping all local infrastructure...$(NC)"
	@$(MAKE) -C postgres stop
	@$(MAKE) -C redis stop
	@$(MAKE) -C k8s stop
	@echo "$(GREEN)All services stopped$(NC)"

clean: stop
	@echo "$(YELLOW)Removing network 'test'...$(NC)"
	@docker network rm test 2>/dev/null || echo "Network 'test' already removed"

status:
	@echo "$(GREEN)=== Kubernetes ===$(NC)"
	@$(MAKE) -C k8s status
	@echo ""
	@echo "$(GREEN)=== PostgreSQL ===$(NC)"
	@$(MAKE) -C postgres status
	@echo ""
	@echo "$(GREEN)=== Redis ===$(NC)"
	@$(MAKE) -C redis status
