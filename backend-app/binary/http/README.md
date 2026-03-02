# HTTP Server

The HTTP Server is the entry point for HTTP requests.

## Architecture

```mermaid
flowchart TB
    subgraph Server[HTTP Server]
        Config[Configuration]
        Init[Initialization]
        Router[Gorilla Mux Router]
    end

    subgraph Telemetry[Telemetry]
        Logger[Logger]
        Metrics[Metrics]
        Tracer[Tracer]
    end

    subgraph Infra[Infrastructure]
        DB[PostgreSQL]
        Cache[Redis]
    end

    subgraph Domains[Domains]
        URL[URL Shortener]
        Friend[Friend]
        Message[Message]
        User[User]
        Graph[Graph]
        Health[Healthcheck]
    end

    Config --> Init
    Init --> Router
    Init --> Logger
    Init --> Metrics
    Init --> Tracer
    Init --> DB
    Init --> Cache

    Router --> URL
    Router --> Friend
    Router --> Message
    Router --> User
    Router --> Graph
    Router --> Health
```

## Startup Sequence

```mermaid
sequenceDiagram
    participant Main
    participant Config
    participant Logger
    participant Telemetry
    participant Infra
    participant Router
    participant Server

    Main->>Config: Load configuration
    Config-->>Main: Config

    Main->>Logger: Initialize logger
    Logger-->>Main: Ready

    Main->>Telemetry: Initialize tracing/metrics
    Telemetry-->>Main: Ready

    Main->>Infra: Connect DB/Cache
    Infra-->>Main: Connected

    Main->>Router: Setup routes
    Router->>Router: Register domain handlers

    Main->>Server: Start HTTP server
    Server-->>Main: Listening
```

## Configuration

Environment variables (`.env`):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 5000 | HTTP server port |
| `BASE_URL` | https://short.est | Base URL for shortened links |
| `POSTGRES_HOST` | localhost | PostgreSQL host |
| `POSTGRES_PORT` | 5432 | PostgreSQL port |
| `REDIS_HOST` | localhost:6379 | Redis address |
| `SERVICE_NAME` | backend-app | Service identifier |
| `LOG_LEVEL` | INFO | Logging level |
| `LOG_FORMAT` | TEXT | Log format |
| `APP_TOKEN_TTL` | 24h | JWT token expiration |

## Related

- [infrastructure/http/README.md](HTTP Infrastructure)
- [infrastructure/http/middleware/README.md](HTTP Middleware)
- [[docs/request-flow.md|Request Flow]]
