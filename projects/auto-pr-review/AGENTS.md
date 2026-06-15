# auto-pr-review

Go service that registers GitHub links, schedules periodic PR reviews via Redis, dispatches work through Kafka, and generates LLM review reports.

## Quick commands

```bash
make test          # go test ./...
make build         # go build ./cmd/api ./cmd/job-orchestrator ./cmd/job-consumer
make env           # merge .env.example into .env (adds missing keys; does NOT update existing)
make up            # docker compose up --build (depends on env)
make down          # docker compose down
```

No lint tool is configured. Do not run `golangci-lint`.

## Three binaries, one Dockerfile

The single Dockerfile builds all three binaries. Docker Compose runs the same image with different `command:` for each service:

| Binary | Deps at runtime | Role |
|--------|----------------|------|
| `cmd/api` | Postgres + Redis | HTTP API server on `:8080` |
| `cmd/job-orchestrator` | Redis + Kafka (producer) | Claims due jobs from Redis sorted set, dispatches to Kafka |
| `cmd/job-consumer` | Kafka (consumer) + Redis + LLM | Consumes `pr_review`, runs PR review, reschedules job in Redis |

After first `make up`, Kafka topics are not auto-created. Run `make kafka-topic-create` once to create `pr_review` and `pr_review_dlq` (3 partitions each, Kraft mode, no Zookeeper).

## Architecture

```
handler -> service -> repository (interface) -> postgres / redis / kafka
```

- `internal/service/` — business logic (GitHub link, periodic job, PR review)
- `internal/postgres/` — pgxpool, prepared statements via `AfterConnect`, auto-creates schema with `IF NOT EXISTS`
- `internal/redisstore/` — sorted set `link:periodic:nexttimestamp`, atomic claim via Lua script
- `internal/kafkaproducer/` — Avro-encoded SyncProducer, keyed by GitHub URL
- `internal/kafkaconsumer/` — consumer group `auto_pr_review`, retry+backoff+jitter, DLQ
- `internal/orchestrator/` — ticker loop, claims due jobs, generates idempotency keys per dispatch
- `internal/llm/` — single `Analyzer` interface with Claude and ChatGPT implementations
- `internal/schemaregistry/` — Avro schema V1/V2 with evolution support

Config lives in a single `internal/config/` package (504 lines). All settings come from env vars, loaded once at startup.

## Gotchas

- **`make env` is append-only**: it adds missing keys from `.env.example` but never updates existing values.
- **`goavro` is indirect-only**: `github.com/linkedin/goavro/v2` is used throughout but only listed as `// indirect` in go.mod. Running `go mod tidy` will promote it to a direct dependency.
- **GitHub token per-request**: `GITHUB_TOKEN_PATH` is read on every API call (supports token rotation without restart).
- **Schema is auto-created**: postgres package creates table and indexes with `IF NOT EXISTS` on connect. No migration tool.
- **Integration test scripts**: `scripts/integration-pr-review.go` and `scripts/check-pr-access/main.go` load `.env` independently (don't share `config.Load()`).
- **Env changes require restart**: config is loaded at process startup only.
- **Sensitive files**: `privatekey/` contains PEM and token files. `.gitignore` excludes `privatekey/*` but keeps `.gitkeep`. Never commit secrets here.
