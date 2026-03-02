# Telemetry Stack

The Telemetry Stack provides comprehensive observability using OpenTelemetry.

## Architecture

```mermaid
flowchart TB
    subgraph App[Backend App]
        Domains[Business Domains]
        Infra[Infrastructure]
    end

    subgraph Telemetry[Telemetry Stack]
        Logger[Structured Logging]
        Metrics[Metrics Collection]
        Tracing[Distributed Tracing]
    end

    subgraph OTel[OTel Collector]
        Collector[Collector]
    end

    subgraph Storage[Storage & Visualization]
        Prometheus[Prometheus]
        Loki[Loki]
        Jaeger[Jaeger]
        Grafana[Grafana]
    end

    Domains --> Logger
    Domains --> Metrics
    Domains --> Tracing
    Infra --> Logger
    Infra --> Metrics
    Infra --> Tracing

    Logger -->|OTLP| Collector
    Metrics -->|OTLP| Collector
    Tracing -->|OTLP| Collector

    Collector --> Prometheus
    Collector --> Loki
    Collector --> Jaeger

    Prometheus --> Grafana
    Loki --> Grafana
    Jaeger --> Grafana
```

## Components

| Component | Location | Purpose |
|-----------|-----------|---------|
| Structured Logging | `logger/` | Log aggregation |
| Metrics Collection | `metrics/` | Metrics collection |
| Distributed Tracing | `span/` | Trace spans |

## Three Pillars

### Logs

- Format: JSON or TEXT
- Levels: DEBUG, INFO, WARN, ERROR
- Export: OTLP to Loki

### Metrics

- Type: Counters, Histograms
- Export: Prometheus
- Record: Request count, latency

### Traces

- Tool: OpenTelemetry
- Export: Jaeger
- Purpose: Track request flow

## Related

- [infrastructure/telemetry/logger/README.md](Structured Logging)
- [infrastructure/telemetry/metrics/README.md](Metrics Collection)
- [infrastructure/telemetry/span/README.md](Distributed Tracing)
- [[docs/architecture-overview.md|Observability]]
