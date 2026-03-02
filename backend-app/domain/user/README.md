# User Domain

The User domain handles authentication and authorization.

## Purpose

Authenticate users using Google OAuth and issue JWT tokens.

## Architecture

```mermaid
flowchart TB
    subgraph User[User Domain]
        Handler[HTTP Handler]
        Service[Service]
        Integration[External Integration]
    end

    subgraph External[External Services]
        Google[Google OAuth]
    end

    Handler --> Service
    Service --> Integration
    Integration --> Google

    Handler -.->|metrics, tracing| Telemetry[Telemetry]
    Service -.->|tracing| Telemetry
```

## Storage

- **None**: Uses external OAuth provider

## Components

| Component | Location | Responsibility |
|-----------|-----------|----------------|
| DTO | `dto/` | User data structures |
| Handler | `handler/` | HTTP request handling |
| Service | `service/` | Authentication logic |
| Integration | `integration/` | Google OAuth client |

## OAuth Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Google
    participant JWT

    Client->>Handler: GET /user/google/auth
    Handler->>Service: GetAuthURL()
    Service->>Google: Generate auth URL
    Google-->>Service: auth_url
    Service-->>Handler: auth_url
    Handler-->>Client: Redirect to Google

    Client->>Google: Authorize
    Google-->>Client: Redirect with code
    Client->>Handler: GET /user/google/callback?code=xxx
    Handler->>Service: ExchangeCodeForToken(ctx, code)
    Service->>Google: Exchange code for token
    Google-->>Service: access_token
    Service->>Service: Get user info
    Service->>JWT: Generate app token
    JWT-->>Service: jwt_token
    Service-->>Handler: jwt_token
    Handler-->>Client: 200 OK {token}
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user/google/auth` | Start OAuth flow |
| GET | `/user/google/callback` | OAuth callback |
| POST | `/user/token/validate` | Validate JWT token |

## Features

- Google OAuth 2.0 authentication
- JWT token generation
- Token validation
- Configurable token TTL

## Related

- OAuth 2.0
- JWT Authentication
- External Integrations
