# Backend App

A comprehensive Go backend application featuring URL shortening, graph algorithms, friend management, messaging, and user authentication. Built with a domain-driven modular architecture following clean code principles.

## Quick Start

```bash
# Start local dependencies (DB + observability)
cd infrastructure/instance/local && make start

# Run the backend
cd binary/http && make up
```

## Documentation Vault

This README serves as a **vault index** for navigating through the codebase documentation. Each section below links to detailed documentation located near the implementation.

### High-Level Architecture

| Topic | File |
|--------|------|
| [Architecture Overview](docs/architecture-overview.md) | High-level design |
| [Code Structure](docs/code-structure.md) | Directory layout |
| [Design Principles](docs/design-principles.md) | Core principles |
| [Request Flow](docs/request-flow.md) | Request lifecycle |

### Design Patterns

| Pattern | File | Description |
|----------|------|-------------|
| [Clean Architecture](docs/clean-architecture.md) | Layered architecture |
| [Domain-Driven Design](docs/domain-driven-design.md) | Bounded contexts |
| [Repository Pattern](docs/repository-pattern.md) | Data access abstraction |
| [Cache-Aside Pattern](docs/cache-aside-pattern.md) | Performance optimization |
| [Dependency Inversion](docs/dependency-inversion.md) | Interface abstractions |

### Business Domains

| Domain | File | Description |
|--------|------|-------------|
| [URL Shortener](domain/url-shortener/README.md) | Shorten and expand URLs |
| [Friend](domain/friend/README.md) | Manage friendships |
| [Message](domain/message/README.md) | Messaging system |
| [User](domain/user/README.md) | Authentication |
| [Graph](domain/graph/README.md) | Graph algorithms |
| [Healthcheck](domain/healthcheck/README.md) | Health monitoring |

### Infrastructure

| Component | File | Description |
|-----------|------|-------------|
| [Database Layer](infrastructure/database/README.md) | Database clients |
| [PostgreSQL Client](infrastructure/database/postgres/README.md) | PostgreSQL connectivity |
| [Redis Client](infrastructure/database/redis/README.md) | Redis caching |
| [HTTP Infrastructure](infrastructure/http/README.md) | HTTP server & middleware |
| [HTTP Middleware](infrastructure/http/middleware/README.md) | Cross-cutting concerns |
| [Telemetry Stack](infrastructure/telemetry/README.md) | Observability |
| [Structured Logging](infrastructure/telemetry/logger/README.md) | Log aggregation |
| [Metrics Collection](infrastructure/telemetry/metrics/README.md) | Metrics tracking |
| [Distributed Tracing](infrastructure/telemetry/span/README.md) | Tracing |
| [Deployment Infrastructure](infrastructure/deploymnet/README.md) | Deployment configs |
| [HTTP Server](binary/http/README.md) | HTTP server entry point |
| [Test Mocks](mock/README.md) | Test doubles |

### Observability Tools

| Tool | File | Description |
|------|------|-------------|
| Prometheus | [Metrics Collection](infrastructure/telemetry/metrics/README.md) | Metrics storage |
| Loki | [Structured Logging](infrastructure/telemetry/logger/README.md) | Log aggregation |
| Jaeger | [Distributed Tracing](infrastructure/telemetry/span/README.md) | Trace visualization |
| Grafana | [Telemetry Stack](infrastructure/telemetry/README.md) | Dashboard visualization |

### Authentication

| Component | File | Description |
|-----------|------|-------------|
| OAuth 2.0 | [User Domain](domain/user/README.md) | Google OAuth integration |
| JWT Authentication | [User Domain](domain/user/README.md) | Token-based auth |

### Kubernetes

| Topic | File | Description |
|------|------|-------------|
| Kubernetes Probes | [Healthcheck Domain](domain/healthcheck/README.md) | Liveness/readiness probes |

### Architecture Decision Records

For detailed architectural decisions, see [ADR documentation](docs/adr/001-code-structure.md).

---

## Obsidian Vault

To open this codebase in Obsidian as a knowledge graph:

1. Open Obsidian
2. Click "Open folder as vault"
3. Select `/Users/m.syamsularifin/go/portofolio/backend-app/`

The graph view will show:
- All documentation files as nodes
- Connections between related concepts
- Clear clusters for domains, infrastructure, and patterns

---

## Table of Contents

- [Quick Start](#quick-start)
- [Documentation Vault](#documentation-vault)
  - [High-Level Architecture](#high-level-architecture)
  - [Design Patterns](#design-patterns)
  - [Business Domains](#business-domains)
  - [Infrastructure](#infrastructure)
  - [Observability Tools](#observability-tools)
  - [Authentication](#authentication)
  - [Kubernetes](#kubernetes)
  - [Architecture Decision Records](#architecture-decision-records)
- [Local Development](#local-development)
- [API Endpoints](#api-endpoints)
- [Technology Stack](#technology-stack)

## Local Development

### Services

| Service | Port | Purpose |
|---------|------|---------|
| Backend App | 5000 | HTTP API |
| OTel Collector | 4317, 4318 | Telemetry ingestion |
| Jaeger UI | 16686 | Trace visualization |
| Prometheus | 9090 | Metrics query |
| Grafana | 3000 | Dashboards |
| Loki | 3100 | Log aggregation |
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Caching layer |

### Access Points

- **API**: `http://localhost:5000`
- **Swagger UI**: `http://localhost:5000/swagger/index.html`
- **Jaeger**: `http://localhost:16686`
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (admin/admin)

## API Endpoints

| Domain | Base Path | Methods |
|--------|-----------|---------|
| URL Shortener | `/url` | `POST /shorten`, `GET /{shortCode}` |
| Graph | `/graph` | Algorithm endpoints |
| Friend | `/friend` | CRUD operations |
| Message | `/message` | Send/receive messages |
| User | `/user` | OAuth, token management |
| Healthcheck | `/health` | Liveness, readiness probes |

## Technology Stack

| Category       | Technology         |
| -------------- | ------------------ |
| Language       | Go 1.25+           |
| HTTP Routing   | `gorilla/mux`      |
| Database       | PostgreSQL, Redis  |
| Authentication | Google OAuth, JWT  |
| Observability  | OpenTelemetry      |
| Documentation  | Swagger/OpenAPI    |
| Deployment     | Docker, Kubernetes |

---

**Last Updated**: 2026-03-02
