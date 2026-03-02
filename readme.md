# Backend-App Documentation

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Code Architecture Diagram](#code-architecture-diagram)
3. [Folder Structure](#folder-structure)
4. [Technology Stack](#technology-stack)
5. [Observability](#observability)
6. [Infrastructure & Domain Connections](#infrastructure--domain-connections)
7. [Local Development Guide](#local-development-guide)

---

## Architecture Overview

The backend-app follows a **monolithic architecture with domain-driven design**. All services run in a single process but are organized by business domain. This approach provides:

- **Simplicity**: Single deployment unit
- **Performance**: In-memory communication between domains
- **Maintainability**: Clear domain boundaries
- **Scalability**: Can be split into microservices if needed

### Key Architectural Patterns

| Pattern | Description | Location |
|---------|-------------|----------|
| **Layered Architecture** | Handler → Service → Repository | All domains |
| **Repository Pattern** | Data access abstraction | `domain/*/repository` |
| **Cache-Aside Pattern** | Redis + PostgreSQL | All domains |
| **Middleware Chain** | Cross-cutting concerns | `infrastructure/http/middleware` |
| **Dependency Injection** | Constructor-based injection | All handlers/services |

---

## Code Architecture Diagram

```mermaid
graph TB
    subgraph "Entry Point"
        HTTP[HTTP Request/Response<br/>Gorilla Mux]
    end

    subgraph "Middleware Chain"
        MW_RECOVERY[Recovery]
        MW_CORS[CORS]
        MW_LOGGING[Logging]
        MW_METRICS[Metrics]
        MW_TRACING[Tracing]
        MW_AUTHN[Authn]
    end

    subgraph "Domain Handlers"
        H1[url-shortener<br/>handler]
        H2[user<br/>handler]
        H3[friend<br/>handler]
        H4[message<br/>handler]
        H5[graph<br/>handler]
        H6[websocket<br/>handler]
        H7[healthcheck<br/>handler]
    end

    subgraph "Domain Services"
        S1[url-shortener<br/>service]
        S2[user<br/>service]
        S3[friend<br/>service]
        S4[message<br/>service]
        S5[graph<br/>service]
        S6[websocket<br/>service]
        S7[healthcheck<br/>service]
    end

    subgraph "Domain Repositories"
        R1[url-shortener<br/>repository]
        R2[user<br/>repository]
        R3[friend<br/>repository]
        R4[message<br/>repository]
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Persistent Store)]
        REDIS[(Redis<br/>Cache Layer)]
    end

    subgraph "Infrastructure Layer"
        TEL[Telemetry/OTel<br/>Tracing/Metrics/Logs]
        LOG[Logger<br/>Structured]
        DB[Database<br/>Postgres/Redis]
    end

    subgraph "External Services"
        OTEL[OTLP Collector<br/>Jaeger/Prometheus/Loki]
        EXT[External Services<br/>Google OAuth/AWS SQS-SNS/DynamoDB]
    end

    HTTP --> MW_RECOVERY
    MW_RECOVERY --> MW_CORS
    MW_CORS --> MW_LOGGING
    MW_LOGGING --> MW_METRICS
    MW_METRICS --> MW_TRACING
    MW_TRACING --> MW_AUTHN

    MW_AUTHN --> H1
    MW_AUTHN --> H2
    MW_AUTHN --> H3
    MW_AUTHN --> H4
    MW_AUTHN --> H5
    MW_AUTHN --> H6
    MW_AUTHN --> H7

    H1 --> S1
    H2 --> S2
    H3 --> S3
    H4 --> S4
    H5 --> S5
    H6 --> S6
    H7 --> S7

    S1 --> R1
    S2 --> R2
    S3 --> R3
    S4 --> R4

    R1 --> PG
    R1 --> REDIS
    R2 --> PG
    R2 --> REDIS
    R3 --> PG
    R3 --> REDIS
    R4 --> PG
    R4 --> REDIS

    S1 -.-> TEL
    S2 -.-> TEL
    S3 -.-> TEL
    S4 -.-> TEL

    TEL --> OTEL
    S1 -.-> EXT
    S2 -.-> EXT

    classDef entry fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef middleware fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef handler fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef service fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef repository fill:#fce4ec,stroke:#880e4f,stroke-width:2px
    classDef data fill:#b3e5fc,stroke:#0277bd,stroke-width:2px
    classDef infra fill:#fff9c4,stroke:#f57f17,stroke-width:2px
    classDef external fill:#c8e6c9,stroke:#33691e,stroke-width:2px

    class HTTP entry
    class MW_RECOVERY,MW_CORS,MW_LOGGING,MW_METRICS,MW_TRACING,MW_AUTHN middleware
    class H1,H2,H3,H4,H5,H6,H7 handler
    class S1,S2,S3,S4,S5,S6,S7 service
    class R1,R2,R3,R4 repository
    class PG,REDIS data
    class TEL,LOG,DB infra
    class OTEL,EXT external
```

### Request Flow Example (URL Shortener)

```mermaid
sequenceDiagram
    participant Client
    participant MW as Middleware Chain
    participant Handler as URL Handler
    participant Service as URL Service
    participant Repo as URL Repository
    participant Cache as Redis Cache
    participant DB as PostgreSQL

    Client->>MW: POST /url/shorten
    activate MW

    MW->>MW: Recovery: Catch panics
    MW->>MW: CORS: Handle cross-origin
    MW->>MW: Logging: Log request start
    MW->>MW: Metrics: Record request count
    MW->>MW: Tracing: Create span

    MW->>Handler: ParseRequest()
    activate Handler
    Handler->>Handler: Extract URL from JSON body
    Handler-->>MW: Return parsed request
    deactivate Handler

    MW->>Service: ShortenURL()
    activate Service
    Service->>Service: Validate URL
    Service->>Service: Generate short code

    Service->>Repo: Store()
    activate Repo

    Repo->>Cache: Redis.Set()
    Cache-->>Repo: Success
    Repo->>DB: PostgreSQL.Exec()
    DB-->>Repo: Success

    Repo-->>Service: Success
    deactivate Repo

    Service-->>MW: Return shortened URL
    deactivate Service

    MW->>MW: Logging: Log request complete
    MW->>MW: Metrics: Update metrics
    MW->>MW: Tracing: Close span

    MW-->>Client: Return JSON with shortened URL
    deactivate MW
```

---

## Folder Structure

```
backend-app/
├── binary/                          # Application entry points
│   ├── main.go                     # Main binary with all domains
│   └── http/                       # HTTP-specific binary
│       ├── main.go                 # HTTP server entry point
│       └── docs/                   # Swagger documentation
│
├── domain/                          # Business logic domains
│   ├── url-shortener/              # URL shortening service
│   │   ├── dto/                    # Data transfer objects
│   │   │   └── url.go             # URL structs & mappings
│   │   ├── handler/                # HTTP handlers
│   │   │   └── url.go             # Request/response handling
│   │   ├── service/                # Business logic
│   │   │   └── url.go             # URL shortening logic
│   │   └── repository/             # Data access
│   │       ├── url.go             # Repository interface
│   │       ├── cache.go           # Redis implementation
│   │       └── persistent.go      # PostgreSQL implementation
│   │
│   ├── user/                       # User authentication service
│   │   ├── dto/                    # User DTOs
│   │   ├── handler/                # User HTTP handlers
│   │   ├── service/                # User business logic
│   │   ├── repository/             # User data access
│   │   └── integration/            # External integrations
│   │       └── google_oauth.go    # Google OAuth integration
│   │
│   ├── friend/                     # Friend management service
│   ├── graph/                      # Graph visualization service
│   ├── message/                    # Message handling service
│   └── healthcheck/                # Health monitoring service
│
├── infrastructure/                  # Shared infrastructure
│   │
│   ├── database/                   # Database abstractions
│   │   ├── postgres/               # PostgreSQL interface
│   │   │   ├── client.go          # Connection management
│   │   │   └── postgres.go        # Query operations
│   │   └── redis/                  # Redis interface
│   │       ├── client.go          # Connection management
│   │       └── redis.go           # Cache operations
│   │
│   ├── http/                       # HTTP infrastructure
│   │   ├── handler/                # Common handlers
│   │   │   └── response.go       # Response utilities
│   │   └── middleware/            # HTTP middleware
│   │       ├── chain.go           # Middleware chaining
│   │       ├── cors.go            # CORS handling
│   │       ├── logging.go         # Request logging
│   │       ├── metrics.go         # Metrics collection
│   │       ├── recovery.go        # Panic recovery
│   │       ├── response_time.go   # Response timing
│   │       ├── tracing.go         # Distributed tracing
│   │       └── auth.go            # Authentication
│   │
│   ├── telemetry/                 # Observability stack
│   │   ├── logger/                # Structured logging
│   │   │   ├── logger.go         # Logger implementation
│   │   │   └── middleware.go     # Logging middleware
│   │   ├── metrics/               # Metrics collection
│   │   │   ├── metrics.go        # OTEL metrics setup
│   │   │   └── middleware.go     # Metrics middleware
│   │   └── span/                  # Distributed tracing
│   │       ├── span.go           # OTEL tracing setup
│   │       └── middleware.go     # Tracing middleware
│   │
│   ├── instance/                  # Infrastructure instances
│   │   ├── local/                # Local development
│   │   │   └── docker-compose.yaml
│   │   └── prod/                 # Production configs
│   │
│   └── deployment/                # Deployment configs
│       ├── collector/            # OpenTelemetry collector
│       └── deployments/          # Kubernetes manifests
│
├── makefile                       # Build & deployment automation
├── go.mod                         # Go module definition
└── go.sum                         # Dependency checksums
```

### Domain Service Structure Pattern

```mermaid
graph LR
    subgraph "Domain/<service-name>"
        DTO[DTO/<br/>Data Transfer Objects]
        Handler[Handler/<br/>HTTP Layer]
        Service[Service/<br/>Business Logic Layer]
        Repo[Repository/<br/>Data Access Layer]
        Cache[Cache.go<br/>Redis Implementation]
        Persistent[Persistent.go<br/>PostgreSQL Implementation]
    end

    Handler --> DTO
    Handler --> Service
    Service --> Repo
    Repo --> Cache
    Repo --> Persistent

    classDef layer fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef impl fill:#c8e6c9,stroke:#33691e,stroke-width:2px

    class DTO,Handler,Service,Repo layer
    class Cache,Persistent impl
```

Each domain follows this consistent structure with:
- **DTO**: Request/response structs
- **Handler**: HTTP handlers
- **Service**: Service interface + implementation
- **Repository**: Repository interface
- **Cache.go**: Redis implementation
- **Persistent.go**: PostgreSQL implementation

---

## Technology Stack

### Core Technologies

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Language** | Go | 1.25 | Primary language |
| **HTTP Router** | Gorilla Mux | Latest | HTTP routing |
| **Database ORM** | SQLx | Latest | PostgreSQL queries |
| **Cache** | Redis (go-redis) | Latest | Caching layer |
| **Authentication** | JWT (golang-jwt) | v5 | Token management |
| **OAuth** | OAuth2 | Latest | Google authentication |
| **API Docs** | go-swagger | Latest | Swagger documentation |

### Observability Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Tracing** | OpenTelemetry + Jaeger | Distributed tracing |
| **Metrics** | OpenTelemetry + Prometheus | Metrics collection |
| **Logging** | OpenTelemetry + Loki | Log aggregation |
| **Visualization** | Grafana | Metrics dashboards |
| **Service Mesh** | Kiali + Istio | Service visualization |

### Infrastructure

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Database** | PostgreSQL 15+ | Primary data store |
| **Cache** | Redis 7+ | Caching layer |
| **NoSQL** | DynamoDB | URL cache (AWS) |
| **Message Queue** | SQS/SNS | Async messaging |
| **Container** | Docker | Containerization |
| **Orchestration** | Kubernetes (Kind/EKS) | Container orchestration |
| **Ingress** | Istio | Service mesh & routing |

### Development Tools

| Tool | Purpose |
|------|---------|
| **Make** | Build automation |
| **Docker Compose** | Local development environment |
| **Kind** | Local Kubernetes cluster |
| **CloudFormation** | AWS infrastructure as code |
| **kubectl** | Kubernetes management |

---

## Observability

The backend-app implements a comprehensive observability stack using **OpenTelemetry** as the unified framework for logs, metrics, and traces.

### Observability Architecture

```mermaid
graph TB
    subgraph "Application Layer"
        D1[Domain 1]
        D2[Domain 2]
        D3[Domain 3]
        DN[Domain N]
    end

    subgraph "OpenTelemetry SDK"
        LP[Log Processor<br/>Batch]
        MP[Metric Processor<br/>Periodic]
        SP[Span Processor<br/>Batch]
    end

    subgraph "OpenTelemetry Collector"
        RX[Receiver]
        PROC[Processor]
        BATCH[Batcher]
        EXP[Exporter]
    end

    subgraph "Backend Storage"
        JGR[Jaeger<br/>Trace Storage<br/>Port: 16686]
        PROM[Prometheus<br/>Metrics<br/>Port: 9090]
        LOKI[Loki<br/>Log Storage<br/>Port: 3100]
    end

    subgraph "Visualization"
        GRF[Grafana Dashboard<br/>Unified View<br/>Port: 3000]
    end

    D1 -->|Logs| LP
    D1 -->|Metrics| MP
    D1 -->|Traces| SP

    D2 -->|Logs| LP
    D2 -->|Metrics| MP
    D2 -->|Traces| SP

    D3 -->|Logs| LP
    D3 -->|Metrics| MP
    D3 -->|Traces| SP

    DN -->|Logs| LP
    DN -->|Metrics| MP
    DN -->|Traces| SP

    LP --> RX
    MP --> RX
    SP --> RX

    RX --> PROC
    PROC --> BATCH
    BATCH --> EXP

    EXP --> JGR
    EXP --> PROM
    EXP --> LOKI

    JGR --> GRF
    PROM --> GRF
    LOKI --> GRF

    classDef app fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef sdk fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef collector fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef storage fill:#c8e6c9,stroke:#33691e,stroke-width:2px
    classDef viz fill:#fce4ec,stroke:#880e4f,stroke-width:2px

    class D1,D2,D3,DN app
    class LP,MP,SP sdk
    class RX,PROC,BATCH,EXP collector
    class JGR,PROM,LOKI storage
    class GRF viz
```

### 1. Logging

#### Implementation Location
`infrastructure/telemetry/logger/logger.go`

#### Features
- **Structured Logging**: Key-value pairs for machine parsing
- **Multiple Formats**: TEXT (human-readable) and JSON (machine-readable)
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Automatic Metadata**: Timestamp, file/line number, caller info
- **OTLP Export**: Sends logs to OpenTelemetry Collector

#### Log Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| **DEBUG** | Detailed debugging info | Development troubleshooting |
| **INFO** | General informational messages | Normal operation tracking |
| **WARN** | Warning messages | Potential issues that don't stop execution |
| **ERROR** | Error messages | Failures that affect functionality |

#### Usage Example

```go
import "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"

func HandleRequest(r *http.Request) {
    logger.Info("request started", map[string]any{
        "method": r.Method,
        "path":   r.URL.Path,
        "query":  r.URL.RawQuery,
    })

    // Business logic...

    logger.Error("operation failed", map[string]any{
        "error": err.Error(),
        "context": "shorten_url",
    })
}
```

#### Log Format

**TEXT Format:**
```
2026-02-14T10:30:45Z INFO [url-shortener] request started method=POST path=/url/shorten
```

**JSON Format:**
```json
{
  "timestamp": "2026-02-14T10:30:45Z",
  "level": "INFO",
  "service": "url-shortener",
  "message": "request started",
  "method": "POST",
  "path": "/url/shorten"
}
```

### 2. Metrics

#### Implementation Location
`infrastructure/telemetry/metrics/metrics.go`

#### Pre-defined Instruments

| Metric Name | Type | Description | Labels |
|-------------|------|-------------|--------|
| `http_requests_total` | Counter | Total HTTP requests | method, path, status_code |
| `http_request_duration_seconds` | Histogram | Request duration in seconds | method, path |
| `http_response_time_latest_seconds` | ObservableGauge | Latest response time | method, path |

#### Metrics Collection Flow

```mermaid
sequenceDiagram
    participant Client
    participant MW as Metrics Middleware
    participant Handler
    participant Prometheus

    Client->>MW: HTTP Request
    activate MW
    MW->>MW: Record start time
    MW->>MW: Wrap ResponseWriter

    MW->>Handler: Process request
    activate Handler
    Handler-->>MW: Response
    deactivate Handler

    MW->>MW: Calculate duration
    MW->>MW: Increment request counter
    MW->>MW: Record duration histogram
    MW->>MW: Update response time gauge

    MW-->>Client: HTTP Response
    deactivate MW

    Note over Prometheus: Every 15 seconds
    Prometheus->>MW: GET /metrics
    activate MW
    MW-->>Prometheus: Metrics data
    deactivate MW
```

#### Example Metrics Output

```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="POST",path="/url/shorten",status_code="200"} 1523

# HELP http_request_duration_seconds Request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="POST",path="/url/shorten",le="0.005"} 100
http_request_duration_seconds_bucket{method="POST",path="/url/shorten",le="0.01"} 450
http_request_duration_seconds_bucket{method="POST",path="/url/shorten",le="0.025"} 800
http_request_duration_seconds_bucket{method="POST",path="/url/shorten",le="+Inf"} 1523
http_request_duration_seconds_sum{method="POST",path="/url/shorten"} 12.5
http_request_duration_seconds_count{method="POST",path="/url/shorten"} 1523
```

### 3. Distributed Tracing (Spans)

#### Implementation Location
`infrastructure/telemetry/span/span.go`

#### Tracing Features
- **OpenTelemetry SDK** for distributed tracing
- **OTLP gRPC exporter** for sending traces
- **Configurable sampling rate** (default 100%)
- **Batch processor** for efficient trace collection
- **W3C Trace Context** for context propagation

#### Span Architecture

```mermaid
graph LR
    subgraph "Trace: Unique Identifier"
        S1[Span 1<br/>HTTP Handler]
        S2[Span 2<br/>Service Logic]
        S3[Span 3<br/>Database Query]
        S4[Span 4<br/>Cache Lookup]
    end

    S1 ==> S2
    S2 ==> S3
    S3 ==> S4

    S1 -.Parent-Child.-> S2
    S2 -.Parent-Child.-> S3
    S3 -.Parent-Child.-> S4

    classDef span fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    class S1,S2,S3,S4 span
```

#### Span Attributes

Each span captures:
- **Service Name**: Which service generated the span
- **Operation Name**: What operation was performed
- **Start/End Time**: When the operation occurred
- **Duration**: How long the operation took
- **Attributes**: Key-value pairs (HTTP method, URL, status code)
- **Events**: Timestamped events within the span
- **Links**: Relationships to other spans

#### Trace Context Propagation

```go
// Incoming request carries trace context in headers
TraceParent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01

// Handler extracts and uses existing trace context
// or creates new one if not present

// Service calls pass context to maintain trace
ctx := r.Context()
service.DoSomething(ctx)

// Database and cache calls inherit trace context
repo.Query(ctx)
cache.Get(ctx)
```

#### Viewing Traces in Jaeger

```
Jaeger UI: http://localhost:16686

Search by:
- Service Name (url-shortener, user, friend, etc.)
- Operation Name (HTTP method + path)
- Tags (status_code, error)
- Time range

Trace view shows:
- Overall timeline
- Individual spans with duration
- Service-to-service calls
- Database queries
- Cache operations
```

### Environment Configuration

All observability components are configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_COLLECTOR_ENDPOINT` | OTLP collector address | `localhost:4317` |
| `SERVICE_NAME` | Service identifier | `backend-app` |
| `ENVIRONMENT` | Deployment environment | `local` |
| `LOG_LEVEL` | Log verbosity | `INFO` |
| `LOG_FORMAT` | Output format | `TEXT` |
| `OTEL_METRICS_INTERVAL` | Metrics push interval (ms) | `15000` |

### Observability in Action: Request Lifecycle

```mermaid
sequenceDiagram
    participant Client
    participant MW as Middleware
    participant Handler
    participant Service
    participant Cache as Redis Cache
    participant DB as PostgreSQL
    participant OTEL as OpenTelemetry

    Client->>MW: HTTP Request
    activate MW
    MW->>OTEL: Create root span
    MW->>MW: Log "request started"
    MW->>MW: Start metrics timer

    MW->>Handler: Process request
    activate Handler
    Handler->>OTEL: Create child span
    Handler->>MW: Log "processing request"

    Handler->>Service: Business logic
    activate Service
    Service->>OTEL: Create child span
    Service->>MW: Log "validating input"

    Service->>Cache: Check cache
    activate Cache
    Cache-->>Service: Cache miss
    deactivate Cache

    Service->>DB: Query database
    activate DB
    DB-->>Service: Result
    deactivate DB

    Service->>MW: Log "operation complete"
    Service-->>Handler: Result
    deactivate Service

    Handler-->>MW: Result
    deactivate Handler

    MW->>OTEL: Record metrics
    MW->>MW: Log "request completed"
    MW->>OTEL: End span with attributes
    MW-->>Client: HTTP Response
    deactivate MW
```

---

## Infrastructure & Domain Connections

### Infrastructure Architecture Diagram

```mermaid
graph TB
    subgraph "External Clients"
        BR[Browser]
        MA[Mobile App]
        TP[3rd Party]
        DV[Developer]
    end

    subgraph "DNS & Ingress"
        DNS[Route53 DNS<br/>syamsul.online<br/>url.syamsul.online<br/>user.syamsul.online]
        INGRESS[Istio Ingress Gateway<br/>TLS Termination]
    end

    subgraph "Kubernetes Cluster"
        subgraph "Services"
            FE[Frontend Services]
            BE[Backend Services]
            OBS[Observability Stack]
        end

        subgraph "Default Namespace"
            URL[url-shortener]
            USR[user]
            FR[friend]
            GRPH[graph]
            MSG[message]
            WS[websocket]
            HC[healthcheck]
        end

        subgraph "Obs Namespace"
            JGR[Jaeger]
            PROM[Prometheus]
            GRF[Grafana]
            LOKI[Loki]
            KIALI[Kiali]
        end
    end

    subgraph "Data Layer"
        PG[(PostgreSQL<br/>RDS/ElastiCache)]
        RDS[(Redis<br/>ElastiCache)]
        DYN[(DynamoDB<br/>URL Cache)]
        SQS[SQS/SNS<br/>Messaging]
    end

    BR --> DNS
    MA --> DNS
    TP --> DNS
    DV --> DNS

    DNS --> INGRESS

    INGRESS --> FE
    INGRESS --> BE
    INGRESS --> OBS

    BE --> URL
    BE --> USR
    BE --> FR
    BE --> GRPH
    BE --> MSG
    BE --> WS
    BE --> HC

    OBS --> JGR
    OBS --> PROM
    OBS --> GRF
    OBS --> LOKI
    OBS --> KIALI

    URL --> PG
    USR --> PG
    FR --> PG
    GRPH --> PG
    MSG --> PG
    WS --> PG

    URL --> RDS
    USR --> RDS
    FR --> RDS
    GRPH --> RDS
    MSG --> RDS
    WS --> RDS

    URL --> DYN
    MSG --> SQS

    classDef client fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef ingress fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef service fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef obs fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef data fill:#c8e6c9,stroke:#33691e,stroke-width:2px

    class BR,MA,TP,DV client
    class DNS,INGRESS ingress
    class FE,BE,URL,USR,FR,GRPH,MSG,WS,HC service
    class OBS,JGR,PROM,GRF,LOKI,KIALI obs
    class PG,RDS,DYN,SQS data
```

### Domain Communication Map

```mermaid
graph TB
    subgraph "Services"
        USVC[User Service]
        FSVC[Friend Service]
        MSVC[Message Service]
        SSVC[URL Shortener]
    end

    subgraph "Communication Types"
        IM[In-Memory<br/>same process]
        HREST[HTTP/REST<br/>external]
        WS[WebSocket<br/>real-time]
    end

    subgraph "Data Layer"
        PGDB[(PostgreSQL/Databases)]
    end

    HTTP[HTTP Requests]

    HTTP --> USVC
    HTTP --> FSVC
    HTTP --> MSVC
    HTTP --> SSVC

    USVC -.Auth verification.-> FSVC
    USVC -.User lookup.-> MSVC
    USVC -.User notifications.-> SSVC

    USVC --> PGDB
    FSVC --> PGDB
    MSVC --> PGDB
    SSVC --> PGDB

    USVC --> IM
    FSVC --> IM
    MSVC --> IM
    SSVC --> IM

    classDef service fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef comm fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef data fill:#c8e6c9,stroke:#33691e,stroke-width:2px
    classDef entry fill:#f3e5f5,stroke:#4a148c,stroke-width:2px

    class USVC,FSVC,MSVC,SSVC service
    class IM,HREST,WS comm
    class PGDB data
    class HTTP entry
```

### External Service Dependencies

```mermaid
graph LR
    subgraph "Internal Services"
        USVC[User Service]
        MSVC[Message Service]
        SSVC[URL Shortener]
    end

    subgraph "External Services"
        GOOG[Google OAuth]
        AWS_SNS[SQS/SNS]
        DYNDB[DynamoDB]
    end

    USVC <-->|OAuth2| GOOG
    MSVC <-->|SNS/SQS<br/>Pub/Sub| AWS_SNS
    SSVC <-->|GetItem/SetItem| DYNDB

    classDef internal fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef external fill:#c8e6c9,stroke:#33691e,stroke-width:2px

    class USVC,MSVC,SSVC internal
    class GOOG,AWS_SNS,DYNDB external
```

### Infrastructure Components

#### VPC & Networking (AWS)

```mermaid
graph TB
    subgraph "VPC 10.0.0.0/16"
        subgraph "Public Subnets<br/>10.0.1.0/24, 10.0.2.0/24"
            ISTIO[Istio Ingress]
            BASTION[Bastion Host]
            NAT[NAT Gateway]
        end

        subgraph "Private Subnets<br/>10.0.10.0/24, 10.0.11.0/24, 10.0.12.0/24"
            EKS[EKS Nodes]
            RDS[(RDS PostgreSQL)]
            ECACHE[ElastiCache Redis)]
        end
    end

    classDef public fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef private fill:#e1f5fe,stroke:#01579b,stroke-width:2px

    class ISTIO,BASTION,NAT public
    class EKS,RDS,ECACHE private
```

#### Data Flow Diagram

```mermaid
sequenceDiagram
    participant Client
    participant DNS as Route53
    participant Ingress as Istio Ingress
    participant Mesh as Service Mesh
    participant K8s as Kubernetes Service
    participant App as Application Pod
    participant Cache as Redis Cache
    participant DB as PostgreSQL

    Client->>DNS: 1. DNS Resolution
    DNS-->>Client: IP Address

    Client->>Ingress: 2. HTTPS Request
    activate Ingress
    Ingress->>Ingress: TLS Termination
    Ingress->>Ingress: Authentication
    Ingress->>Ingress: Routing Decision
    Ingress-->>Mesh: 3. Forward to Service Mesh
    deactivate Ingress

    activate Mesh
    Mesh->>Mesh: mTLS between services
    Mesh->>Mesh: Traffic management
    Mesh->>Mesh: Telemetry collection
    Mesh-->>K8s: 4. Forward to K8s Service
    deactivate Mesh

    activate K8s
    K8s-->>App: 5. Forward to Pod
    deactivate K8s

    activate App
    App->>App: Middleware Chain
    App->>App: Business Logic
    App->>Cache: 6a. Check Cache
    activate Cache
    Cache-->>App: Cache Miss
    deactivate Cache
    App->>DB: 6b. Query Database
    activate DB
    DB-->>App: Result
    deactivate DB
    App-->>Client: 7. Response
    deactivate App
```

---

## Local Development Guide

This section walks you through setting up and running the backend-app locally.

### Prerequisites

Ensure you have the following installed:

| Tool | Version | Check Command |
|------|---------|---------------|
| Go | 1.25+ | `go version` |
| Docker | 20+ | `docker --version` |
| Docker Compose | 2+ | `docker-compose --version` |
| Make | Any | `make --version` |
| kubectl | 1.28+ | `kubectl version` |
| Kind | 0.20+ | `kind version` |

### Quick Start

The fastest way to get started:

```bash
# 1. Navigate to backend-app directory
cd /Users/m.syamsularifin/go/portofolio/backend-app

# 2. Start infrastructure dependencies
docker-compose -f infrastructure/instance/local/docker-compose.yaml up -d

# 3. Set environment variables
export $(cat .env | xargs)

# 4. Run the application
make local
```

The application will be available at `http://localhost:10000`

### Step-by-Step Walkthrough

#### Step 1: Start Infrastructure Services

The local infrastructure runs in Docker Compose and includes:

| Service | Port | Description |
|---------|------|-------------|
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Cache layer |
| OTel Collector | 4317 | Telemetry collector |
| Jaeger | 16686 | Trace visualization (UI) |
| Prometheus | 9090 | Metrics collection |
| Grafana | 3000 | Metrics dashboard |
| Loki | 3100 | Log aggregation |

```bash
# Start all infrastructure services
docker-compose -f infrastructure/instance/local/docker-compose.yaml up -d

# Verify services are running
docker-compose -f infrastructure/instance/local/docker-compose.yaml ps

# View logs
docker-compose -f infrastructure/instance/local/docker-compose.yaml logs -f
```

#### Step 2: Configure Environment Variables

Each service uses environment variables for configuration. Create a `.env` file in the root directory:

```bash
# Copy example env (if available)
cp .env.example .env

# Or create manually
cat > .env << 'EOF'
# Service Configuration
SERVICE_NAME=backend-app
ENVIRONMENT=local
HTTP_PORT=10000

# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=portfolio

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Observability
OTEL_COLLECTOR_ENDPOINT=localhost:4317
LOG_LEVEL=INFO
LOG_FORMAT=TEXT

# Authentication
JWT_TOKEN_TTL=24h
APP_TOKEN_SECRET=your-secret-key-here

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:10000/user/callback

# AWS (for SQS/SNS/DynamoDB - optional for local)
AWS_REGION=us-east-1
EOF
```

#### Step 3: Initialize Database

```bash
# Connect to PostgreSQL
docker exec -it backend-app-postgres-1 psql -U postgres -d portfolio

# Create tables (if not using auto-migration)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Add more tables as needed...
```

#### Step 4: Run the Application

```bash
# Using make
make local

# Or directly with go
go run binary/main.go

# Or build and run
go build -o bin/backend binary/main.go
./bin/backend
```

#### Step 5: Verify the Application

```bash
# Health check
curl http://localhost:10000/health

# View swagger docs
open http://localhost:10000/docs

# View Prometheus metrics
open http://localhost:10000/metrics

# View Jaeger traces
open http://localhost:16686

# View Grafana dashboards
open http://localhost:3000
```

### Local Services Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/url/shorten` | POST | Create shortened URL |
| `/url/{code}` | GET | Resolve shortened URL |
| `/user/login` | POST | User authentication |
| `/friend/add` | POST | Add friend |
| `/friend/list` | GET | List friends |
| `/message/send` | POST | Send message |
| `/graph/viz` | POST | Graph visualization |
| `/metrics` | GET | Prometheus metrics |
| `/docs` | GET | Swagger documentation |

### Debugging Tips

#### Check Application Logs

```bash
# Application logs are stdout/stderr
# View via docker-compose if running in container
docker-compose -f infrastructure/instance/local/docker-compose.yaml logs -f backend

# Or use Loki/Grafana for centralized logging
open http://localhost:3000
```

#### Trace a Request

```bash
# Make a request with trace ID
curl -H "Traceparent: 00-$(uuidgen)-$(uuidgen)-01" http://localhost:10000/health

# View in Jaeger
open http://localhost:16686
```

#### Connect to Database

```bash
# PostgreSQL
docker exec -it backend-app-postgres-1 psql -U postgres -d portfolio

# Redis CLI
docker exec -it backend-app-redis-1 redis-cli
```

### Common Development Tasks

#### Hot Reload during Development

```bash
# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << 'EOF'
root = "."
tmp_dir = "tmp"
[build]
cmd = "go build -o ./tmp/main ./binary/main.go"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
delay = 1000
EOF

# Run with hot reload
air
```

#### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific domain
go test ./domain/url-shortener/...

# Run tests with verbose output
go test -v ./...
```

#### Build Docker Image

```bash
# Build image
docker build -t backend-app:local .

# Run container
docker run -p 10000:10000 --env-file .env backend-app:local
```

### Troubleshooting

#### Port Already in Use

```bash
# Find process using port 10000
lsof -i :10000

# Kill the process
kill -9 <PID>
```

#### Database Connection Failed

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check database logs
docker logs backend-app-postgres-1

# Verify connection string
echo $POSTGRES_HOST:$POSTGRES_PORT
```

#### Redis Connection Failed

```bash
# Check Redis is running
docker ps | grep redis

# Test Redis connection
docker exec -it backend-app-redis-1 redis-cli ping
```

#### Metrics Not Showing

```bash
# Check OTLP collector is running
docker ps | grep otel

# Verify collector endpoint
echo $OTEL_COLLECTOR_ENDPOINT

# Check Prometheus scraping
open http://localhost:9090/targets
```

### Cleaning Up

```bash
# Stop all services
docker-compose -f infrastructure/instance/local/docker-compose.yaml down

# Remove volumes (deletes data)
docker-compose -f infrastructure/instance/local/docker-compose.yaml down -v

# Remove built artifacts
rm -rf bin/ tmp/
```

---

## Summary

This backend-app demonstrates a well-structured, production-ready Go application with:

- **Clean Architecture**: Layered design with clear separation of concerns
- **Domain-Driven Design**: Business logic organized by domain
- **Full Observability**: Comprehensive logs, metrics, and traces
- **Modern Infrastructure**: Kubernetes deployment with Istio service mesh
- **Developer Experience**: Hot reload, comprehensive testing, and clear documentation

For questions or contributions, please refer to the project repository.

---

**Document Version**: 1.0
**Last Updated**: 2026-02-14
