# Request Flow

The request flow shows how an HTTP request travels through the architecture.

## Request Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant Client as HTTP Client
    participant MW as Middleware Chain
    participant Handler as Domain Handler
    participant Service as Domain Service
    participant Repo as Repository
    participant DB as PostgreSQL
    participant Cache as Redis

    Client->>MW: HTTP Request
    Note over MW: Metrics, Tracing, CORS, Recovery, Logging
    MW->>Handler: Route to handler
    Handler->>Handler: Parse & validate
    Handler->>Service: Call business logic
    Service->>Repo: Find data
    Repo->>Cache: Check cache

    alt Cache Hit
        Cache-->>Repo: Found
        Repo-->>Service: Data (cached)
    else Cache Miss
        Cache-->>Repo: Not found
        Repo->>DB: Query
        DB-->>Repo: Data
        Repo->>Cache: Set cache
        Repo-->>Service: Data (from DB)
    end

    Service-->>Handler: Result
    Handler->>Handler: Format response
    Handler->>MW: Return response
    MW-->>Client: HTTP Response
```

## Middleware Chain

```mermaid
flowchart LR
    Request[Incoming Request] --> Metrics[Metrics]
    Metrics --> Tracing[Tracing]
    Tracing --> CORS[CORS]
    CORS --> Recovery[Recovery]
    Recovery --> Logging[Logging]
    Logging --> ContentType[Content-Type]
    ContentType --> ResponseTime[Response Time]
    ResponseTime --> Handler[Domain Handler]
    Handler --> Response[Outgoing Response]
```

## Cache-Aside Flow

For URL Shortener requests:

```mermaid
flowchart TB
    App[Application] -->|1. Check Cache| Cache[Redis Cache]
    Cache -->|2. Cache Hit?| Decision{Found?}

    Decision -->|Yes| Return[Return Cached Data]
    Decision -->|No| DB[PostgreSQL]

    DB -->|3. Query| Result[Result]
    Result -->|4. Update Cache| Cache
    Cache -->|5. Return| Return
```

## Related

- [[docs/architecture-overview.md|Architecture Overview]]
- [infrastructure/http/middleware/README.md](HTTP Middleware)
- [[docs/cache-aside-pattern.md|Cache-Aside Pattern]]
