# Clean Architecture

Clean architecture principles guide the backend app structure.

## Layer Separation

```mermaid
flowchart TB
    subgraph Layers[Clean Architecture Layers]
        subgraph Entry[Entry Points]
            HTTP[HTTP Server]
        end

        subgraph Business[Business Domains]
            Handler[Handlers]
            Service[Services]
            Repo[Repositories]
        end

        subgraph Infra[Infrastructure]
            DB[Database]
            Cache[Cache]
            HTTPInfra[HTTP Components]
        end
    end

    HTTP --> Handler
    Handler --> Service
    Service --> Repo

    Repo -->|Interface| DB
    Repo -->|Interface| Cache
    Handler -->|Uses| HTTPInfra
```

## Dependency Rule

Dependencies point **inward**:

```mermaid
flowchart LR
    Entry[Entry Points] --> Business[Business Logic]
    Business --> Infra[Infrastructure Interfaces]

    Infra -.->|implements| Impl[Infrastructure Implementations]
```

## Benefits

- Framework independent
- Testable
- UI independent
- Database independent
- External services independent

## Related

- [[docs/architecture-overview.md|Architecture Overview]]
- [[docs/dependency-inversion.md|Dependency Inversion]]
- [[docs/domain-driven-design.md|Domain-Driven Design]]
