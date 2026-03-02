# Distributed Tracing

Distributed tracing tracks request flow across components.

## Architecture

```mermaid
flowchart TB
    subgraph Tracing[Distributed Tracing]
        Tracer[Tracer]
        Span[Span Manager]
    end

    subgraph Domains[Using Domains]
        URL[URL Shortener]
        Friend[Friend]
        Message[Message]
        User[User]
        Graph[Graph]
        Health[Healthcheck]
    end

    subgraph Export[Export]
        Jaeger[Jaeger]
    end

    Tracer --> Span
    Span --> Jaeger

    URL --> Tracer
    Friend --> Tracer
    Message --> Tracer
    User --> Tracer
    Graph --> Tracer
    Health --> Tracer
```

## Features

- Trace context propagation
- Hierarchical spans (parent-child)
- Attribute and event annotation
- Error tracking

## Usage

```go
tracer := otel.Tracer("service-name")
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

span.SetAttributes(attribute.String("key", "value"))
span.AddEvent("event-name")
```

## Span Hierarchy

```mermaid
sequenceDiagram
    participant Handler
    participant Service
    participant Repo

    Handler->>Service: Call
    activate Service

    Service->>Service: Create span
    Service->>Repo: Query
    activate Repo

    Repo-->>Service: Result
    deactivate Repo

    Service->>Service: Process
    Service-->>Handler: Response
    deactivate Service
```

## Export

- **Tool**: OpenTelemetry
- **Visualization**: Jaeger
- **Endpoint**: `http://localhost:16686`

## Related

- [infrastructure/telemetry/README.md](Telemetry Stack)
- Jaeger
- [[docs/architecture-overview.md|Observability]]
