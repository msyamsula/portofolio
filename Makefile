INFRA_DIR ?= backend-app/infrastructure/instance/local
HTTP_DIR ?= backend-app/binary/http

.PHONY: infra-start infra-stop swagger up stop

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
