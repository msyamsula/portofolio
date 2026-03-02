# Database Layer

The Database Layer provides database clients and connection management.

## Architecture

```mermaid
flowchart TB
    subgraph Database[Database Layer]
        PG[PostgreSQL Client]
        Redis[Redis Client]
    end

    subgraph Domains[Using Domains]
        URL[URL Shortener]
        Friend[Friend]
        Message[Message]
    end

    PG --> URL
    PG --> Friend
    PG --> Message
    Redis --> URL

    PG -.->|connection| PGSrv[(PostgreSQL Server)]
    Redis -.->|connection| RedisSrv[(Redis Server)]
```

## Components

| Component | Location | Purpose |
|-----------|-----------|---------|
| PostgreSQL Client | `postgres/` | PostgreSQL connectivity |
| Redis Client | `redis/` | Redis connectivity |
| Migrations | `migration/` | Schema versioning |

## Features

- Connection pooling
- Health checking
- Query execution
- Transaction support
- Cache-first lookups

## Related

- [infrastructure/database/postgres/README.md](PostgreSQL Client)
- [infrastructure/database/redis/README.md](Redis Client)
- [[docs/repository-pattern.md|Repository Pattern]]
