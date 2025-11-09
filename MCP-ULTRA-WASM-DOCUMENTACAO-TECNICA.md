# MCP-ULTRA-WASM-DOCUMENTACAO-TECNICA

**Versão:** 9.0.0  
**Data:** 2025-11-09  
**Status:** ✅ ULTRA VERIFIED CERTIFIED  
**Licença:** MIT  

---

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Arquitetura Geral](#arquitetura-geral)
3. [Componentes MCP Ultra WASM](#componentes-mcp-ultra-wasm)
4. [SDK Ultra WASM](#sdk-ultra-wasm)
5. [Integração entre MCP e SDK](#integração-entre-mcp-e-sdk)
6. [Arquitetura WASM](#arquitetura-wasm)
7. [Ciclo de Vida](#ciclo-de-vida)
8. [Mecanismos de Segurança](#mecanismos-de-segurança)
9. [Performance e Monitoramento](#performance-e-monitoramento)
10. [Deploy e Operação](#deploy-e-operação)
11. [Guia de Implementação](#guia-de-implementação)
12. [Boas Práticas](#boas-práticas)
13. [Troubleshooting](#troubleshooting)

---

## 🎯 Visão Geral

O **MCP Ultra WASM** é uma plataforma enterprise-grade para construção de aplicações SaaS inteligentes com suporte a WebAssembly (WASM) e Model Context Protocol (MCP). A plataforma é composta por dois componentes principais:

### Estrutura Bipartida

```
┌─────────────────────────────────────────────────────────────┐
│                    MCP Ultra WASM                         │
├─────────────────────┬───────────────────────────────────┤
│      MCP Core      │         SDK Ultra WASM          │
├─────────────────────┼───────────────────────────────────┤
│  Aplicação Principal│      Framework de Extensão        │
│  Event-Driven       │      Plugins e Contratos           │
│  Multi-tenant       │      Auto-registro                │
│  Observabilidade     │      Type-safe Registry           │
└─────────────────────┴───────────────────────────────────┘
```

**MCP Core**: Aplicação principal com arquitetura enterprise, agents MCP, multi-tenancy e observabilidade completa.

**SDK Ultra WASM**: Framework de extensão que permite criar plugins personalizados sem modificar o código base, com suporte a WebAssembly.

### Características Principais

- **🏗️ Enterprise Architecture**: Clean Architecture + Event-Driven
- **🤖 MCP Agents**: Sistema cognitivo baseado em Model Context Protocol
- **🔌 Plugin System**: Framework extensível com auto-registro
- **🌐 WASM Support**: WebAssembly para high-performance e sandboxing
- **🏢 Multi-tenancy**: Isolamento completo via Row Level Security
- **📊 Observabilidade**: Prometheus + Grafana + Jaeger + OpenTelemetry
- **🔒 Security**: JWT + RBAC + LGPD/GDPR ready
- **⚡ Performance**: NATS JetStream + Redis + PostgreSQL
- **🚀 Production Ready**: 100% validação (20/20 score)

---

## 🏗️ Arquitetura Geral

### Visão Arquitetural

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway                              │
│                (HTTP/gRPC + Auth + CORS)                     │
└─────────────┬───────────────────────────────────────────────┘
              │
    ┌─────────▼─────────────────────────────────────────────┐
    │                  MCP Core                              │
    │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
    │  │   Handlers   │  │   Services   │  │  Repositories│ │
    │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
    │         │                 │                 │         │
    │  ┌──────▼──────┐   ┌─────▼─────┐   ┌─────▼─────┐   │
    │  │  Event Bus  │   │ Database  │   │    Cache   │   │
    │  │ (NATS)      │   │(PostgreSQL)│   │  (Redis)   │   │
    │  └──────┬──────┘   └───────────┘   └───────────┘   │
    │         │                                        │
    │  ┌──────▼──────────────────────────────────────┐   │
    │  │            MCP Agents                     │   │
    │  │  Seed | Trainer | Evaluator | Reflector │   │
    │  └─────────────────────────────────────────────┘   │
    └─────────────────────────────────────────────────────┘
                          │
    ┌─────────────────▼───────────────────────────────────┐
    │                  SDK Ultra WASM                    │
    │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
    │  │   Bootstrap  │  │   Registry   │  │  Contracts  │ │
    │  │   System     │  │   System     │  │    System    │ │
    │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
    │         │                 │                 │         │
    │  ┌──────▼─────────────────────────────────────┐   │
    │  │           Plugin Layer (WASM Ready)      │   │
    │  │  ┌──────────┐  ┌──────────┐  ┌──────────┐ │   │
    │  │  │ Plugin A │  │ Plugin B  │  │ Plugin N │ │   │
    │  │  └──────────┘  └──────────┘  └──────────┘ │   │
    │  └─────────────────────────────────────────────┘   │
    └─────────────────────────────────────────────────────┘
```

### Fluxo de Dados

1. **Request**: HTTP/gRPC → API Gateway (auth + validation)
2. **Processing**: Handler → Service (business logic)
3. **Persistence**: Repository → Database + Cache
4. **Events**: Service → Event Bus (NATS) → MCP Agents
5. **Extensions**: SDK Registry → Plugin Layer → Custom Logic
6. **Monitoring**: Observability stack captura métricas em todos os pontos

---

## 🧩 Componentes MCP Ultra WASM

### MCP Core

**Localização**: `E:\vertikon\.endurance\templates\mcp-ultra-wasm\mcp\mcp-ultra-wasm`

#### Estrutura Principal

```
mcp/mcp-ultra-wasm/
├── cmd/                     # Entry points
│   └── mcp-model-ultra/
├── internal/               # Lógica interna
│   ├── config/             # Configuração
│   ├── handlers/           # HTTP handlers
│   ├── services/           # Business logic
│   ├── repository/         # Data access
│   ├── domain/             # Domain models
│   ├── events/             # Event handlers
│   ├── ai/                 # AI components
│   ├── analytics/          # Analytics
│   ├── cache/              # Cache layer
│   ├── compliance/         # Compliance (LGPD/GDPR)
│   ├── dashboard/          # Admin dashboard
│   ├── features/           # Feature flags
│   ├── lifecycle/          # Lifecycle management
│   ├── metrics/            # Metrics collection
│   ├── observability/      # Observability
│   ├── ratelimit/          # Rate limiting
│   ├── security/           # Security components
│   ├── telemetry/          # Telemetry
│   └── tracing/            # Distributed tracing
├── pkg/                    # Bibliotecas reutilizáveis
│   ├── httpx/              # HTTP utilities
│   ├── logger/             # Logging framework
│   ├── metrics/            # Prometheus metrics
│   ├── observability/      # OpenTelemetry
│   ├── redisx/             # Redis utilities
│   └── types/              # Type definitions
├── migrations/             # Database migrations
├── test/                   # Test infrastructure
├── deploy/                 # Deployment manifests
├── grafana/                # Grafana dashboards
└── scripts/                # Automation scripts
```

#### Principais Features

**1. Event-Driven Architecture**
- NATS JetStream para mensageria assíncrona
- Schemas validados para todos os eventos
- Retry automático e dead letter queue
- Event sourcing para audit trail

**2. Multi-tenancy**
- Row Level Security (RLS) no PostgreSQL
- Isolamento completo de dados por tenant
- Rate limiting por tenant
- Configuração de limites personalizável

**3. Agentes MCP**
- **Seed Agent**: Inicialização de contexto
- **Trainer Agent**: Aprendizado contínuo
- **Evaluator Agent**: Avaliação de qualidade
- **Reflector Agent**: Auto-análise e melhoria

**4. Observabilidade**
- Prometheus metrics personalizadas
- Jaeger distributed tracing
- Grafana dashboards pré-configurados
- Structured logging com contexto

---

## 🔧 SDK Ultra WASM

**Localização**: `E:\vertikon\.endurance\templates\mcp-ultra-wasm\sdk\sdk-ultra-wasm`

### Arquitetura do SDK

```
sdk/sdk-ultra-wasm/
├── cmd/                    # Entry points
│   ├── ultra-sdk-cli/     # CLI scaffolding
│   └── main.go             # Servidor principal
├── pkg/                    # Framework core
│   ├── bootstrap/          # Inicialização do SDK
│   │   ├── bootstrap.go
│   │   └── health.go
│   ├── contracts/          # Contratos de extensão
│   │   ├── route.go        # RouteInjector
│   │   ├── middleware.go   # MiddlewareInjector
│   │   ├── job.go          # Job interface
│   │   ├── service.go      # Service interface
│   │   └── version.go      # SemVer
│   ├── registry/           # Plugin registry
│   │   └── registry.go
│   ├── router/             # HTTP abstrações
│   │   ├── mux.go          # Gorilla Mux wrapper
│   │   └── middleware/     # Built-in middlewares
│   └── policies/           # Security policies
│       ├── jwt.go          # JWT authentication
│       ├── rbac.go         # Role-based access control
│       └── context.go      # Identity context
├── internal/               # Lógica interna
│   └── handlers/           # HTTP handlers
└── seed-examples/          # Exemplos de plugins
    └── waba/               # WhatsApp Business API
```

### Contratos do SDK

#### 1. RouteInjector
```go
type RouteInjector interface {
    Name() string
    Version() string
    Routes() []Route
}
```

#### 2. MiddlewareInjector
```go
type MiddlewareInjector interface {
    Name() string
    Priority() int
    Middleware() func(http.Handler) http.Handler
}
```

#### 3. Job
```go
type Job interface {
    Name() string
    Schedule() string
    Run(ctx context.Context) error
}
```

#### 4. Service
```go
type Service interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() error
}
```

### Sistema de Registry

```go
// Auto-registro type-safe
func init() {
    _ = registry.Register("my-plugin", &Plugin{})
}

// Registry com segregação por tipo
type PluginRegistry struct {
    routes      map[string]RouteInjector
    middlewares map[string]MiddlewareInjector
    jobs        map[string]Job
    services    map[string]Service
    mu          sync.RWMutex
}
```

---

## 🔗 Integração entre MCP e SDK

### Comunicação entre Componentes

```
┌─────────────────┐    NATS Events    ┌─────────────────┐
│   MCP Core      │◄─────────────────►│   SDK Ultra     │
│                 │                   │    WASM         │
│  Events:        │                   │  Extensions:    │
│  • user.created │                   │  • custom.routes│
│  • payment.done│                   │  • custom.jobs  │
│  • agent.cycle  │                   │  • custom.services│
└─────────────────┘                   └─────────────────┘
         ▲                                    ▲
         │                                    │
┌─────────────────┐                 ┌─────────────────┐
│   PostgreSQL   │                 │    Plugins      │
│                 │                 │                 │
│  • users        │                 │  • Business     │
│  • payments     │                 │    Logic       │
│  • audit_log    │                 │  • Integration  │
└─────────────────┘                 └─────────────────┘
```

### Fluxo de Integração

1. **Event Publication**: MCP Core publica eventos no NATS
2. **Plugin Subscriptions**: SDK plugins consomem eventos relevantes
3. **Custom Processing**: Plugins implementam lógica específica
4. **Feedback Loop**: Plugins podem publicar eventos de volta

### Exemplo de Integração

```go
// MCP Core publica evento
event := UserCreatedEvent{
    UserID:    user.ID,
    TenantKey: user.TenantKey,
    Email:     user.Email,
}
nats.Publish("users.created", event)

// SDK Plugin consome e processa
func (p *AnalyticsPlugin) handleUserCreated(msg *nats.Msg) {
    var event UserCreatedEvent
    json.Unmarshal(msg.Data, &event)
    
    // Lógica custom do plugin
    p.trackUserAnalytics(event)
}
```

---

## 🌐 Arquitetura WASM

### WebAssembly Integration

O MCP Ultra WASM suporta WebAssembly para:

**1. Performance**
- Execução em velocidade nativa
- Sandbox seguro para código customizado
- Baixo consumo de memória

**2. Portabilidade**
- Cross-platform deployment
- Container-friendly
- Edge computing ready

**3. Segurança**
- Sandboxing automático
- Resource limits
- Memory safety

### Implementação WASM

```go
// Plugin compilado para WASM
//go:build js,wasm

package main

import (
    "context"
    "encoding/json"
)

//export GetUserAnalytics
func GetUserAnalytics(userID string) ([]byte, error) {
    analytics := calculateAnalytics(userID)
    return json.Marshal(analytics)
}

//export ProcessPayment
func ProcessPayment(paymentData []byte) ([]byte, error) {
    var payment Payment
    err := json.Unmarshal(paymentData, &payment)
    if err != nil {
        return nil, err
    }
    
    result := processPaymentInternal(payment)
    return json.Marshal(result)
}

// Função interna do WASM
func calculateAnalytics(userID string) UserAnalytics {
    // Lógica de analytics
    return UserAnalytics{
        UserID:    userID,
        Events:    getEventsForUser(userID),
        Metrics:   calculateMetrics(userID),
    }
}
```

### Compilação WASM

```bash
# Compilar plugin para WASM
GOOS=js GOARCH=wasm go build \
    -o plugin.wasm \
    ./plugins/analytics/main.go

# Deploy no SDK
curl -X POST http://localhost:8080/plugins/wasm \
  -F "file=@plugin.wasm"
```

---

## 🔄 Ciclo de Vida

### Inicialização

```
┌─────────────────────────────────────────────────────────────┐
│                    Bootstrap Sequence                    │
│                                                             │
│  1. Load Configuration (env vars, vault, etc)              │
│  2. Initialize Observability (prometheus, jaeger)          │
│  3. Setup Database Connections (PostgreSQL, Redis)          │
│  4. Start Event Bus (NATS JetStream)                       │
│  5. Initialize MCP Agents                                   │
│  6. Bootstrap SDK Registry                                   │
│  7. Load Plugins (auto-registro via init())                  │
│  8. Start HTTP Server                                         │
└─────────────────────────────────────────────────────────────┘
```

### Operação

```
┌─────────────────────────────────────────────────────────────┐
│                     Runtime Flow                         │
│                                                             │
│  HTTP Request → Auth → Handler → Service → Repository       │
│       │                                                    │
│       ▼                                                    │
│  Event Publication → NATS → MCP Agent → Processing          │
│       │                                                    │
│       ▼                                                    │
│  SDK Plugin → Custom Logic → Response → Client               │
│                                                             │
│  Metrics & Tracing captured at all layers                    │
└─────────────────────────────────────────────────────────────┘
```

### Shutdown

```
┌─────────────────────────────────────────────────────────────┐
│                 Graceful Shutdown                        │
│                                                             │
│  1. Stop accepting new requests                           │
│  2. Wait for active requests to complete                    │
│  3. Stop MCP Agents                                        │
│  4. Shutdown Plugin Registry                               │
│  5. Close Database Connections                             │
│  6. Flush Metrics & Logs                                   │
│  7. Exit                                                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔒 Mecanismos de Segurança

### Security Stack

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Layers                         │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                Network Security                     │   │
│  │  • TLS 1.3+ (mTLS available)                        │   │
│  │  • CORS Configuration                              │   │
│  │  • Rate Limiting                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                              ▲                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Application Security                    │   │
│  │  • JWT Authentication                             │   │
│  │  • RBAC Authorization                             │   │
│  │  • API Key Management                            │   │
│  │  • Input Validation (JSON Schemas)               │   │
│  └─────────────────────────────────────────────────────┘   │
│                              ▲                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                 Data Security                          │   │
│  │  • Encryption at Rest (AES-256)                     │   │
│  │  • Encryption in Transit                           │   │
│  │  • PII Scanning & Masking                           │   │
│  │  • Data Retention Policies                        │   │
│  └─────────────────────────────────────────────────────┘   │
│                              ▲                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Infrastructure Security                 │   │
│  │  • Secrets Management (Vault/K8s)                  │   │
│  │  • Container Security (Docker)                     │   │
│  │  • Kubernetes Network Policies                   │   │
│  │  • Security Scanning (Grype/Trivy)                  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### JWT + RBAC

```go
// JWT Token Structure
type JWTCustomClaims struct {
    jwt.RegisteredClaims
    UserID    string   `json:"user_id"`
    TenantKey string   `json:"tenant_key"`
    Roles     []string `json:"roles"`
    Metadata  map[string]interface{} `json:"metadata"`
}

// RBAC Implementation
func (p *RBACPolicy) CheckPermission(
    ctx context.Context, 
    resource string, 
    action string,
) error {
    identity := FromIdentity(ctx)
    if identity == nil {
        return ErrUnauthorized
    }
    
    // Check role-based permissions
    for _, role := range identity.Roles {
        if p.roleHasPermission(role, resource, action) {
            return nil
        }
    }
    
    return ErrForbidden
}
```

### Multi-tenancy Security

```sql
-- Row Level Security Policy
CREATE POLICY tenant_isolation ON resources
    FOR ALL
    TO application_user
    USING (
        tenant_key = current_setting('app.current_tenant')::VARCHAR
    );

-- Automatic Tenant Context
CREATE OR REPLACE FUNCTION set_tenant_context()
RETURNS trigger AS $$
BEGIN
    PERFORM set_config(
        'app.current_tenant',
        COALESCE(NULLIF(current_setting('request.jwt.claims'), '{}')::jsonb
            ->> 'tenant_key',
        'unknown'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

---

## 📊 Performance e Monitoramento

### Observability Stack

```
┌─────────────────────────────────────────────────────────────┐
│                 Observability Pipeline                     │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐          │
│  │   HTTP     │  │  Business  │  │   System    │          │
│  │  Requests  │  │   Logic    │  │  Metrics    │          │
│  │             │  │             │  │             │          │
│  │    │        │  │      │     │  │      │     │          │
│  ▼    ▼        ▼  ▼      ▼     ▼  ▼      ▼     ▼          │
│  ┌────▼────┐  ┌─────▼─────┐  ┌─────▼─────┐          │
│  │ Prometheus│  │   Jaeger   │  │   Loki     │          │
│  │  Metrics  │  │  Tracing   │  │   Logs    │          │
│  └───────────┘  └───────────┘  └───────────┘          │
│                              ▲                              │
│  ┌─────────────────────────────────────────────┐          │
│  │              Grafana Dashboards          │          │
│  │  • Overview                              │          │
│  │  • MCP Agents                            │          │
│  │  • WASM Performance                    │          │
│  │  • Business Metrics                     │          │
│  └─────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Métricas Principais

#### HTTP Metrics
```go
// HTTP Request Metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)
```

#### MCP Agent Metrics
```go
// Agent Performance Metrics
var (
    agentCyclesTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_agent_cycles_total",
            Help: "Total number of agent cycles",
        },
        []string{"agent_type", "tenant_key"},
    )
    
    agentProcessingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "mcp_agent_processing_duration_seconds",
            Help: "Agent processing duration in seconds",
        },
        []string{"agent_type"},
    )
)
```

#### WASM Performance Metrics
```go
// WASM Plugin Metrics
var (
    wasmPluginExecutions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "wasm_plugin_executions_total",
            Help: "Total WASM plugin executions",
        },
        []string{"plugin_name", "status"},
    )
    
    wasmPluginDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "wasm_plugin_duration_seconds",
            Help: "WASM plugin execution duration",
            Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
        },
        []string{"plugin_name"},
    )
)
```

### Distributed Tracing

```go
// OpenTelemetry Tracing
func (s *Service) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    ctx, span := tracer.Start(ctx, "service.process_request")
    defer span.End()
    
    // Add span attributes
    span.SetAttributes(
        attribute.String("service.name", "mcp-ultra-wasm"),
        attribute.String("user.id", req.UserID),
        attribute.String("tenant.key", req.TenantKey),
        attribute.String("request.id", req.ID),
    )
    
    // Process request
    result, err := s.doProcessRequest(ctx, req)
    
    // Record error if present
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else {
        span.SetStatus(codes.Ok, "Request processed successfully")
    }
    
    return result, err
}
```

---

## 🚀 Deploy e Operação

### Docker Deployment

#### Multi-stage Dockerfile
```dockerfile
# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=js GOARCH=wasm go build \
    -o /app/main.wasm \
    ./cmd/main.go

# Runtime Stage
FROM scratch

WORKDIR /
COPY --from=builder /app/main.wasm .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main.wasm"]
```

#### Docker Compose
```yaml
version: '3.8'

services:
  mcp-ultra-wasm:
    build:
      context: .
      dockerfile: Dockerfile.wasm
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - NATS_URL=nats://nats:4222
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      nats:
        condition: service_healthy

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: mcp_ultra_wasm
      POSTGRES_USER: mcp_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U mcp_user -d mcp_ultra_wasm"]

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]

  nats:
    image: nats:2.10-alpine
    command: ["--jetstream", "--store_dir", "/nats-data"]
    ports:
      - "8222:8222"
    volumes:
      - nats_data:/nats-data

volumes:
  postgres_data:
  redis_data:
  nats_data:
```

### Kubernetes Deployment

#### Deployment Manifest
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-ultra-wasm
  labels:
    app: mcp-ultra-wasm
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-ultra-wasm
  template:
    metadata:
      labels:
        app: mcp-ultra-wasm
        version: v1
    spec:
      containers:
      - name: mcp-ultra-wasm
        image: mcp-ultra-wasm:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: host
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

#### HPA Configuration
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mcp-ultra-wasm-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: mcp-ultra-wasm
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

---

## 🛠️ Guia de Implementação

### Setup Inicial

#### 1. Clonar Repositório
```bash
git clone https://github.com/vertikon/mcp-ultra-wasm-wasm.git
cd mcp-ultra-wasm
```

#### 2. Configurar Ambiente
```bash
# Copiar variáveis de ambiente
cp .env.example .env

# Editar configurações
vim .env

# Gerar secrets seguros
openssl rand -base64 64 > .jwt-secret
openssl rand -base64 32 > .encryption-key
```

#### 3. Iniciar Serviços
```bash
# Via Docker Compose
docker-compose up -d

# Verificar status
docker-compose ps

# Verificar logs
docker-compose logs -f mcp-ultra-wasm
```

### Criar Plugin Personalizado

#### 1. Estrutura do Plugin
```go
// plugins/analytics/plugin.go
package analytics

import (
    "context"
    "encoding/json"
    "net/http"
    
    "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
    "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
)

func init() {
    // Auto-registro do plugin
    _ = registry.Register("analytics", &Plugin{})
}

type Plugin struct {
    config Config
    client AnalyticsClient
}

// Implementação dos contratos
func (p *Plugin) Name() string    { return "analytics" }
func (p *Plugin) Version() string { return "1.0.0" }

func (p *Plugin) Routes() []contracts.Route {
    return []contracts.Route{
        {
            Method:  "GET",
            Path:    "/analytics/events",
            Handler: p.getEvents,
        },
        {
            Method:  "POST",
            Path:    "/analytics/track",
            Handler: p.trackEvent,
        },
    }
}

func (p *Plugin) Start(ctx context.Context) error {
    p.client = NewAnalyticsClient(p.config.APIKey)
    return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
    return p.client.Close()
}

func (p *Plugin) Health() error {
    return p.client.Ping()
}
```

#### 2. Compilar Plugin para WASM
```bash
# Compilar para WASM
GOOS=js GOARCH=wasm go build \
    -o analytics.wasm \
    -ldflags="-s -w" \
    ./plugins/analytics/plugin.go

# Deploy no MCP
curl -X POST http://localhost:8080/sdk/plugins \
  -H "Content-Type: application/wasm" \
  --data-binary @analytics.wasm
```

#### 3. Testar Plugin
```bash
# Test localmente
curl http://localhost:8080/analytics/events \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Key: tenant-123"

# Testar métricas
curl http://localhost:8080/metrics | grep analytics
```

### Configurar Agentes MCP

#### 1. Configurar Agent Seed
```go
// internal/agents/seed.go
type SeedAgent struct {
    config SeedConfig
    nats   *nats.Conn
}

func (a *SeedAgent) InitializeTenant(ctx context.Context, req SeedRequest) error {
    // Criar contexto inicial do tenant
    context := TenantContext{
        TenantKey:    req.TenantKey,
        UserID:       req.UserID,
        Preferences:  req.Preferences,
        CreatedAt:    time.Now(),
    }
    
    // Salvar no banco
    err := a.saveTenantContext(ctx, context)
    if err != nil {
        return err
    }
    
    // Publicar evento de inicialização
    event := TenantInitializedEvent{
        TenantKey: req.TenantKey,
        UserID:    req.UserID,
        Timestamp: time.Now(),
    }
    
    return a.nats.Publish("tenant.initialized", event)
}
```

#### 2. Configurar Agent Trainer
```go
// internal/agents/trainer.go
type TrainerAgent struct {
    interval time.Duration
    nats     *nats.Conn
    models   map[string]*Model
}

func (a *TrainerAgent) RunTrainingCycle(ctx context.Context) error {
    // Coletar dados de treinamento
    trainingData, err := a.collectTrainingData(ctx)
    if err != nil {
        return err
    }
    
    // Treinar modelos
    for modelType, model := range a.models {
        err := model.Train(ctx, trainingData)
        if err != nil {
            log.Printf("Error training model %s: %v", modelType, err)
            continue
        }
    }
    
    // Avaliar performance
    metrics := a.evaluateModels(ctx)
    
    // Publicar resultados
    event := TrainingCompletedEvent{
        Timestamp:    time.Now(),
        ModelMetrics: metrics,
    }
    
    return a.nats.Publish("training.completed", event)
}
```

---

## ✅ Boas Práticas

### Code Organization

#### 1. Structure de Diretórios
```
project/
├── cmd/                     # Entry points
├── internal/               # Lógica interna (não exportável)
├── pkg/                    # Bibliotecas reutilizáveis
├── plugins/                 # Plugins personalizados
├── tests/                   # Testes
├── docs/                    # Documentação
└── deployments/             # Configurações de deploy
```

#### 2. Naming Conventions
```go
// Interfaces e tipos exportados
type RouteInjector interface {
    Name() string
    Version() string
    Routes() []Route
}

// Implementações não exportadas
type plugin struct {
    name    string
    version string
    client  *Client
}

// Variáveis privadas
var (
    defaultTimeout = 30 * time.Second
    maxRetries     = 3
)

// Constantes exportadas
const (
    DefaultPort    = 8080
    DefaultTimeout = 30 * time.Second
)
```

#### 3. Error Handling
```go
// Tipos de erro específicos
var (
    ErrPluginNotFound   = errors.New("plugin not found")
    ErrInvalidVersion   = errors.New("invalid plugin version")
    ErrUnauthorized    = errors.New("unauthorized")
)

// Error com contexto
type PluginError struct {
    Code    string
    Message string
    Cause   error
}

func (e *PluginError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
```

### Testing Strategy

#### 1. Unit Tests
```go
func TestPlugin_Register(t *testing.T) {
    tests := []struct {
        name        string
        plugin      interface{}
        expectError bool
    }{
        {
            name:        "valid plugin",
            plugin:      &mockPlugin{},
            expectError: false,
        },
        {
            name:        "invalid plugin",
            plugin:      &invalidPlugin{},
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            registry.Reset()
            
            err := registry.Register("test", tt.plugin)
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### 2. Integration Tests
```go
func TestPlugin_Integration(t *testing.T) {
    ctx := context.Background()
    
    // Setup test environment
    pgContainer, err := postgres.RunContainer(ctx, testcontainers.WithDatabaseName("test"))
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)
    
    // Test plugin com banco real
    db := setupTestDB(pgContainer.ConnectionString())
    plugin := &AnalyticsPlugin{db: db}
    
    err = plugin.Start(ctx)
    assert.NoError(t, err)
    
    // Test functionality
    result, err := plugin.ProcessEvent(ctx, Event{Type: "click"})
    assert.NoError(t, err)
    assert.NotNil(t, result)
    
    // Cleanup
    err = plugin.Stop(ctx)
    assert.NoError(t, err)
}
```

#### 3. E2E Tests
```go
func TestE2E_AnalyticsWorkflow(t *testing.T) {
    // Setup infrastructure
    env := testutils.NewTestEnvironment(t)
    defer env.Cleanup()
    
    // Create client
    client := http.Client{}
    baseURI := env.GetBaseURI()
    
    // Test complete workflow
    // 1. Login
    token := login(t, client, baseURI)
    
    // 2. Track event
    resp := trackEvent(t, client, baseURI, token, Event{
        Type:   "purchase",
        UserID: "user123",
        Value:  99.99,
    })
    
    // 3. Verify analytics
    analytics := getAnalytics(t, client, baseURI, token)
    assert.Contains(t, analytics.Events, resp.ID)
    
    // 4. Verify metrics
    metrics := getMetrics(t, client, baseURI)
    assert.Greater(t, metrics.EventCount, 0)
}
```

### Performance Optimization

#### 1. Database Optimization
```go
// Connection pooling
func setupDatabase(cfg DatabaseConfig) (*sql.DB, error) {
    config, err := pgxpool.ParseConfig(cfg.URL)
    if err != nil {
        return nil, err
    }
    
    // Otimizações de performance
    config.MaxConns = 50
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.HealthCheckPeriod = 30 * time.Second
    
    return pgxpool.ConnectConfig(context.Background(), config)
}

// Prepared statements
func (r *Repository) GetEvents(ctx context.Context, userID string) ([]Event, error) {
    query := `
        SELECT id, type, user_id, data, created_at
        FROM events
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2
    `
    
    rows, err := r.db.Query(ctx, query, userID, 100)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    return scanEvents(rows)
}
```

#### 2. Caching Strategy
```go
type Cache struct {
    redis *redis.Client
    ttl   time.Duration
}

func (c *Cache) GetOrSet(
    ctx context.Context,
    key string,
    fetcher func() (interface{}, error),
) (interface{}, error) {
    // Tentar cache
    cached, err := c.redis.Get(ctx, key).Result()
    if err == nil {
        var result interface{}
        if json.Unmarshal([]byte(cached), &result) == nil {
            return result, nil
        }
    }
    
    // Cache miss - buscar e armazenar
    result, err := fetcher()
    if err != nil {
        return nil, err
    }
    
    data, err := json.Marshal(result)
    if err != nil {
        return nil, err
    }
    
    // Armazenar em cache
    c.redis.Set(ctx, key, data, c.ttl)
    
    return result, nil
}
```

#### 3. Concurrent Processing
```go
func (p *EventProcessor) ProcessEvents(ctx context.Context) error {
    // Worker pool
    workers := 10
    jobs := make(chan Event, workers*2)
    
    // Start workers
    for i := 0; i < workers; i++ {
        go p.worker(ctx, jobs)
    }
    
    // Process events
    for {
        select {
        case event := <-p.events:
            jobs <- event
        case <-ctx.Done():
            close(jobs)
            return ctx.Err()
        }
    }
}

func (p *EventProcessor) worker(ctx context.Context, jobs <-chan Event) {
    for event := range jobs {
        if err := p.processEvent(ctx, event); err != nil {
            log.Printf("Error processing event %s: %v", event.ID, err)
        }
    }
}
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. Plugin Registration Failed

**Sintoma**: Plugin não aparece no registry

**Causas Comuns**:
- Implementação incorreta das interfaces
- Erro no auto-registro via `init()`
- Nome duplicado no registry

**Soluções**:
```go
// Verificar implementação
var _ contracts.RouteInjector = (*Plugin)(nil)

// Verificar auto-registro
func init() {
    if err := registry.Register("unique-name", &Plugin{}); err != nil {
        log.Fatal("Failed to register plugin:", err)
    }
}

// Debug do registry
log.Printf("Registered plugins: %v", registry.ListPlugins())
```

#### 2. NATS Connection Issues

**Sintoma**: Eventos não são publicados/consumidos

**Causas Comuns**:
- NATS server não iniciado
- Configuração de conexão incorreta
- Falha de autenticação

**Soluções**:
```bash
# Verificar status do NATS
docker-compose logs nats

# Testar conexão
telnet localhost 4222

# Verificar configuração
nats-server -js -m 8222 -sd /nats-data
```

#### 3. WASM Plugin Issues

**Sintoma**: Plugin WASM não funciona corretamente

**Causas Comuns**:
- Compilação incorreta para WASM
- Dependências não suportadas em WASM
- Resource limits excedidos

**Soluções**:
```bash
# Verificar dependências compatíveis
go list -m | grep "cgo"  # deve estar vazio

# Compilar com flags corretas
GOOS=js GOARCH=wasm go build \
    -ldflags="-s -w" \
    -o plugin.wasm \
    ./plugin.go

# Verificar tamanho do arquivo
ls -la plugin.wasm  # deve ser < 10MB para maioria dos casos
```

#### 4. Performance Issues

**Sintoma**: Alta latência ou timeouts

**Diagnóstico**:
```bash
# Verificar métricas
curl http://localhost:8080/metrics | grep histogram

# Verificar traces
curl http://localhost:16686/api/traces?service=mcp-ultra-wasm

# Verificar logs
kubectl logs -f deployment/mcp-ultra-wasm
```

**Otimizações**:
```go
// Adicionar cache
result, err := cache.GetOrSet("key", func() (interface{}, error) {
    return expensiveOperation()
})

// Usar connection pooling
db, err := setupDatabasePool(config)

// Limitar goroutines
semaphore := make(chan struct{}, 10)
```

### Debug Tools

#### 1. Registry Inspector
```go
// Endpoint de debug para inspecion do registry
func (h *DebugHandler) InspectRegistry(w http.ResponseWriter, r *http.Request) {
    info := map[string]interface{}{
        "plugins": registry.ListPlugins(),
        "routes":  registry.ListRoutes(),
        "jobs":    registry.ListJobs(),
        "services": registry.ListServices(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}
```

#### 2. Event Tracer
```go
// NATS message tracer
func (t *EventTracer) TraceMessage(msg *nats.Msg) {
    var event map[string]interface{}
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        log.Printf("Failed to unmarshal event: %v", err)
        return
    }
    
    log.Printf("Event traced: %s -> %s", msg.Subject, event["type"])
    
    // Adicionar ao trace atual
    if span := otel.SpanFromContext(msg.Context); span != nil {
        span.SetAttributes(
            attribute.String("nats.subject", msg.Subject),
            attribute.String("event.type", fmt.Sprintf("%v", event["type"])),
        )
    }
}
```

#### 3. Health Check Avançado
```go
func (h *HealthHandler) DetailedHealth(ctx context.Context) HealthStatus {
    status := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }
    
    // Verificar cada dependência
    status.Checks["database"] = h.checkDatabase(ctx)
    status.Checks["redis"] = h.checkRedis(ctx)
    status.Checks["nats"] = h.checkNATS(ctx)
    status.Checks["wasm_plugins"] = h.checkWASMPlugins(ctx)
    
    // Determinar status geral
    for _, check := range status.Checks {
        if check.Status != "healthy" {
            status.Status = "degraded"
        }
    }
    
    return status
}
```

---

## 📚 Referências e Links

### Documentação Oficial

- **MCP Protocol**: https://modelcontextprotocol.io/
- **WebAssembly**: https://webassembly.org/
- **NATS Documentation**: https://docs.nats.io/
- **OpenTelemetry**: https://opentelemetry.io/
- **Prometheus**: https://prometheus.io/
- **Grafana**: https://grafana.com/

### Ferramentas de Desenvolvimento

- **Go Documentation**: https://golang.org/doc/
- **Docker**: https://docs.docker.com/
- **Kubernetes**: https://kubernetes.io/docs/
- **golangci-lint**: https://golangci-lint.run/

### Padrões e Best Practices

- **Effective Go**: https://golang.org/doc/effective_go.html
- **Go Code Review**: https://github.com/golang/go/wiki/CodeReviewComments
- **Clean Architecture**: https://blog.cleancoder.com/clean-code/
- **Domain-Driven Design**: https://en.wikipedia.org/wiki/Domain-driven_design

### Repositórios Relacionados

- **MCP Ultra**: https://github.com/vertikon/mcp-ultra
- **MCP Ultra SDK**: https://github.com/vertikon/mcp-ultra-sdk-custom
- **Vertikon Templates**: https://github.com/vertikon/

---

## 🆘 Suporte e Contribuição

### Como Obterter Ajuda

**Issues e Bugs**:
- GitHub Issues: https://github.com/vertikon/mcp-ultra-wasm/issues
- Security Issues: security@vertikon.com

**Discussões e Comunidade**:
- GitHub Discussions: https://github.com/vertikon/mcp-ultra-wasm/discussions
- Email: dev@vertikon.com

### Como Contribuir

1. **Fork** o repositório
2. **Criar branch** de feature
3. **Implementar** sua contribuição
4. **Testar** completamente
5. **Submit** Pull Request
6. **Revisar** e mesclar

### Checklist de Contribuição

- [ ] Código compila sem erros
- [ ] Testes passam (>80% cobertura)
- [ ] Linting sem warnings
- [ ] Documentação atualizada
- [ ] CHANGELOG atualizado
- [ ] Version SemVer correta

---

**Desenvolvido com ❤️ pela equipe Vertikon**

*Este documento está em constante evolução. Última atualização: 2025-11-09*

**Status da Plataforma**: ✅ **Production Ready** (Score 20/20)