# Message Domain

The Message domain handles conversation storage and retrieval.

## Purpose

Store and retrieve messages between users.

## Architecture

```mermaid
flowchart TB
    subgraph Message[Message Domain]
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

- **Primary**: [infrastructure/database/postgres/README.md](PostgreSQL) - Messages and conversations

## Components

| Component | Location | Responsibility |
|-----------|-----------|----------------|
| DTO | `dto/` | Message data structures |
| Handler | `handler/` | HTTP request handling |
| Service | `service/` | Message business logic |
| Repository | `repository/` | Message data access |

## Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Repo
    participant DB

    Client->>Handler: POST /message/send {to, content}
    Handler->>Handler: Validate input
    Handler->>Service: SendMessage(ctx, from, to, content)
    Service->>Repo: Save(ctx, message)
    Repo->>DB: INSERT INTO messages
    DB-->>Repo: message
    Repo-->>Service: message
    Service-->>Handler: success
    Handler-->>Client: 201 Created

    Client->>Handler: GET /message/conversations
    Handler->>Service: ListConversations(ctx, userID)
    Service->>Repo: FindByUserID(ctx, userID)
    Repo->>DB: SELECT * FROM messages
    DB-->>Repo: conversations
    Repo-->>Service: conversations
    Service-->>Handler: conversations
    Handler-->>Client: 200 OK {conversations}
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/message/send` | Send message |
| GET | `/message/conversations` | List conversations |
| GET | `/message/{conversation_id}` | Get conversation history |

## Features

- Send messages
- Retrieve conversation history
- List user conversations

## Related

- [[docs/repository-pattern.md|Repository Pattern]]
- [[domain/url-shortener/README.md|Domain Services]]
