# Friend Domain

The Friend domain manages user friendship relationships.

## Purpose

Add, remove, and list friend connections between users.

## Architecture

```mermaid
flowchart TB
    subgraph Friend[Friend Domain]
        Handler[HTTP Handler]
        Service[Service]
        Repo[Repository Interface]
    end

    subgraph Storage[Storage Layer]
        PG[PostgreSQL]
    end

    Handler --> Service
    Service --> Repo
    Repo --> PG

    Handler -.->|metrics, tracing| Telemetry[Telemetry]
    Service -.->|tracing| Telemetry
```

## Storage

- **Primary**: [infrastructure/database/postgres/README.md](PostgreSQL) - Friend relationships

## Components

| Component | Location | Responsibility |
|-----------|-----------|----------------|
| DTO | `dto/` | Friend data structures |
| Handler | `handler/` | HTTP request handling |
| Service | `service/` | Friend business logic |
| Repository | `repository/` | Friend data access |

## Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repo
    participant DB

    Client->>Handler: POST /friend/add {friend_id}
    Handler->>Handler: Validate input
    Handler->>Service: AddFriend(ctx, userID, friendID)
    Service->>Repo: Save(ctx, friendship)
    Repo->>DB: INSERT INTO friendships
    DB-->>Repo: success
    Repo-->>Service: friendship
    Service-->>Handler: success
    Handler-->>Client: 201 Created

    Client->>Handler: GET /friend/list
    Handler->>Service: ListFriends(ctx, userID)
    Service->>Repo: FindByUserID(ctx, userID)
    Repo->>DB: SELECT * FROM friendships
    DB-->>Repo: friends
    Repo-->>Service: friends
    Service-->>Handler: friends
    Handler-->>Client: 200 OK {friends}
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/friend/add` | Add friend |
| DELETE | `/friend/remove` | Remove friend |
| GET | `/friend/list` | List user's friends |

## Features

- Add friend connections
- Remove friend connections
- List user's friends

## Related

- [[docs/repository-pattern.md|Repository Pattern]]
- [[domain/url-shortener/README.md|Domain Services]]
