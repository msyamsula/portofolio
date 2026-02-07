# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based portfolio project demonstrating microservices architecture with Kubernetes deployment. It showcases various backend services (URL shortener, graph visualizer, chat application, user management, etc.) deployed on Kubernetes with comprehensive observability.

## Repository Structure

```
portofolio/
├── backend-app/          # Microservices backend applications
│   ├── pkg/             # Shared packages (logger, middleware, telemetry, cache, rate-limiter)
│   ├── url-shortener/   # URL shortening service
│   ├── user/            # User authentication service
│   ├── friend/          # Friend management service
│   ├── graph/           # Graph visualization service
│   ├── message/         # Message handling service
│   ├── websocket-server/# WebSocket real-time communication
│   ├── observability/   # Jaeger, Prometheus, Grafana, Kiali configs
│   └── ingress/         # Ingress and Route53 configurations
├── frontend-app/        # Simple HTML/JavaScript frontends
│   ├── main-page/       # Portfolio main page
│   ├── chat/            # Chat application UI
│   └── code-review/     # Code review UI
└── learning/            # Educational projects (pub-sub, k8s, AWS, face-detector, etc.)
```

## Development Commands

### Running Services Locally

Each service in `backend-app/` has its own makefile with these common targets:

```bash
# Run service locally (with .env file)
cd backend-app/<service>
make local

# Build Docker image
make build

# Deploy to Kubernetes
make kube
```

### Kubernetes Development with Kind

For local Kubernetes development using Kind:

```bash
# Start Kind cluster
kind create cluster --name <cluster-name>

# Load Docker images to Kind
kind load docker-image <image-name> --name <cluster-name>

# Apply Kubernetes deployments
kubectl apply -f deployment.yaml
```

### Building and Deploying

```bash
# Build a specific service
cd backend-app/url-shortener
docker build --platform linux/arm64 -t url-shortener .

# Create/ update ConfigMap from .env
make setup-env

# Deploy to Kubernetes
kubectl apply -f deployment.yaml

# Port forward to test locally
kubectl port-forward svc/<service-name> <local-port>:<service-port>
```

### AWS Cloud Deployment

The `backend-app/makefile` contains CloudFormation commands:

```bash
cd backend-app

# Create ECR repositories
make repo

# Create full tech stack (EKS, VPC, etc.)
make create

# Connect kubectl to EKS
make kube
```

## Architecture Patterns

### Service Structure

Each backend service follows this layered architecture:

```
<service>/
├── handler/          # HTTP handlers (request/response handling)
├── service/          # Business logic layer
├── persistent/       # Data access layer (PostgreSQL, DynamoDB)
├── cache/            # Caching layer (Redis)
├── main.go           # Application entry point
├── makefile          # Build/deploy commands
├── dockerfile        # Container definition
└── deployment.yaml   # Kubernetes manifests
```

### Shared Packages (`backend-app/pkg/`)

- **logger** - Structured logging with logrus
- **middleware** - Common middleware (authentication, CORS)
- **telemetry** - OpenTelemetry tracing initialization
- **cache** - Redis and DynamoDB caching interfaces
- **rate-limiter** - Rate limiting utilities
- **randomizer** - String/number generation utilities

### Dependency Management

- Services import `github.com/msyamsula/portofolio/backend-app/pkg` for shared utilities
- Each service has its own `go.mod` for service-specific dependencies
- The root `go.mod` is for the learning projects only

### Observability Stack

All services are instrumented with:
- **OpenTelemetry** - Distributed tracing with OTLP
- **Jaeger** - Trace visualization (deployed in `obs` namespace)
- **Prometheus** - Metrics collection (`/metrics` endpoint on each service)
- **Grafana** - Metrics dashboard
- **Kiali** - Service mesh visualization (Istio)

### Service Communication

- HTTP/REST between services using Gorilla Mux
- WebSocket for real-time features (chat service)
- Services expose ClusterIP services within Kubernetes
- Ingress controller routes external traffic

## Environment Configuration

Each service uses a `.env` file for configuration. Kubernetes ConfigMaps are created from these files:

```bash
kubectl create configmap <service>-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
```

Key environment variables include:
- Database connection strings
- Redis host/port
- JWT secrets
- OAuth credentials
- Telemetry collector endpoints

## Testing

Run services locally with `make local` which:
1. Sources the `.env` file
2. Runs `go run main.go`

For API testing, use the Postman collection referenced in the README.

## Module Import Notes

**Important**: The `learning/pub-sub` directory contains Google Cloud Pub/Sub examples that import `cloud.google.com/go/pubsub`. This is a standalone learning project and should not be imported by backend services. The main backend services use AWS SQS/SNS for messaging.

## Architecture Diagram

The system architecture is documented at:
https://app.diagrams.net/#G16osGglyMotNDbr098PLH6Mqbj7DCFZUl

## Key Services

### URL Shortener
- Custom hash algorithm with configurable collision probability
- PostgreSQL persistence + Redis caching
- Metrics endpoint for Prometheus
- Canary deployments supported

### User Service
- Google OAuth authentication
- JWT token generation/validation
- Acts as authentication authority for other services

### Graph Visualizer
- Visualizes popular graph algorithms
- Interactive frontend

### WebSocket Server
- Real-time bidirectional communication
- Used by chat application

## Technology Stack

- **Go 1.25** - Primary language
- **Gorilla Mux** - HTTP routing
- **PostgreSQL** - Primary database
- **Redis** - Caching layer
- **DynamoDB** - Alternative persistence (AWS)
- **Kubernetes (Kind)** - Local orchestration
- **AWS EKS** - Production orchestration
- **Istio** - Service mesh (optional)
- **OpenTelemetry + Jaeger** - Distributed tracing
- **Prometheus + Grafana** - Metrics and monitoring
