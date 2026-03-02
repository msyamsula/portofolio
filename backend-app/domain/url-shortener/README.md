# URL Shortener Domain

The URL Shortener domain handles URL shortening and expansion functionality.

## Purpose

Convert long URLs into short, shareable codes and redirect back to the original URLs.

## Architecture

```mermaid
flowchart TB
    subgraph URL_Shortener[URL Shortener Domain]
        Handler[HTTP Handler]
        Service[Service]
        Repo[Repository Interface]
    end

    subgraph Storage[Storage Layer]
        PG[PostgreSQL]
        Redis[Redis]
    end

    Handler --> Service
    Service --> Repo
    Repo --> PG
    Repo --> Redis

    Handler -.->|metrics, tracing| Telemetry[Telemetry]
    Service -.->|tracing| Telemetry
```

## Storage

- **Primary**: [infrastructure/database/postgres/README.md](PostgreSQL) - Persistent storage for URL mappings
- **Cache**: [infrastructure/database/redis/README.md](Redis) - Cache-aside pattern for performance

## Components

| Component | Location | Responsibility |
|-----------|-----------|----------------|
| DTO | `dto/` | Request/response contracts |
| Handler | `handler/` | HTTP request handling |
| Service | `service/` | Business logic |
| Repository | `repository/` | Data access abstraction |

## Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repo
    participant Cache
    participant DB

    Client->>Handler: POST /url/shorten {longURL}
    Handler->>Handler: Validate input
    Handler->>Service: Shorten(ctx, longURL)
    Service->>Repo: FindByLongURL(ctx, longURL)

    alt Cache Hit
        Cache-->>Repo: mapping found
        Repo-->>Service: shortURL (cached)
    else Cache Miss
        Cache-->>Repo: not found
        Service->>Service: Generate short code
        Service->>Repo: Save(ctx, shortCode, longURL)
        Repo->>Cache: Set(shortCode, mapping)
        Repo->>DB: INSERT INTO url_mappings
        DB-->>Repo: success
        Repo-->>Service: success
        Service-->>Handler: shortURL (new)
    end

    Handler-->>Client: 201 Created {shortURL}
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/url/shorten` | Create short URL |
| GET | `/url/{shortCode}` | Redirect to original URL |

## Features

- SHA256-based short code generation
- Cache-first lookup for performance
- 301 permanent redirects
- Duplicate URL detection

## Related

- [[docs/repository-pattern.md|Repository Pattern]]
- [[docs/cache-aside-pattern.md|Cache-Aside Pattern]]
- [[docs/request-flow.md|Request Flow]]
- [[infrastructure/http/middleware/README.md|HTTP Middleware]]
