# Observability Data Flow Documentation

## Architecture Overview

The observability stack follows a centralized collector pattern where all telemetry data flows through the OpenTelemetry Collector before reaching backend storage systems.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Services (url-shortener, user, etc.)                │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐                       │
│  │ HTTP Layer  │  │ Business     │  │ Data Layer   │                       │
│  │ otelhttp    │  │ Manual       │  │ DB tracing   │                       │
│  └──────┬──────┘  └──────┬───────┘  └──────┬───────┘                       │
│         │                │                  │                                 │
│         └────────────────┴──────────────────┘                                 │
│                            │                                                 │
│                    OTLP Exporter                                             │
│                            │                                                 │
│                    TRACER_COLLECTOR_ENDPOINT                                 │
└────────────────────────────┼─────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   OpenTelemetry Collector (config.yaml)                      │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                        Receivers                                    │     │
│  │  ┌─────────────────┐  ┌─────────────────┐                          │     │
│  │  │ OTLP (gRPC)     │  │ OTLP (HTTP)     │  ← port 4317, 4318     │     │
│  │  │ :4317           │  │ :4318           │                          │     │
│  │  └─────────────────┘  └─────────────────┘                          │     │
│  │                                                                   │     │
│  │  ┌─────────────────┐                                               │     │
│  │  │ Prometheus      │  ← scrapes :8889 (collector self-metrics)   │     │
│  │  │ Scraper         │                                               │     │
│  │  └─────────────────┘                                               │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                              │                                              │
│                              ▼                                              │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                        Processors                                   │     │
│  │  ┌─────────────────┐  ┌─────────────────┐                          │     │
│  │  │ Memory Limiter  │  │ Batch (1s,      │                          │     │
│  │  │ 75% limit       │→ │ 1024 size)      │                          │     │
│  │  └─────────────────┘  └─────────────────┘                          │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                              │                                              │
│                              ▼                                              │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │                        Exporters                                    │     │
│  │  ┌─────────────────┐  ┌─────────────────┐                          │     │
│  │  │ Jaeger (OTLP)   │  │ Prometheus      │                          │     │
│  │  │ jaeger:4317     │  │ Remote Write    │                          │     │
│  │  └─────────────────┘  │ :9090/api/v1/   │                          │     │
│  │                       │ write            │                          │     │
│  │                       └─────────────────┘                          │     │
│  │                                                                   │     │
│  │  ┌─────────────────┐                                               │     │
│  │  │ Prometheus      │  ← exposes :8889 for scraping                 │     │
│  │  │ Exporter        │                                               │     │
│  │  └─────────────────┘                                               │     │
│  └────────────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────────────┘
                             │                   │
                             ▼                   ▼
┌────────────────────────┐  ┌─────────────────────────────────────────────┐
│      Jaeger            │  │            Prometheus                       │
│  (obs namespace)       │  │        (obs namespace)                     │
│                        │  │                                             │
│  • Query UI :16686     │  │  • Scrape interval: 15s                    │
│  • OTLP :4317          │  │  • Retention: 15d                          │
│  • Storage: OpenSearch │  │  • Evaluates: every 1m                     │
└────────────────────────┘  └─────────────────────────────────────────────┘
                                                                      │
                                                                      ▼
                                                          ┌───────────────────┐
                                                          │     Grafana       │
                                                          │   (obs namespace) │
                                                          │                   │
                                                          │  • Dashboard UI   │
                                                          │  • Port :3000     │
                                                          └───────────────────┘
```

## Data Flow: Traces (Jaeger)

### 1. Service-Side Generation

Location: `pkg/telemetry/telemetry.go`

```go
func InitializeTelemetryTracing(serviceName, collectorEndpoint string) func()
```

Each service generates spans through:
- **Automatic instrumentation**: `otelhttp.NewHandler()` wraps HTTP handlers
- **Manual instrumentation**: Custom spans for business logic

**Trace attributes included:**
- `service.name`: Application name
- `service.version`: Application version
- `deployment.environment`: Deployment environment

### 2. Export to Collector
```
Service → OTLP gRPC → Collector (otlp receiver :4317)
```

Configuration via environment variable:
```bash
TRACER_COLLECTOR_ENDPOINT=collector:4317
```

### 3. Collector Processing

From `config.yaml`:
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  memory_limiter:
    check_interval: 1s
    limit_percentage: 75
    spike_limit_percentage: 30
  batch:
    timeout: 1s
    send_batch_size: 1024

exporters:
  jaeger:
    endpoint: jaeger:4317
    tls:
      insecure: true
```

### 4. Jaeger Storage
```
Collector → Jaeger Collector (jaeger:4317) → OpenSearch → Query UI (:16686)
```

## Data Flow: Metrics (Prometheus)

### 1. Service-Side Collection

Each service exposes metrics at `/metrics` endpoint via `prometheus/client_golang`.

### 2. Collection Methods

**Method A: Direct Scraping (Traditional)**
```
Prometheus → scrape → /metrics endpoint
```

**Method B: OTLP → Collector (Remote Write)**
```
Service → OTLP → Collector → Prometheus Remote Write API
```

### 3. Collector Metrics Pipeline

From `config.yaml`:
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318
  prometheus:
    config:
      scrape_configs:
        - job_name: 'otel-collector'
          static_configs:
            - targets: ['localhost:8889']

exporters:
  prometheusremotewrite:
    endpoint: http://prometheus:9090/api/v1/write
    tls:
      insecure: true
    send_queue_size: 1000
    send_queue_shard_count: 10
  prometheus:
    endpoint: "0.0.0.0:8889"
    const_labels:
      cluster: "kubernetes"
      environment: "production"
    namespace: "otelcol"
```

## Deployment Configuration

### Namespace
All observability components deployed in `obs` namespace:
```bash
kubectl create namespace obs
```

### Service Endpoints

| Component | Internal | External |
|-----------|----------|----------|
| Collector | `collector:4317` (OTLP gRPC) | Port-forward for debugging |
| Jaeger UI | `jaeger-query:16686` | `kubectl port-forward -n obs svc/jaeger 16686:16686` |
| Prometheus | `prometheus:9090` | `kubectl port-forward -n obs svc/prometheus 9090:9090` |
| Grafana | `grafana:3000` | `kubectl port-forward -n obs svc/grafana 3000:3000` |

## Environment Variables Required

For each service:
```bash
TRACER_COLLECTOR_ENDPOINT=collector:4317
```

## Key Configuration Files

| File | Purpose |
|------|---------|
| `infrastructure/telemetry/collector/config.yaml` | Collector pipeline configuration |
| `pkg/telemetry/telemetry.go` | Service-side initialization |
| `observability/jaeger.yaml` | Jaeger deployment |
| `observability/prom.yaml` | Prometheus deployment |
| `observability/grafana.yaml` | Grafana deployment |
