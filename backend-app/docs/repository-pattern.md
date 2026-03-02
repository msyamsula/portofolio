# Repository Pattern

The Repository Pattern abstracts data access in the backend app.

## Purpose

- Separate domain logic from data access
- Enable easy testing through mocking
- Allow swapping storage implementations

## Implementation

### Interface (Domain Layer)

```mermaid
flowchart TB
    subgraph Domain[Domain Layer]
        Service[Service]
        RepoIface[Repository Interface]
    end

    Service -->|uses| RepoIface
```

### Implementation (Infrastructure Layer)

```mermaid
flowchart TB
    subgraph Infra[Infrastructure Layer]
        PGRepo[PostgreSQL Repository]
        RedisRepo[Redis Repository]
    end

    PGRepo -->|implements| RepoIface[Repository Interface]
    RedisRepo -->|implements| RepoIface
```

## Example Interface

```go
type Repository interface {
    FindByID(ctx context.Context, id string) (*Entity, error)
    FindAll(ctx context.Context) ([]*Entity, error)
    Save(ctx context.Context, entity *Entity) error
    Delete(ctx context.Context, id string) error
}
```

## Usage

```mermaid
sequenceDiagram
    participant Service as Domain Service
    participant Repo as Repository Interface
    participant PG as PostgreSQL

    Service->>Repo: FindByID(id)
    Repo->>PG: SELECT * FROM entities
    PG-->>Repo: entity
    Repo-->>Service: entity
```

## Benefits

- Easy testing with [mock/README.md](mocks)
- Swappable implementations
- Clear contract between layers

## Related

- [[docs/dependency-inversion.md|Dependency Inversion]]
- [domain/url-shortener/README.md](Domain Services)
- [mock/README.md](Test Mocks)
