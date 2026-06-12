# auto-pr-review

Small Go HTTP API for registering GitHub links in Postgres.

## Run locally

```bash
make env
make up
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

Configuration is read from environment variables at process startup. Changing
configuration requires restarting the service container.
