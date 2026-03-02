# Design Principles

The backend app follows these core design principles to maintain a clean, maintainable codebase.

## Principles Overview

```mermaid
mindmap
  root((Design Principles))
    Domain Isolation
      Bounded Contexts
      Independent Development
      Minimal Coupling
    Dependency Inversion
      Interface Abstractions
      Test Mocks
      Swappable Implementations
    Clear Layering
      Presentation
      Application
      Persistence
      Infrastructure
    Infrastructure Reuse
      Shared Components
      Consistent Behavior
      Centralized Updates
    Protocol Agnostic
      HTTP
      gRPC (future)
      Message Queue (future)
```

## 1. Domain Isolation

Each domain is a bounded context with clear boundaries.

### Benefits

- Clear ownership
- Reduced cognitive load
- Easier testing
- Parallel development

### Example

- URL Shortener doesn't need to know about Friend
- User doesn't need to know about Message
- Each domain can evolve independently

## 2. Dependency Inversion

High-level modules don't depend on low-level modules. Both depend on abstractions.

### Implementation

```mermaid
flowchart TB
    subgraph Domain[Domain Layer]
        Service[Service]
        RepoIface[Repository Interface]
    end

    subgraph Infra[Infrastructure Layer]
        PGRepo[PostgreSQL Repo]
        RedisRepo[Redis Repo]
    end

    Service -->|depends on| RepoIface
    PGRepo -->|implements| RepoIface
    RedisRepo -->|implements| RepoIface
```

## 3. Clear Layering

Each component has a single, well-defined responsibility.

### Layers

```mermaid
flowchart LR
    Entry[Entry Points] --> Pres[Presentation Layer]
    Pres --> App[Application Layer]
    App --> Pers[Persistence Layer]
    Pers --> Infra[Infrastructure Layer]
```

## 4. Infrastructure Reuse

Technical components are shared across all domains.

### Shared Components

- HTTP Middleware
- Database Clients
- Telemetry Stack
- Response Utilities

## 5. Protocol Agnostic

Business logic is independent of the transport protocol.

### Benefits

- Can add new protocols without changing business logic
- Services can be reused across different entry points
- Easier testing of business logic

## Related

- [[docs/architecture-overview.md|Architecture Overview]]
- [[docs/code-structure.md|Code Structure]]
- [[docs/clean-architecture.md|Clean Architecture]]
