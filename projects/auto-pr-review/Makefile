.PHONY: env up down test build kafka-topics kafka-topic-create kafka-topic-describe kafka-topic-offsets kafka-dlq-topic-offsets kafka-topic-read kafka-consume-debug kafka-group-describe kafka-group-state

env:
	@touch .env
	@while IFS= read -r line; do \
		case "$$line" in ""|\#*) continue ;; esac; \
		key=$${line%%=*}; \
		grep -q "^$${key}=" .env || printf "%s\n" "$$line" >> .env; \
	done < .env.example

up: env
	docker compose up --build --scale job-consumer=1 -d
	@echo "Waiting for kafka to be healthy..."
	@while ! docker compose exec -T kafka /opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list >/dev/null 2>&1; do sleep 2; done
	@$(MAKE) kafka-topic-create

down:
	docker compose down

test:
	go test ./...

build:
	go build ./cmd/api ./cmd/job-orchestrator ./cmd/job-consumer

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

kafka-topic-offsets:
	docker compose exec -T kafka /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --topic $(KAFKA_TOPIC) --time -1

kafka-dlq-topic-offsets:
	docker compose exec -T kafka /opt/kafka/bin/kafka-get-offsets.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --topic $(KAFKA_DLQ_TOPIC) --time -1

kafka-topic-read:
	docker compose exec -T kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --topic $(KAFKA_TOPIC) --from-beginning --formatter-property print.key=true --formatter-property print.partition=true --formatter-property print.timestamp=true

kafka-consume-debug:
	docker compose exec -T kafka /opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --topic $(KAFKA_TOPIC) --group $(KAFKA_DEBUG_GROUP) --from-beginning --property print.key=true --property print.partition=true --property print.timestamp=true

kafka-group-describe:
	docker compose exec -T kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --describe --group $(KAFKA_DEBUG_GROUP)

kafka-group-state:
	docker compose exec -T kafka /opt/kafka/bin/kafka-consumer-groups.sh --bootstrap-server $(KAFKA_BOOTSTRAP_SERVER) --describe --group $(KAFKA_RUNNING_GROUP)
