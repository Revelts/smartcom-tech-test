# Integration Platform - Microservices Monorepo

A production-grade Go microservices platform for processing and routing integration events. Built with clean architecture principles, this system demonstrates how to build independently deployable services within a single monorepo.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Why Microservices in a Monorepo](#why-microservices-in-a-monorepo)
- [Service Boundaries](#service-boundaries)
- [Event Flow](#event-flow)
- [Concurrency Model](#concurrency-model)
- [Retry & Failure Strategy](#retry--failure-strategy)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Running Services](#running-services)
- [Configuration](#configuration)
- [Testing](#testing)
- [Future Improvements](#future-improvements)

## Architecture Overview

This platform consists of two independent microservices:

1. **Middleware Integration Service** - Receives events via HTTP, normalizes them, and forwards to external systems asynchronously
2. **External Alert Endpoint Service** - Simulates an external alerting system that receives processed events

Both services follow Clean Architecture principles:
- **Domain Layer**: Business logic and entities (framework-agnostic)
- **Use Case Layer**: Application-specific business rules
- **Interface Adapters**: HTTP handlers, repositories, infrastructure
- **Infrastructure**: External dependencies (HTTP clients, databases, etc.)

## Why Microservices in a Monorepo

### Benefits of Microservices

- **Independent Deployment**: Each service can be deployed without affecting others
- **Isolated Failures**: A failure in one service doesn't cascade to others
- **Technology Flexibility**: Services can use different tech stacks if needed
- **Team Autonomy**: Different teams can own different services
- **Scalability**: Scale services independently based on load

### Benefits of Monorepo

- **Shared Code**: Common utilities (logger, HTTP client, config) in `/pkg`
- **Atomic Changes**: Update shared dependencies across all services in one commit
- **Code Reuse**: Avoid code duplication across services
- **Simplified Development**: Single checkout, unified tooling
- **Easier Refactoring**: Cross-service refactoring is straightforward
- **Consistent Standards**: Enforce coding standards across all services

### Best of Both Worlds

The monorepo structure provides shared infrastructure while maintaining strict service boundaries:
- Services CANNOT import each other's internal packages
- Shared code ONLY through `/pkg`
- Each service has its own lifecycle, configuration, and build
- Easy to add new microservices without creating new repositories

## Service Boundaries

### Middleware Integration Service

**Responsibility**: Accept external events, normalize them, and route to appropriate destinations

**Owns**:
- Event validation and normalization
- Severity to priority mapping
- Async processing queue
- Worker pool management
- Retry logic for outbound requests

**Exposes**: `POST /integrations/events`

**Internal Components**:
- `domain/`: Event entity and interfaces
- `usecase/`: Business logic (mapping, processing)
- `handler/`: HTTP request handling
- `worker/`: Async worker pool
- `repository/`: Event queue implementation
- `infrastructure/`: ID generation, external dependencies

### External Alert Endpoint Service

**Responsibility**: Simulate an external alerting system

**Owns**:
- Alert reception
- Request logging
- Response generation

**Exposes**: `POST /external/alerts`

**Internal Components**:
- `handler/`: HTTP request handling
- `infrastructure/`: Minimal infrastructure needs

## Event Flow

```
┌─────────┐                                  ┌──────────────────┐
│ Client  │                                  │   Middleware     │
│         │                                  │   Service        │
└────┬────┘                                  │   (Port 8080)    │
     │                                       └────────┬─────────┘
     │ 1. POST /integrations/events               │
     │    (source, event_type, severity,          │
     │     message, metadata)                     │
     │──────────────────────────────────────────>│
     │                                            │
     │ 2. 200 OK (event_id, correlation_id)      │ 3. Generate correlation_id
     │<──────────────────────────────────────────│    Map incoming event
     │                                            │    Enqueue normalized event
     │                                            │
     │                                            │ 4. Worker picks event
     │                                            │    from queue
     │                                            │
     │                                            │ 5. HTTP POST with retry
     │                                            │    (timeout: 3s, retries: 3)
     │                                            │
     │                                            ▼
     │                                 ┌────────────────────┐
     │                                 │   External         │
     │                                 │   Endpoint Service │
     │                                 │   (Port 8081)      │
     │                                 └─────────┬──────────┘
     │                                           │
     │                                           │ 6. Log alert with
     │                                           │    correlation_id
     │                                           │
     │                                           │ 7. 200 OK
     │                                           │
     │                                           ▼
```

### Flow Details

1. **Client Request**: External client sends event to middleware service
2. **Immediate Response**: Handler validates, generates correlation ID, enqueues event, returns HTTP 200 immediately
3. **Async Processing**: Worker pool processes events from bounded channel
4. **External Integration**: HTTP client sends processed event to external endpoint with retry logic
5. **Failure Handling**: Failures are logged but don't block the queue or affect other events

## Concurrency Model

### Middleware Service

**Bounded Channel Queue**:
- Default size: 1000 events
- Prevents unbounded memory growth during traffic bursts
- Backpressure: Returns 503 when queue is full

**Fixed Worker Pool**:
- Default: 10 workers
- Each worker runs in its own goroutine
- Workers share a single queue via channel
- Context-based lifecycle management

**Graceful Shutdown**:
```
1. SIGINT/SIGTERM received
2. HTTP server stops accepting new requests
3. Existing HTTP requests complete (30s timeout)
4. Queue closed (no new events accepted)
5. Workers drain remaining events
6. Workers shutdown (30s timeout)
7. Process exits
```

**Concurrency Control**:
- No shared mutable state between workers
- Each event processed independently
- Worker ID tracked via context
- Correlation ID propagated through entire flow

### Why This Design

- **Predictable Resource Usage**: Fixed workers = predictable CPU/memory
- **No Goroutine Leaks**: Context cancellation ensures clean shutdown
- **Backpressure Handling**: Bounded queue protects against overload
- **Failure Isolation**: One slow/failed event doesn't block others
- **Observable**: Each event tracked via correlation ID

## Retry & Failure Strategy

### HTTP Client Configuration

```
Timeout:       3 seconds
Max Retries:   3 attempts
Base Delay:    500ms
Backoff:       Exponential (500ms, 1s, 2s)
```

### Retry Logic

1. **First Attempt**: Immediate
2. **Second Attempt**: After 500ms delay
3. **Third Attempt**: After 1s delay (cumulative)
4. **Fourth Attempt**: After 2s delay (cumulative)

### Retry Conditions

**Retries on**:
- Network errors
- Timeouts
- HTTP 5xx errors

**No retry on**:
- Context cancellation
- HTTP 4xx errors (client errors)
- After max retries exhausted

### Failure Handling

**Philosophy**: Log failures, never crash

- Failed events are logged with correlation ID
- Worker continues processing other events
- No Dead Letter Queue (see Future Improvements)
- No infinite retries (prevents resource exhaustion)

**Structured Logging**:
```json
{
  "level": "error",
  "msg": "failed to send event",
  "correlation_id": "abc123...",
  "event_id": "def456...",
  "error": "context deadline exceeded",
  "timestamp": "2024-02-10T12:00:00Z"
}
```

## Project Structure

```
.
├── services/                    # Microservices
│   ├── middleware/
│   │   ├── cmd/
│   │   │   └── main.go         # Service entry point
│   │   └── internal/           # Internal packages (not importable)
│   │       ├── domain/         # Business entities & interfaces
│   │       ├── usecase/        # Business logic
│   │       ├── handler/        # HTTP handlers
│   │       ├── worker/         # Worker pool
│   │       ├── repository/     # Event queue
│   │       └── infrastructure/ # ID generation, etc.
│   └── external-endpoint/
│       ├── cmd/
│       │   └── main.go
│       └── internal/
│           ├── handler/
│           └── infrastructure/
│
├── pkg/                        # Shared packages
│   ├── logger/                 # Structured logging
│   ├── correlation/            # Correlation ID utilities
│   ├── httpclient/             # HTTP client with retry
│   ├── config/                 # Environment configuration
│   └── errors/                 # Error types & utilities
│
├── go.work                     # Go workspace
├── Makefile                    # Build automation
├── docker-compose.yml          # Multi-service orchestration
└── README.md
```

### Key Principles

- **internal/**: Prevents accidental imports between services
- **/pkg**: Explicitly shared code only
- **Clean Architecture**: Domain layer has no framework dependencies
- **Dependency Injection**: Constructor injection throughout
- **Interface-based**: Use case layer depends on interfaces, not implementations

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker & Docker Compose (optional, for containerized setup)
- Make (optional, for convenience commands)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd smartcom-tech-test

# Download dependencies
go work sync
cd pkg && go mod tidy
cd ../services/middleware && go mod tidy
cd ../external-endpoint && go mod tidy
```

## Deployment Options

### Google Cloud Platform (GCP)

Deploy to separate VMs in production (optimized for Asia regions).

#### 📖 Choose Your Deployment Method

**Not sure which to use?** → Start with **[DEPLOYMENT_OPTIONS.md](DEPLOYMENT_OPTIONS.md)**

**Manual Deployment** (⭐ Recommended for Production & Learning):
```bash
# Complete step-by-step guide with NO automation scripts
# Full control, easy troubleshooting, production-ready
open GCP_MANUAL_DEPLOYMENT.md
```

**Automated Deployment** (Quick Testing):
```bash
# One command deployment - deployed in 5-10 minutes
# Defaults to Singapore (asia-southeast1-a)
./deploy-gcp.sh
```

#### 📚 Complete GCP Documentation

| Guide | Purpose | Time |
|-------|---------|------|
| **[GCP_DEPLOYMENT_INDEX.md](GCP_DEPLOYMENT_INDEX.md)** | Documentation overview | 5 min |
| **[DEPLOYMENT_OPTIONS.md](DEPLOYMENT_OPTIONS.md)** | Choose deployment method | 10 min |
| **[GCP_MANUAL_DEPLOYMENT.md](GCP_MANUAL_DEPLOYMENT.md)** | Manual step-by-step (no scripts) | 30-45 min |
| **[GCP_QUICKSTART.md](GCP_QUICKSTART.md)** | Automated quick start | 5-10 min |
| **[GCP_ASIA_ZONES.md](GCP_ASIA_ZONES.md)** | Asia region selection | 10 min |
| **[DEPLOYMENT_CHECKLIST.md](DEPLOYMENT_CHECKLIST.md)** | Production checklist | During deployment |

**Default Region**: Singapore (`asia-southeast1-a`) - Optimized for Southeast Asia with excellent connectivity.

## Running Services Locally

### Option 1: Run Services Independently (Development)

**Terminal 1 - External Endpoint**:
```bash
# Run external endpoint service
cd services/external-endpoint
PORT=8081 go run ./cmd/main.go
```

**Terminal 2 - Middleware**:
```bash
# Run middleware service
cd services/middleware
export PORT=8080
export EXTERNAL_ENDPOINT_URL=http://localhost:8081/external/alerts
export QUEUE_SIZE=1000
export WORKER_COUNT=10
export HTTP_TIMEOUT=3s
export MAX_RETRIES=3
export BASE_DELAY=500ms
go run ./cmd/main.go
```

### Option 2: Using Make

```bash
# Terminal 1
make run-external

# Terminal 2
make run-middleware
```

### Option 3: Docker Compose (Production-like)

```bash
# Build and start all services
docker-compose up --build

# Or using Make
make docker-up

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Testing the System

**Send an event**:
```bash
curl -X POST http://localhost:8080/integrations/events \
  -H "Content-Type: application/json" \
  -d '{
    "source": "monitoring-system",
    "event_type": "server_down",
    "severity": "critical",
    "message": "Production server is not responding",
    "metadata": {
      "server_id": "prod-web-01",
      "region": "us-east-1"
    }
  }'
```

**Expected Response**:
```json
{
  "status": "accepted",
  "event_id": "a1b2c3d4...",
  "correlation_id": "e5f6g7h8..."
}
```

**Check Logs**:
```bash
# Middleware logs (shows event acceptance and processing)
docker-compose logs middleware

# External endpoint logs (shows received alert)
docker-compose logs external-endpoint
```

## Configuration

### Middleware Service Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `EXTERNAL_ENDPOINT_URL` | `http://localhost:8081/external/alerts` | Target endpoint for events |
| `QUEUE_SIZE` | `1000` | Event queue buffer size |
| `WORKER_COUNT` | `10` | Number of worker goroutines |
| `HTTP_TIMEOUT` | `3s` | HTTP request timeout |
| `MAX_RETRIES` | `3` | Maximum retry attempts |
| `BASE_DELAY` | `500ms` | Initial retry delay |

### External Endpoint Service Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8081` | HTTP server port |

### Severity to Priority Mapping

| Severity | Priority | Examples |
|----------|----------|----------|
| `critical`, `fatal`, `emergency` | `critical` | System down, data loss |
| `high`, `error` | `high` | Service degraded, errors |
| `medium`, `warning`, `warn` | `medium` | Performance issues |
| Other values | `low` | Info, debug, trace |

## Testing

### Run Tests

```bash
# All tests
make test

# Package tests only
go test ./pkg/... -v

# Service tests
cd services/middleware && go test ./... -v
cd services/external-endpoint && go test ./... -v
```

### Build Services

```bash
# Build all
make build

# Build individual services
make build-middleware
make build-external

# Binaries created in ./bin/
./bin/middleware
./bin/external-endpoint
```

## Future Improvements

### Operational

1. **Dead Letter Queue (DLQ)**
   - Store failed events after retry exhaustion
   - Manual replay or automated retry with backoff
   - Investigate root causes without data loss

2. **Observability**
   - Distributed tracing (OpenTelemetry, Jaeger)
   - Metrics (Prometheus, Grafana)
   - Request/response times, queue depth, worker utilization
   - Error rates and retry statistics

3. **Health Checks**
   - Liveness probes: `/health`
   - Readiness probes: Check queue, worker pool, downstream dependencies
   - Kubernetes-compatible endpoints

### Security

4. **Authentication & Authorization**
   - API key validation
   - JWT token validation
   - Rate limiting per client
   - Request signing for external integrations

5. **Encryption**
   - TLS/HTTPS for all HTTP traffic
   - Sensitive data encryption at rest
   - Secret management (HashiCorp Vault, AWS Secrets Manager)

### Resilience

6. **Circuit Breaker**
   - Prevent cascade failures
   - Fast-fail when downstream is degraded
   - Automatic recovery detection

7. **Rate Limiting**
   - Per-client limits
   - Global system limits
   - Token bucket or leaky bucket algorithm

8. **Bulkhead Pattern**
   - Separate worker pools per integration
   - Isolate failures to specific downstream systems

### Functionality

9. **Event Persistence**
   - Store events in database before processing
   - Event replay capability
   - Audit trail

10. **Webhook Registry**
    - Dynamic endpoint registration
    - Multiple target endpoints per event type
    - Endpoint health monitoring

11. **Event Filtering & Routing**
    - Route events based on type, source, priority
    - Filter events before processing
    - Conditional routing rules

12. **Batching**
    - Batch multiple events in single HTTP request
    - Configurable batch size and flush interval
    - Reduce HTTP overhead for high-volume scenarios

### Development

13. **Integration Tests**
    - End-to-end test scenarios
    - Docker-based test environment
    - Chaos engineering tests

14. **Load Testing**
    - Benchmark throughput and latency
    - Identify bottlenecks
    - Capacity planning

15. **API Versioning**
    - Support multiple API versions
    - Backward compatibility
    - Graceful deprecation

---

## License

This project is for evaluation purposes.

## Contact

For questions or issues, please contact the development team.
# smartcom-tech-test
