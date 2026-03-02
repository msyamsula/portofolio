# Copilot Instructions for AI Agents

## Project Overview
- **Monorepo** with Go backend (`backend-app`), Node.js/JS frontend (`frontend-app`), and learning/experiments (`learning/`).
- Backend follows **Clean Architecture** and **Domain-Driven Design**. See `backend-app/docs/clean-architecture.md` and `backend-app/docs/domain-driven-design.md`.
- Infrastructure and observability are first-class: OpenTelemetry, Prometheus, Jaeger, Loki, Grafana, PostgreSQL, Redis.

## Key Architectural Patterns
- **Layered backend**: Entry (HTTP/gRPC) → Handler → Service → Repository → Infrastructure. All dependencies point inward.
- **Domain structure**: Each domain (e.g., `user`, `friend`, `message`) has `handler/`, `service/`, `repository/`, `dto/`, `integration/`.
- **Repository pattern**: Data access is abstracted; see `backend-app/docs/repository-pattern.md`.
- **Observability**: All services instrumented for metrics, logs, and traces. See `backend-app/infrastructure/telemetry/README.md`.
- **Deployment**: Local via Docker Compose (`make infra-start`), production via Kubernetes manifests.

## Developer Workflows
- **Start local infra**: `make infra-start` (runs Docker Compose for DB, cache, telemetry, etc.)
- **Stop local infra**: `make infra-stop`
- **Run backend HTTP server**: `make -C backend-app/binary/http up`
- **Swagger docs**: `make -C backend-app/binary/http swagger`
- **Frontend**: See `frontend-app/*/makefile` for per-app commands.
- **Testing**: (Add/expand here if test conventions are found)

## Conventions & Practices
- **Environment config**: Use `.env` files and environment variables. See `backend-app/infrastructure/telemetry/config.md` for telemetry vars.
- **Documentation**: All major design docs in `backend-app/docs/` and `backend-app.md`. Use Obsidian for knowledge graph (`backend-app/readme.md`).
- **Observability**: Always instrument new services for metrics/logs/traces. Use provided telemetry packages.
- **Kubernetes**: Manifests in `backend-app/infrastructure/deploymnet/deployments/`.
- **Makefile targets**: Root Makefile proxies to sub-makefiles for infra and HTTP server.

## Integration Points
- **External services**: Google OAuth, AWS SQS/SNS, DynamoDB (see `backend-app/domain/user/README.md` and infra docs).
- **Telemetry**: All signals exported to OTel Collector, visualized in Grafana/Jaeger.

## Examples
- **Add a new domain**: Copy structure from `backend-app/domain/user/`.
- **Instrument a handler**: Use telemetry client from `infrastructure/telemetry/`.
- **Add infra service**: Update Docker Compose in `backend-app/infrastructure/instance/local/` and relevant Makefile.

## References
- `backend-app/docs/architecture-overview.md` — High-level design
- `backend-app/docs/code-structure.md` — Directory layout & domain pattern
- `backend-app/docs/clean-architecture.md` — Layered architecture
- `backend-app/infrastructure/telemetry/README.md` — Observability stack
- `backend-app/infrastructure/deploymnet/README.md` — Deployment infrastructure (note: directory is spelled `deploymnet` in the repo)

---
If any section is unclear or missing, please provide feedback for further refinement.
