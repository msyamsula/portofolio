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
not-yet-due jobs, atomically claims due jobs from Redis, and dispatches them to
Kafka.

The web dropdown period options come from `JOB_PERIOD_OPTIONS`, for example
`1s,10m,5h`. Periodic interval state is persisted in Postgres on each GitHub
link; Redis only stores the next-claim sorted set used by the orchestrator.

The `job-consumer` container runs the PR review agent. It consumes from `pr_review`
and processes PRs in the following workflow:

1. **Consume** — Read a message from Kafka (keyed by GitHub URL, partitioned for order).
2. **Lock** — Acquire a distributed Redis lock (`lock:pr:<github_url>`, TTL 5m) to prevent
   concurrent processing of the same PR across partitions or containers.
3. **Fetch** — Pull the PR discussion (title, description, issue comments, review
   summaries, inline review comments) from GitHub via the API. Gets the feature
   branch name.
4. **Prepare worktree** — `git fetch origin`, force-update local branch to remote,
   remove any stale worktree for the same branch, `git worktree add` from the
   feature branch, `git merge origin/develop`.
5. **Apply fixes** — Run the LLM agent (DeepSeek default) with tool-calling:
   - Survey the PR discussion and relevant source files
   - Apply targeted code fixes via `write_file`
   - Validate with `make test` and `make lint-gci-new`
   - Produce a Markdown report with commit message and diff summary
6. **Commit** — `git add -A && git commit --no-verify` (no push, manual review).
7. **Report** — Write the agent report Markdown to `reports/agent/`.
8. **Reschedule** — Re-set the periodic job in Redis for the next review cycle.
9. **Release** — Release the distributed lock.

On retry (after transient failures), the agent reads the previous JSONL trace file
to avoid repeating work, and stale worktree paths in the trace are rewritten to the
current workspace directory.

The agent supports up to `PR_REVIEW_AGENT_MAX_TURNS` (default 50) LLM turns per
review. Each turn can include multiple tool calls (read_file, write_file,
run_command, etc.).

### SSH for private repos

The agent container needs SSH to pull/fetch from private GitHub repositories.
Copy your SSH key and config to `privatekey/ssh/`. Example:

```bash
cp ~/.ssh/id_ed25519_astronautsid privatekey/ssh/
chmod 600 privatekey/ssh/id_ed25519_astronautsid
```

Create `privatekey/ssh/config`:
```
Host github-astronautsid
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519_astronautsid
    IdentitiesOnly yes
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

The repo's git remote must use the SSH alias:
```bash
git remote set-url origin git@github-astronautsid:owner/repo.git
```

Files are mounted read-only to `/home/appuser/.ssh` in the container. Not needed
if the repo uses HTTPS.

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
- `make kafka-topic-offsets`
- `make kafka-dlq-topic-offsets`
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

Check only GitHub PR access without running the LLM review:

```bash
go run ./scripts/check-pr-access https://github.com/astronautsid/astro-ads-be/pull/659
```

Both scripts load `.env` and use `GITHUB_TOKEN_PATH` for GitHub access. The
token file is read for each GitHub request, so refreshing the file does not
require restarting the consumer.

LLM review calls are selected through one provider interface. The current default is DeepSeek:

- `LLM_PROVIDER=deepseek`
- `DEEPSEEK_API_KEY`
- `DEEPSEEK_MODEL` (`deepseek-v4-pro`)

Switching to Claude or ChatGPT only requires changing `LLM_PROVIDER` and
providing the corresponding API key. Request timeout, retry count,
retry backoff, and HTTP connection pooling are configured with
`LLM_REQUEST_TIMEOUT`, `LLM_MAX_RETRIES`, `LLM_RETRY_BACKOFF`,
`LLM_MAX_IDLE_CONNS`, `LLM_MAX_IDLE_CONNS_PER_HOST`, and
`LLM_MAX_CONNS_PER_HOST`.

GitHub API access uses `GITHUB_TOKEN_PATH`. Generated Markdown reports are
written under `PR_REVIEW_OUTPUT_DIR` (`reports` by default).
