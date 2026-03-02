# HTTP Infrastructure

The HTTP Infrastructure provides HTTP server and middleware components.

## Architecture

```mermaid
flowchart TB
    subgraph HTTPInfra[HTTP Infrastructure]
        Server[HTTP Server]
        Middleware[Middleware Chain]
        HandlerUtils[Handler Utilities]
    end

    subgraph Domains[Using Domains]
        URL[URL Shortener]
        Friend[Friend]
        Message[Message]
        User[User]
        Graph[Graph]
        Health[Healthcheck]
    end

    Server --> Middleware
    Middleware --> URL
    Middleware --> Friend
    Middleware --> Message
    Middleware --> User
    Middleware --> Graph
    Middleware --> Health
```

## Components

| Component | Location | Purpose |
|-----------|-----------|---------|
| HTTP Server | `server.go` | Server configuration |
| Middleware | `middleware/` | Cross-cutting HTTP concerns |
| Handler Utils | `handler/` | Shared response utilities |

## Middleware Chain

```mermaid
flowchart LR
    Request[Request] --> MW1[Metrics]
    MW1 --> MW2[Tracing]
    MW2 --> MW3[CORS]
    MW3 --> MW4[Recovery]
    MW4 --> MW5[Logging]
    MW5 --> MW6[Content-Type]
    MW6 --> Handler[Handler]
    Handler --> Response[Response]
```

## Middleware Order

| Order | Middleware | Purpose |
|--------|------------|---------|
| 1 | Metrics | Record request count/latency |
| 2 | Tracing | Create distributed trace spans |
| 3 | CORS | Handle cross-origin requests |
| 4 | Recovery | Panic recovery |
| 5 | Logging | Request/response logging |
| 6 | Content-Type | Enforce JSON |

## Related

- [binary/http/README.md](HTTP Server)
- [infrastructure/http/middleware/README.md](HTTP Middleware)
- [domain/url-shortener/README.md](HTTP Handlers)
