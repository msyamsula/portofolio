# Graph Domain

The Graph domain provides graph algorithm implementations.

## Purpose

Expose common graph algorithms as HTTP endpoints.

## Architecture

```mermaid
flowchart TB
    subgraph Graph[Graph Domain]
        Handler[HTTP Handler]
        Service[Service]
    end

    subgraph Algorithms[Algorithms]
        DFS[DFS]
        BFS[BFS]
        Cycle[Cycle Detection]
        SCC[SCC]
        Topo[Topological Sort]
    end

    Handler --> Service
    Service --> DFS
    Service --> BFS
    Service --> Cycle
    Service --> SCC
    Service --> Topo

    Handler -.->|metrics, tracing| Telemetry[Telemetry]
    Service -.->|tracing| Telemetry
```

## Storage

- **None**: Pure algorithms, no database dependency

## Components

| Component | Location | Responsibility |
|-----------|-----------|----------------|
| DTO | `dto/` | Graph data structures |
| Handler | `handler/` | HTTP request handling |
| Service | `service/` | Algorithm implementations |

## Algorithms

| Algorithm | Description | Complexity |
|------------|-------------|-------------|
| DFS | Depth-First Search | O(V + E) |
| BFS | Breadth-First Search | O(V + E) |
| Cycle Detection | Detect cycles in graph | O(V + E) |
| SCC | Strongly Connected Components | O(V + E) |
| Articulation Points | Find critical vertices | O(V + E) |
| Eulerian Paths | Paths using each edge once | O(V + E) |
| Topological Sort | Linear order in DAG | O(V + E) |

## Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Handler
    participant Service
    participant Algorithm

    Client->>Handler: POST /graph/dfs {graph, start}
    Handler->>Handler: Validate input
    Handler->>Service: DFS(ctx, graph, start)
    Service->>Algorithm: ExecuteDFS()
    Algorithm-->>Service: result
    Service-->>Handler: result
    Handler-->>Client: 200 OK {result}
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/graph/dfs` | Depth-First Search |
| POST | `/graph/bfs` | Breadth-First Search |
| POST | `/graph/cycle` | Detect cycles |
| POST | `/graph/scc` | Strongly Connected Components |
| POST | `/graph/articulation` | Articulation Points |
| POST | `/graph/eulerian` | Eulerian Paths |
| POST | `/graph/topological` | Topological Sort |

## Related

- Domain Services
- Graph Algorithms
