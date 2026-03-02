INFRA_DIR ?= backend-app/infrastructure/instance/local
HTTP_DIR ?= backend-app/binary/http
CV_DIR ?= frontend-app/cv

.PHONY: infra-start infra-stop swagger up stop start-backend stop-backend start-frontend stop-frontend start-all stop-all

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
