# Cache-Aside Pattern

The Cache-Aside Pattern is used for performance optimization.

## How It Works

```mermaid
flowchart TB
    App[Application] -->|1. Check Cache| Cache[Cache]
    Cache -->|2. Cache Hit?| Decision{Found?}

    Decision -->|Yes| Return[Return Cached Data]
    Decision -->|No| DB[Database]

    DB -->|3. Query| Result[Result]
    Result -->|4. Update Cache| Cache
    Cache -->|5. Return| Return
```

## Request Flow

```mermaid
sequenceDiagram
    participant App
    participant Cache
    participant DB

    App->>Cache: Get(key)
    alt Cache Hit
        Cache-->>App: Data
    else Cache Miss
        Cache-->>App: Not Found
        App->>DB: Query(key)
        DB-->>App: Data
        App->>Cache: Set(key, data)
        Cache-->>App: Success
    end
```

## Implementation

Used in URL Shortener:

1. Check cache for existing short code
2. On miss, query PostgreSQL
3. Update cache for next request

## Benefits

- Reduces database load
- Improves response time
- Scales read-heavy workloads

## Related

- [infrastructure/database/postgres/README.md](PostgreSQL Client)
- [infrastructure/database/redis/README.md](Redis Client)
- [[docs/request-flow.md|Request Flow]]
