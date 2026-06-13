.PHONY: env up down test build kafka-topics kafka-topic-create kafka-topic-describe kafka-consume-debug kafka-group-describe kafka-group-state

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

KAFKA_BOOTSTRAP_SERVER ?= localhost:9092
KAFKA_TOPIC ?= pr_review
KAFKA_DLQ_TOPIC ?= pr_review_dlq
KAFKA_DEBUG_GROUP ?= auto_pr_debugging
KAFKA_RUNNING_GROUP ?= auto_pr_review

kafka-topics:
	docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --list

kafka-topic-create:
	docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --create --if-not-exists --topic $(KAFKA_TOPIC) --partitions 3 --replication-factor 1
	docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --create --if-not-exists --topic $(KAFKA_DLQ_TOPIC) --partitions 3 --replication-factor 1

kafka-topic-describe:
	docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --describe --topic $(KAFKA_TOPIC)

kafka-consume-debug:
	docker compose exec -T kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --topic $(KAFKA_TOPIC) --group $(KAFKA_DEBUG_GROUP) --from-beginning --property print.key=true --property print.partition=true --property print.timestamp=true

kafka-group-describe:
	docker compose exec -T kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --describe --group $(KAFKA_DEBUG_GROUP)

kafka-group-state:
	docker compose exec -T kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --describe --group $(KAFKA_RUNNING_GROUP)
