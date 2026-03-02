# Redis Client

The Redis Client provides Redis cache connectivity.

## Architecture

```mermaid
flowchart TB
    subgraph RedisClient[Redis Client]
        Config[Redis Config]
        Client[Client Interface]
    end

    subgraph Repositories[Repositories]
        URL[URL Shortener Repo]
    end

    Config --> Client
    Client --> URL

    Client -.->|connections| RedisSrv[(Redis Server)]
```

## Features

- Get/Set operations
- Health checking (ping)
- Expiration handling
- Connection pooling

## Cache-Aside Pattern

```mermaid
sequenceDiagram
    participant App as Application
    participant Cache as Redis Cache
    participant DB as PostgreSQL

    App->>Cache: Get(key)
    alt Cache Hit
        Cache-->>App: value
    else Cache Miss
        Cache-->>App: not found
        App->>DB: Query(key)
        DB-->>App: value
        App->>Cache: Set(key, value)
        Cache-->>App: success
    end
```

## Usage

```go
type Client interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Del(ctx context.Context, keys ...string) error
}
```

## Used By

- [domain/url-shortener/README.md](URL Shortener) - URL mapping cache

## Related

- [infrastructure/database/README.md](Database Layer)
- [[docs/cache-aside-pattern.md|Cache-Aside Pattern]]
- [infrastructure/database/postgres/README.md](PostgreSQL Client)
