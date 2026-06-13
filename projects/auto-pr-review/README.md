# auto-pr-review

Small Go service for registering GitHub links in Postgres, scheduling periodic
jobs in Redis, and dispatching work through Kafka.

## Run locally

```bash
make env
make up
make kafka-topic-create
```

The web UI and API are exposed on `localhost:8080` by default, and Postgres is
exposed on `localhost:5432`. Redis is exposed on `localhost:6379`.

Open the web page:

```bash
open http://localhost:8080
```

Register a GitHub link:

```bash
curl -i -X POST http://localhost:8080/post \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://github.com/octocat/Hello-World"}'
```

Register a periodic job in Redis:

```bash
curl -i -X POST http://localhost:8080/jobs \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://github.com/octocat/Hello-World","period":"5s"}'
```

Jobs are written to the Redis sorted set `link:periodic:nexttimestamp`. The
member is `github_link:period`, for example
`https://github.com/octocat/Hello-World:5s`, and the score is `next_timestamp`
ascending.

A separate `job-orchestrator` container runs the background loop. On each
`JOB_ORCHESTRATOR_INTERVAL`, it checks the smallest Redis score, skips empty or
not-yet-due jobs, atomically advances due jobs by their period, and currently
logs a `periodic job done` message as the placeholder processing function.

The web dropdown period options come from `JOB_PERIOD_OPTIONS`, for example
`1s,10m,5h`. Periodic interval state is persisted in Postgres on each GitHub
link; Redis only stores the next-claim sorted set used by the orchestrator.

The `job-consumer` container listens to `pr_review`, runs the PR review service
for each GitHub link, writes a Markdown report, retries with backoff and jitter,
and sends the message to `pr_review_dlq` after the retry budget is exhausted.

Kafka is used for dispatch after a claim:
- topic `pr_review` with 3 partitions
- dead-letter topic `pr_review_dlq`
- max retries `5`
- retry backoff `100ms`
- consumer retry budget `5`
- consumer retry backoff `5s`
- consumer attempt timeout `5s`
- max in-flight requests `10`
- consumer max in-flight `15`
- messages are keyed by GitHub link so the same link lands in the same partition
- Kafka broker limits connections with the configured broker settings
- Kafka retention is set to 2 days for this project
- payloads use an Avro record schema stored in the repo

Local monitoring and topic management:
- `make kafka-topics`
- `make kafka-topic-describe`
- `make kafka-consume-debug`
- `make kafka-group-describe`
- `make kafka-group-state`

The debug consumer group is `auto_pr_debugging`, so it reads without competing
with the running consumer group. Offset commits are per consumer group, so the
debug group and the running group track acknowledgements independently.
`make kafka-group-state` shows the running group `auto_pr_review`.

Configuration is read from environment variables at process startup. Changing
configuration requires restarting the service container.

Manually integration-test PR review without Kafka:

```bash
go run scripts/integration-pr-review.go https://github.com/owner/repo/pull/123
```

LLM review calls are selected through one provider interface. The current
default is Claude:

- `LLM_PROVIDER=claude`
- `LLM_API_KEY` or `ANTHROPIC_API_KEY`
- `LLM_MODEL` or `ANTHROPIC_MODEL`

Switching to ChatGPT later only requires changing `LLM_PROVIDER=chatgpt` and
providing `LLM_API_KEY` or `OPENAI_API_KEY`. Request timeout, retry count,
retry backoff, and HTTP connection pooling are configured with
`LLM_REQUEST_TIMEOUT`, `LLM_MAX_RETRIES`, `LLM_RETRY_BACKOFF`,
`LLM_MAX_IDLE_CONNS`, `LLM_MAX_IDLE_CONNS_PER_HOST`, and
`LLM_MAX_CONNS_PER_HOST`. GitHub API access uses `GITHUB_TOKEN`, and generated
Markdown reports are written under `PR_REVIEW_OUTPUT_DIR` (`reports` by
default).
