# Message V2 — State Machine with Outbox + Kafka

## Context

The existing `message` domain is a simple insert/query system with no state tracking or reliability guarantees. `messagev2` replaces this with a full state machine (SENT → RECEIVED → READ | FAIL) backed by the outbox pattern and Kafka for event-driven state transitions, with WebSocket-based real-time delivery and Redis-backed client address registry.

---

## Production Readiness Critique of the Original Plan

The design is solid but has these gaps that must be addressed:

| Gap | Fix |
|-----|-----|
| **START state is dead code** | Removed. API inserts directly as SENT. |
| **Outbox not atomic** | INSERT message + INSERT outbox must be in the same DB transaction. Without this, you can have messages with no outbox event (silent loss). |
| **No Kafka DLQ** | Consumer processing errors go to `dlq.messagev2` immediately (commit offset, don't block). Without DLQ, poison-pill messages halt the consumer. |
| **Consumer group not specified** | All consumers must be in the same consumer group ID per topic so horizontal scaling works without duplicate processing. |
| **Max retry undefined** | Set to 3. The Reconciler owns retry tracking — after retry==3 on an expired SENT message, it marks FAILED and publishes to `messagev2.failed`. SENT consumer never checks retry count. |
| **Conversation pointer storage** | Add `user_conversation_pointer` table in Postgres. |
| **RECEIVED reconciliation gap** | RECEIVED messages are not reconciled by design — READ is terminal and client-triggered. Acceptable, but acknowledge: if client never opens, message stays RECEIVED forever. |
| **WebSocket forward failure** | Forward is fire-and-forget (non-blocking). Failure is safe — client polls via HTTP. No circuit breaker needed. |
| **Expiry of 2min too short** | Make `EXPIRY_DURATION` configurable via env var (default 2min). Reconciliation interval also configurable. |
| **No message content** | `data TEXT NOT NULL` in schema; required in API request. |

---

## Architecture

```
API (POST /messagev2/send)
  └─ TX: INSERT messages_v2 (SENT) + INSERT outbox
         └─ Outbox poller (goroutine)
               └─ Kafka: messagev2.sent
                    └─ SENT Consumer
                         ├─ Append to conversation inbox (Redis list)
                         ├─ WebSocket forward (non-blocking, Redis registry lookup)
                         ├─ UPDATE messages_v2 status=RECEIVED (idempotent)
                         └─ (no outbox — READ is client-triggered)

API (POST /messagev2/read)
  └─ TX: UPDATE messages_v2 RECEIVED→READ WHERE conversation_id AND receiver_id=me
         └─ UPSERT user_conversation_pointer

Background Reconciler (every 1 min)
  └─ SELECT expired SENT messages (expiry < NOW AND status=SENT)
       └─ retry++ → Kafka: messagev2.sent  (if retry < 3)
       └─ retry==3 → status=FAILED + outbox → Kafka: messagev2.failed

FAILED Consumer
  └─ Alert/log (idempotent, DLQ-safe)

Kafka DLQ: dlq.messagev2 (for consumer-level failures)
```

---

## Database Schema

**File**: `backend-app/infrastructure/database/migration/messagev2_up.sql`

Migration UP:

```sql
-- Messages V2
CREATE TABLE IF NOT EXISTS messages_v2 (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id      TEXT NOT NULL UNIQUE,   -- client UUIDv7, idempotency key
    conversation_id TEXT NOT NULL,
    sender_id       BIGINT NOT NULL,
    receiver_id     BIGINT NOT NULL,
    data            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'SENT',
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retry           INTEGER NOT NULL DEFAULT 0,
    expiry          TIMESTAMPTZ NOT NULL
);

-- Outbox (transactional)
CREATE TABLE IF NOT EXISTS outbox (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  TEXT NOT NULL,
    event_type  TEXT NOT NULL,   -- 'SENT', 'FAILED'
    payload     JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ      -- NULL = pending
);

-- Per-user read pointer per conversation
CREATE TABLE IF NOT EXISTS user_conversation_pointer (
    user_id              BIGINT NOT NULL,
    conversation_id      TEXT NOT NULL,
    last_read_message_id TEXT NOT NULL,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, conversation_id)
);

CREATE INDEX IF NOT EXISTS idx_mv2_conversation ON messages_v2(conversation_id);
CREATE INDEX IF NOT EXISTS idx_mv2_sent_expiry  ON messages_v2(expiry) WHERE status = 'SENT';
CREATE INDEX IF NOT EXISTS idx_outbox_pending   ON outbox(created_at) WHERE processed_at IS NULL;
```

**File**: `backend-app/infrastructure/database/migration/messagev2_down.sql`

Migration DOWN (reverses UP completely):

```sql
-- Drop indexes
DROP INDEX IF EXISTS idx_outbox_pending;
DROP INDEX IF EXISTS idx_mv2_sent_expiry;
DROP INDEX IF EXISTS idx_mv2_conversation;

-- Drop tables (reverse order of creation to respect foreign key safety)
DROP TABLE IF EXISTS user_conversation_pointer;
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS messages_v2;
```

---

## Kafka Topics

| Topic | Producer | Consumer Group | Purpose |
|-------|----------|---------------|---------|
| `messagev2.sent`   | Outbox poller + Reconciler | `cg-messagev2-sent`   | Drive SENT → RECEIVED |
| `messagev2.failed` | Outbox poller              | `cg-messagev2-failed` | Alert on terminal failure |
| `dlq.messagev2`    | All consumers on error     | —                     | Dead letter queue |

---

## Redis Keys

| Key Pattern | Type | Purpose |
|------------|------|---------|
| `ws:client:{user_id}` | String | WebSocket server address for user |

---

## New Files to Create

### Infrastructure
- `backend-app/infrastructure/kafka/producer.go` — confluent-kafka-go producer wrapper with retry
- `backend-app/infrastructure/kafka/consumer.go` — confluent-kafka-go consumer wrapper (DLQ on error)

### Domain: messagev2
- `backend-app/domain/messagev2/dto/dto.go` — Message struct, status constants (SENT/RECEIVED/READ/FAIL), request/response DTOs
- `backend-app/domain/messagev2/repository/interface.go` — Repository interface
- `backend-app/domain/messagev2/repository/repository.go` — Postgres impl (InsertWithOutbox in TX, UpdateStatus idempotent, MarkRead batch, GetExpiredSent, MarkOutboxProcessed)
- `backend-app/domain/messagev2/service/interface.go` — Service interface
- `backend-app/domain/messagev2/service/service.go` — Business logic: SendMessage, ProcessSent, ReadConversation, Reconcile, ProcessFailed
- `backend-app/domain/messagev2/handler/handler.go` — HTTP handlers: POST /messagev2/send, POST /messagev2/read, POST /messagev2/ws/register, DELETE /messagev2/ws/register
- `backend-app/domain/messagev2/consumer/sent_consumer.go` — Kafka consumer for messagev2.sent topic
- `backend-app/domain/messagev2/consumer/failed_consumer.go` — Kafka consumer for messagev2.failed topic
- `backend-app/domain/messagev2/reconciler/reconciler.go` — Ticker goroutine (configurable interval, default 1min)

### WebSocket Registry
- `backend-app/infrastructure/websocket/registry.go` — Redis-backed `RegisterClient(userID, addr)`, `GetClient(userID)`, `DeregisterClient(userID)`

## New Binaries

### WebSocket Server: `backend-app/binary/ws/`
- `main.go` — WebSocket server: accepts WS connections, registers client in Redis on connect, deregisters on disconnect, forwards inbound messages to connected clients. Exposes `POST /internal/forward` for the SENT consumer to push messages.
- `Makefile` — mirrors `backend-app/binary/http/Makefile` (Docker build/start/stop/up)
- `Dockerfile` — mirrors the http Dockerfile

## Files to Modify

- `backend-app/infrastructure/database/migration/messagev2_up.sql` — New file: create messages_v2, outbox, user_conversation_pointer tables + indexes
- `backend-app/infrastructure/database/migration/messagev2_down.sql` — New file: drop all tables and indexes (rollback)
- `backend-app/infrastructure/instance/local/docker-compose.yaml` — Add Kafka (KRaft mode, single broker) service
- `backend-app/infrastructure/instance/local/makefile` — Add `create-topics` target
- `backend-app/binary/http/main.go` — Wire messagev2 handler, start consumers + reconciler + outbox poller goroutines, add Kafka + WS registry config
- `Makefile` (root) — Add ws server targets + migrate-up/down targets
- `backend-app/go.mod` / `go.sum` — Add `github.com/confluentinc/confluent-kafka-go/v2`

---

## Makefile Changes

### `backend-app/binary/ws/Makefile` (new — mirrors http Makefile)
```makefile
ENV_FILE ?= .env

-include $(ENV_FILE)
export

IMAGE     ?= portofolio-ws:local
CONTAINER ?= portofolio-ws
NETWORK   ?= test
PORT      ?= $(WS_PORT)

.PHONY: build start stop up

build:
	docker build -f Dockerfile -t $(IMAGE) ../../..

up: build start

start:
	docker run -d --rm --name $(CONTAINER) --network $(NETWORK) \
		--env-file $(ENV_FILE) \
		-p $(PORT):$(WS_PORT) \
		$(IMAGE)

stop:
	docker stop $(CONTAINER)
```

### Root `Makefile` additions
Add to existing variables and targets:
```makefile
WS_DIR ?= backend-app/binary/ws

# add to .PHONY: start-ws stop-ws migrate-up migrate-down

start-backend: infra-start up start-ws    # updated

stop-backend: stop-ws stop infra-stop     # updated

start-ws:
	$(MAKE) -C $(WS_DIR) up

stop-ws:
	$(MAKE) -C $(WS_DIR) stop

# Migration targets (requires psql in PATH and DB env vars set)
migrate-up:
	psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL)" \
		-f backend-app/infrastructure/database/migration/messagev2_up.sql

migrate-down:
	psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL)" \
		-f backend-app/infrastructure/database/migration/messagev2_down.sql
```

### `backend-app/infrastructure/instance/local/docker-compose.yaml` — Kafka addition
Add Kafka in KRaft mode (no Zookeeper) to the existing services:
```yaml
  kafka:
    image: confluentinc/cp-kafka:7.7.0
    container_name: local-kafka
    ports:
      - "${KAFKA_PORT:-9092}:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://local-kafka:9092
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@local-kafka:9093
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_LOG_DIRS: /var/lib/kafka/data
      CLUSTER_ID: ${KAFKA_CLUSTER_ID:-MkU3OEVBNTcwNTJENDM2Qk}
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "false"
    volumes:
      - kafka_data:/var/lib/kafka/data
    healthcheck:
      test: ["CMD", "kafka-topics", "--bootstrap-server", "localhost:9092", "--list"]
      interval: 10s
      timeout: 10s
      retries: 5
    restart: unless-stopped
```

Also add `kafka_data:` under `volumes:`.

### `backend-app/infrastructure/instance/local/makefile` — create-topics target
Add after existing targets:
```makefile
create-topics:
	@echo "$(GREEN)Creating Kafka topics...$(NC)"
	@docker exec local-kafka kafka-topics --bootstrap-server localhost:9092 \
		--create --if-not-exists --topic messagev2.sent   --partitions 3 --replication-factor 1
	@docker exec local-kafka kafka-topics --bootstrap-server localhost:9092 \
		--create --if-not-exists --topic messagev2.failed --partitions 1 --replication-factor 1
	@docker exec local-kafka kafka-topics --bootstrap-server localhost:9092 \
		--create --if-not-exists --topic dlq.messagev2    --partitions 1 --replication-factor 1
	@echo "$(GREEN)Topics created$(NC)"
```

---

## HTTP API

| Method | Path | Body / Params | Description |
|--------|------|--------------|-------------|
| POST | `/messagev2/send` | `{conversation_id, message_id, sender_id, receiver_id, data}` | Insert message (SENT), idempotent on message_id |
| POST | `/messagev2/read` | `{user_id, conversation_id, last_message_id}` | Mark RECEIVED→READ up to last_message_id |
| POST | `/messagev2/ws/register` | `{user_id, address}` | Register WebSocket server address in Redis |
| DELETE | `/messagev2/ws/register` | `?user_id=X` | Deregister client address |

---

## Key Implementation Details

### SendMessage (API handler)
1. Validate: conversation_id, message_id (UUIDv7 format), sender_id, receiver_id, data all required; sender != receiver
2. Begin TX
3. `INSERT INTO messages_v2 ... ON CONFLICT (message_id) DO NOTHING` — idempotent
4. `INSERT INTO outbox (message_id, event_type, payload)` for the new row only (skip if step 3 was a no-op via rows affected check)
5. Commit TX
6. Return 200 with message_id

### Outbox Poller (goroutine in main.go)
- Poll every N seconds (configurable, default 5s)
- `SELECT * FROM outbox WHERE processed_at IS NULL ORDER BY created_at LIMIT 100`
- Publish to Kafka topic by event_type
- `UPDATE outbox SET processed_at = NOW() WHERE id = $1` on success
- On publish failure: log, retry next poll cycle (outbox row stays unprocessed)

### SENT Consumer
1. Parse message_id from event
2. `SELECT status FROM messages_v2 WHERE message_id = $1`
3. If status already RECEIVED/READ/FAILED → return (idempotent — duplicate delivery safe)
4. `UPDATE messages_v2 SET status='RECEIVED', update_at=NOW() WHERE message_id=$1 AND status='SENT'` — conditional (idempotent)
5. Redis lookup `ws:client:{receiver_id}` → HTTP forward to WS server (goroutine, fire-and-forget, 500ms timeout)
6. On any processing error → publish raw event to `dlq.messagev2` and commit offset (poison-pill safe)

Note: The SENT consumer does NOT check retry count. The `retry` field is owned entirely by the Reconciler. Consumer failures are handled at the Kafka level — if the consumer errors, the message goes to DLQ after Kafka's max delivery attempts.

### ReadConversation (API handler)
1. Validate: user_id, conversation_id, last_message_id required
2. Begin TX
3. `UPDATE messages_v2 SET status='READ', update_at=NOW() WHERE conversation_id=$1 AND receiver_id=$2 AND status='RECEIVED' AND message_id <= $3` (UUIDv7 lexicographic ordering works for time ordering)
4. `INSERT INTO user_conversation_pointer ... ON CONFLICT (user_id, conversation_id) DO UPDATE SET last_read_message_id=$3, updated_at=NOW() WHERE last_read_message_id < $3`
5. Commit TX

### Reconciler
- Ticker every 1 min (configurable via env `RECONCILE_INTERVAL`)
- `SELECT message_id, retry FROM messages_v2 WHERE status='SENT' AND expiry < NOW()`
- For each: if retry < 3 → `UPDATE retry=retry+1, expiry=NOW()+2min` + insert outbox SENT event
- For each: if retry >= 3 → `UPDATE status='FAILED'` + insert outbox FAILED event
- Both in a transaction per message (idempotent updates)

### FAILED Consumer
- Log structured error with message_id, conversation_id, sender_id, receiver_id
- Increment a Prometheus/OTEL counter `messagev2.failed.total`
- Idempotent (duplicate alerts are acceptable per the spec)
- On consumer error → DLQ

---

## Configuration (env vars)

```
KAFKA_BROKERS          # e.g. local-kafka:9092
KAFKA_GROUP_SENT       # default: cg-messagev2-sent
KAFKA_GROUP_FAILED     # default: cg-messagev2-failed
MSG_EXPIRY_DURATION    # default: 2m
RECONCILE_INTERVAL     # default: 1m
OUTBOX_POLL_INTERVAL   # default: 5s
WS_FORWARD_TIMEOUT     # default: 500ms
WS_PORT                # default: 5001 (WebSocket server port)
WS_INTERNAL_BASE_URL   # e.g. http://local-ws:5001 (used by SENT consumer to POST /internal/forward)
```

---

## Verification

1. **Unit tests** — service logic (state transitions, idempotency guards), repository mock tests
2. **Integration test flow**:
   - POST /messagev2/send → assert DB row exists with status=SENT, outbox row created
   - Outbox poller runs → assert Kafka message published, outbox processed_at set
   - SENT consumer runs → assert DB status=RECEIVED
   - POST /messagev2/ws/register → assert Redis key set
   - POST /messagev2/read → assert DB status=READ, user_conversation_pointer updated
3. **Reconciler test**: Insert SENT message with expiry in past, run reconciler → assert retry++ or FAILED
4. **Idempotency test**: Send same message_id twice → assert only 1 DB row, only 1 outbox entry
5. **SENT consumer idempotency**: Deliver same Kafka message twice → assert status stays RECEIVED (not double-updated)
