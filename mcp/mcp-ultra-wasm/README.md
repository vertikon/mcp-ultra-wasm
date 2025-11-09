# 🧠 Vertikon MCP-Ultra

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Validation Score](https://img.shields.io/badge/Validation-20%2F20-success)](docs/JORNADA-100PCT-COMPLETA.md)
[![Code Coverage](https://img.shields.io/badge/Coverage-%E2%89%A580%25-brightgreen)](docs/melhorias/ENHANCED_VALIDATION_REPORT.md)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![GitHub Issues](https://img.shields.io/github/issues/vertikon/mcp-ultra-wasm)](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/issues)
[![GitHub Stars](https://img.shields.io/github/stars/vertikon/mcp-ultra-wasm?style=social)](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/stargazers)

### Plataforma SaaS Inteligente baseada em Model Context Protocol (MCP)

O **MCP-Ultra** é um template **open-source** para construir produtos SaaS inteligentes, com integração nativa entre microserviços, agentes de IA e automação de processos. Template pronto para produção com **100% de validação** (20/20).

> 🎯 **Por que usar MCP-Ultra?**
> Acelere o desenvolvimento de SaaS com arquitetura enterprise-grade, observabilidade completa, multi-tenancy nativo, e sistema cognitivo de IA baseado em MCP. Economize meses de desenvolvimento!

```bash
# Quick Start - 3 comandos para rodar tudo
git clone https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git && cd mcp-ultra-wasm
cp .env.example .env && docker-compose up -d
curl http://localhost:9655/healthz  # ✅ Pronto!
```

---

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Características Principais](#-características-principais)
- [Arquitetura](#-arquitetura)
- [Stack Tecnológica](#-stack-tecnológica)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
  - [Instalação via Docker (Recomendado)](#instalação-via-docker-recomendado)
  - [Instalação Manual](#instalação-manual)
- [Configuração](#-configuração)
- [Uso](#-uso)
- [Agentes MCP](#-agentes-mcp)
- [API](#-api)
- [Desenvolvimento](#-desenvolvimento)
- [Testes](#-testes)
- [Deployment](#-deployment)
- [Observabilidade](#-observabilidade)
- [Segurança e Compliance](#-segurança-e-compliance)
- [Multi-tenancy](#-multi-tenancy)
- [Planos e Billing](#-planos-e-billing)
- [Roadmap](#-roadmap)
- [Contribuindo](#-contribuindo)
- [Licença](#-licença)
- [Suporte](#-suporte)

---

## 🚀 Visão Geral

O **MCP-Ultra** é uma plataforma completa para construção de aplicações SaaS inteligentes com:

- **Arquitetura Event-Driven**: Comunicação via NATS JetStream com schemas validados
- **Clean Architecture**: Separação clara de camadas (handlers, services, repository)
- **Agentes de IA**: Sistema cognitivo baseado em MCP (Model Context Protocol)
- **Multi-tenant**: Isolamento completo via Row Level Security (RLS) no PostgreSQL
- **Observabilidade**: Métricas, tracing e logging prontos para produção
- **Compliance**: LGPD/GDPR ready com audit trail completo

**Status**: ✅ **Pronto para Produção** (Score 20/20)

### 🎯 Casos de Uso Ideais

- 🏢 **B2B SaaS** - CRM, ERP, Billing, etc
- 📊 **Plataformas de Analytics** - Com IA embarcada
- 🤖 **Sistemas Cognitivos** - Agentes autônomos com memória
- 🏗️ **Microserviços** - Template para cada serviço
- 🔄 **Event-Driven Systems** - Workflows complexos

### 💡 O que você ganha "de graça"

- ✅ Autenticação JWT + Multi-tenancy implementados
- ✅ Observabilidade completa (Prometheus + Grafana + Jaeger)
- ✅ Testes com 80%+ de cobertura
- ✅ CI/CD pipeline pronto
- ✅ Docker + Kubernetes manifests
- ✅ Documentação completa
- ✅ Best practices de Go (Clean Architecture, DDD)
- ✅ Segurança e Compliance (LGPD/GDPR)

---

## ✨ Características Principais

### 🎯 Core Features

- ✅ **Event-Driven Architecture** - NATS JetStream com retry e dead letter queue
- ✅ **Clean Architecture** - DDD patterns com separação de concerns
- ✅ **Multi-tenancy** - Isolamento por RLS (Row Level Security)
- ✅ **API REST & gRPC** - Dual protocol support
- ✅ **JWT Authentication** - Auth + TenantKey validation
- ✅ **Rate Limiting** - Por tenant e global
- ✅ **Circuit Breaker** - Proteção contra cascading failures

### 🤖 Agentes MCP (Model Context Protocol)

- **Seed Agent** - Inicialização de contexto e comportamento
- **Trainer Agent** - Aprendizado contínuo (ciclo de 15min)
- **Evaluator Agent** - Avaliação de qualidade e coerência
- **Reflector Agent** - Auto-análise e melhoria automática

### 📊 Observabilidade

- **Prometheus** - Métricas de performance (latência, throughput, errors)
- **Grafana** - Dashboards SaaS pré-configurados
- **Jaeger** - Distributed tracing (OpenTelemetry)
- **Structured Logging** - JSON logs com contexto completo

### 🔒 Segurança & Compliance

- **LGPD/GDPR Ready** - PII scanning, consent tracking, data retention
- **Audit Trail** - Log completo de todas as operações sensíveis
- **Secrets Management** - Suporte a Vault, K8s Secrets
- **TLS/mTLS** - Comunicação segura entre serviços
- **Security Scanning** - Grype + trivy integrados

---

## 🏗️ Arquitetura

O MCP-Ultra segue **Clean Architecture + Event-Driven**, com foco em modularidade e escalabilidade:

```
┌─────────────────┐
│  API Gateway    │ ← HTTP/gRPC + Auth Middleware
└────────┬────────┘
         │
    ┌────▼─────────────────┐
    │     Handlers         │ ← Rate limit, validation
    └────┬─────────────────┘
         │
    ┌────▼─────────────────┐
    │     Services         │ ← Business logic + MCP integration
    └────┬────────┬────────┘
         │        │
    ┌────▼────┐   │   ┌──────▼─────┐
    │Repository│   └──►│ Event Bus  │ ← NATS JetStream
    └────┬────┘       └──────┬─────┘
         │                   │
    ┌────▼────────┐    ┌────▼──────────┐
    │  PostgreSQL │    │  MCP Agents   │ ← Seed/Trainer/Evaluator/Reflector
    │  (RLS)      │    └───────────────┘
    └─────────────┘
```

**Fluxo de Dados**:
1. Request → API Gateway (auth + validation)
2. Handler → Service (business logic)
3. Service → Repository (persist) + Event Bus (publish event)
4. MCP Agent consome evento → processa → atualiza state
5. Observability stack captura métricas em todos os pontos

Documentação completa: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)

---

## 🛠️ Stack Tecnológica

| Componente | Tecnologia | Versão |
|------------|------------|--------|
| **Linguagem** | Go | 1.24+ |
| **Database** | PostgreSQL | 16+ |
| **Cache** | Redis | 7+ |
| **Message Bus** | NATS JetStream | 2.10+ |
| **Tracing** | Jaeger (OpenTelemetry) | Latest |
| **Métricas** | Prometheus | Latest |
| **Dashboards** | Grafana | Latest |
| **Containerização** | Docker + Docker Compose | Latest |
| **Orquestração** | Kubernetes (opcional) | 1.28+ |
| **Testing** | Testify + Testcontainers | Latest |
| **Mocks** | Go Mock | Latest |
| **Linting** | golangci-lint | 1.55+ |

---

## 📋 Pré-requisitos

### Obrigatórios

- **Go** ≥ 1.24.0 ([download](https://go.dev/dl/))
- **Docker** + **Docker Compose** ([download](https://docs.docker.com/get-docker/))
- **Git** ([download](https://git-scm.com/downloads))

### Opcionais (Recomendados)

- **Make** - Para automação de tasks
- **golangci-lint** - Para linting ([install](https://golangci-lint.run/usage/install/))
- **kubectl** - Para deploy em Kubernetes ([install](https://kubernetes.io/docs/tasks/tools/))

### Serviços Externos (Produção)

- Cluster PostgreSQL (ou RDS/Cloud SQL)
- Cluster Redis (ou ElastiCache/Memorystore)
- Cluster NATS (ou NATS Cloud)
- HashiCorp Vault (opcional, para secrets)

---

## ⚙️ Instalação

### Instalação via Docker (Recomendado)

A forma mais rápida de rodar o MCP-Ultra completo:

```bash
# 1. Clone o repositório
git clone https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git
cd mcp-ultra-wasm

# 2. Configure variáveis de ambiente
cp .env.example .env
# Edite .env com suas configurações
# IMPORTANTE: Gere secrets seguros (ver seção Configuração abaixo)

# 3. Inicie todos os serviços (app + postgres + redis + nats + observability)
docker-compose up -d

# 4. Verifique o health
curl http://localhost:9655/healthz

# 5. Acesse os serviços
# - API: http://localhost:9655
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3000 (admin/admin)
# - NATS Monitoring: http://localhost:8222
```

### Instalação Manual

Para desenvolvimento local sem Docker:

```bash
# 1. Clone o repositório
git clone https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git
cd mcp-ultra-wasm

# 2. Instale dependências
go mod download

# 3. Configure ambiente
cp .env.example .env
# Edite .env com suas configurações

# 4. Inicie dependências via Docker
docker-compose up -d postgres redis nats

# 5. Execute migrações do banco
make migrate-up
# ou manualmente:
# psql -h localhost -U postgres -d mcp_ultra_wasm -f migrations/*.sql

# 6. Build a aplicação
go build -o bin/mcp-ultra-wasm cmd/mcp-model-ultra/main.go

# 7. Execute
./bin/mcp-ultra-wasm

# Ou execute diretamente:
go run cmd/mcp-model-ultra/main.go
```

### Verificação da Instalação

```bash
# Health check básico
curl http://localhost:9655/healthz

# Health check detalhado (dependências)
curl http://localhost:9655/health/ready

# Métricas Prometheus
curl http://localhost:9655/metrics

# Resposta esperada do healthz:
# {"status":"healthy","service":"mcp-ultra-wasm","version":"1.0.0","timestamp":"..."}
```

---

## 🔧 Configuração

### Variáveis de Ambiente

O MCP-Ultra usa variáveis de ambiente para configuração. Copie `.env.example` para `.env`:

```bash
cp .env.example .env
```

**Variáveis Obrigatórias**:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=mcp_ultra_wasm_user
DB_PASSWORD=<gerar-senha-forte>  # openssl rand -base64 24
DB_NAME=mcp_ultra_wasm
DB_SSL_MODE=require

# NATS
NATS_URL=nats://localhost:4222
NATS_USERNAME=mcp_ultra_wasm
NATS_PASSWORD=<gerar-senha-forte>

# JWT
JWT_SECRET=<gerar-secret>  # openssl rand -base64 64
JWT_EXPIRATION=24h

# Encryption
ENCRYPTION_MASTER_KEY=<gerar-key>  # openssl rand -base64 32

# Server
SERVER_PORT=8080
LOG_LEVEL=info
```

**Variáveis Opcionais** (ver `.env.example` para lista completa):
- Vault integration
- OpenTelemetry configuration
- Rate limiting
- Circuit breaker
- Feature flags
- Compliance settings

### Secrets Management

**Desenvolvimento**:
```bash
# Usar .env file
SECRETS_BACKEND=env
```

**Produção** (recomendado):
```bash
# Usar HashiCorp Vault
SECRETS_BACKEND=vault
VAULT_ADDR=https://vault.example.com
VAULT_TOKEN=<vault-token>
VAULT_PATH=secret/mcp-ultra-wasm

# Ou Kubernetes Secrets
SECRETS_BACKEND=k8s
```

Documentação completa: [`docs/documentacao-full/CONFIGURACAO.md`](docs/documentacao-full/CONFIGURACAO.md)

---

## 🎮 Uso

### Iniciando a Aplicação

```bash
# Via Docker Compose (produção-like)
docker-compose up -d

# Via Go (desenvolvimento)
go run cmd/mcp-model-ultra/main.go

# Via binary compilado
./bin/mcp-ultra-wasm
```

### Endpoints Principais

**Health & Status**:
```bash
GET  /healthz                # Health check básico
GET  /health/ready           # Health check com dependências
GET  /metrics                # Métricas Prometheus
```

**API v1**:
```bash
# Autenticação
POST /api/v1/auth/login      # Login
POST /api/v1/auth/refresh    # Refresh token

# Recursos (requer auth)
GET    /api/v1/resources     # Listar recursos
POST   /api/v1/resources     # Criar recurso
GET    /api/v1/resources/:id # Obter recurso
PUT    /api/v1/resources/:id # Atualizar recurso
DELETE /api/v1/resources/:id # Deletar recurso
```

### Exemplo de Uso

```bash
# 1. Login
TOKEN=$(curl -X POST http://localhost:9655/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.token')

# 2. Criar recurso
curl -X POST http://localhost:9655/api/v1/resources \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Key: tenant-123" \
  -H "Content-Type: application/json" \
  -d '{"name":"My Resource","description":"Test"}'

# 3. Listar recursos
curl http://localhost:9655/api/v1/resources \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-Key: tenant-123"
```

Documentação completa da API: [`docs/documentacao-full/API.md`](docs/documentacao-full/API.md)

---

## 🤖 Agentes MCP

O MCP-Ultra implementa um sistema cognitivo baseado em **Model Context Protocol**:

### Tipos de Agentes

| Agente | Função | Frequência | NATS Subject |
|--------|--------|------------|--------------|
| **Seed** | Inicializa contexto e comportamento do tenant | No boot | `mcp.agent.seed.>` |
| **Trainer** | Aprendizado contínuo a partir de interações | A cada 15min | `mcp.agent.trainer.>` |
| **Evaluator** | Avalia qualidade e coerência das respostas | Contínuo | `mcp.agent.evaluator.>` |
| **Reflector** | Auto-análise e melhoria de comportamento | On-demand | `mcp.agent.reflector.>` |

### Ciclo de Vida

```
┌──────────┐    ┌──────────┐    ┌───────────┐    ┌───────────┐
│   Seed   │───►│ Trainer  │───►│ Evaluator │───►│ Reflector │
└──────────┘    └──────────┘    └───────────┘    └─────┬─────┘
     ▲                                                   │
     └───────────────────────────────────────────────────┘
                    (Melhoria Contínua)
```

### Eventos MCP

Schemas validados em `internal/schemas/`:
- `mcp.agent.seed.request` - Inicialização de novo tenant
- `mcp.agent.trainer.cycle` - Ciclo de treinamento
- `mcp.agent.evaluator.result` - Resultado de avaliação
- `mcp.agent.reflector.improvement` - Sugestão de melhoria

Documentação: [`docs/NATS_SUBJECTS.md`](docs/NATS_SUBJECTS.md)

---

## 🔌 API

### Autenticação

Todas as rotas protegidas requerem:

```bash
Authorization: Bearer <jwt-token>
X-Tenant-Key: <tenant-identifier>
```

### Rate Limiting

Por padrão:
- **Free**: 60 req/min
- **Pro**: 600 req/min
- **Enterprise**: 6000 req/min

Headers de resposta:
```
X-RateLimit-Limit: 600
X-RateLimit-Remaining: 599
X-RateLimit-Reset: 1234567890
```

### Paginação

```bash
GET /api/v1/resources?page=1&limit=50&sort=created_at&order=desc
```

### Respostas de Erro

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {"field": "email", "message": "Invalid email format"}
    ],
    "request_id": "req-abc123"
  }
}
```

Documentação completa: [`docs/documentacao-full/API.md`](docs/documentacao-full/API.md)

---

## 💻 Desenvolvimento

### Setup do Ambiente de Dev

```bash
# Instale ferramentas de desenvolvimento
make install-tools

# Ou manualmente:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install go.uber.org/mock/mockgen@latest
```

### Comandos Make

```bash
make lint              # Linting com golangci-lint
make test              # Rodar todos os testes
make coverage-html     # Gerar relatório de cobertura HTML
make mocks             # Regenerar mocks
make build             # Build da aplicação
make run               # Rodar aplicação
```

### Estrutura do Código

```
.
├── cmd/
│   └── mcp-model-ultra/        # Entry point da aplicação
├── internal/
│   ├── config/                 # Configuração e env vars
│   ├── handlers/               # HTTP/gRPC handlers
│   ├── services/               # Business logic
│   ├── repository/             # Data access layer
│   ├── models/                 # Domain entities
│   ├── middleware/             # Auth, logging, etc
│   ├── events/                 # NATS event handlers
│   └── schemas/                # JSON schemas para validação
├── pkg/                        # Bibliotecas reutilizáveis
│   ├── httpx/                  # HTTP utilities
│   ├── telemetry/              # Observability
│   └── security/               # Crypto, JWT, etc
├── test/                       # Testes de integração
│   └── mocks/                  # Mocks gerados
├── migrations/                 # SQL migrations
├── deploy/                     # Kubernetes manifests, Dockerfiles
└── docs/                       # Documentação
```

### Padrões de Código

- **Naming**: `camelCase` para unexported, `PascalCase` para exported
- **Errors**: Sempre retornar erros tipados com contexto
- **Logging**: Usar structured logging (zerolog/zap)
- **Testing**: Table-driven tests com testify
- **Comments**: Godoc em todas as funções públicas

---

## 🧪 Testes

### Executar Testes

```bash
# Todos os testes
make test

# Com verbosidade
go test ./... -v -count=1

# Apenas pacote específico
go test ./internal/services/... -v

# Com cobertura
make coverage-html
# Abre coverage.html no browser
```

### Tipos de Teste

**Unitários** (testify):
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

**Integração** (testcontainers):
```go
func TestDatabaseIntegration(t *testing.T) {
    ctx := context.Background()

    // Inicia PostgreSQL em container
    pgContainer, err := postgres.RunContainer(ctx, ...)
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)

    // Testes com banco real
    db := setupDB(pgContainer.ConnectionString())
    // ... testes
}
```

### Mocks

Regenerar mocks após alterar interfaces:

```bash
make mocks
```

### Cobertura de Código

Meta: **≥ 80%**

Verificar cobertura:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Documentação: [`docs/documentacao-full/TESTES.md`](docs/documentacao-full/TESTES.md)

---

## 🚀 Deployment

### Docker

```bash
# Build da imagem
docker build -t mcp-ultra-wasm:latest -f deploy/docker/Dockerfile .

# Run
docker run -p 8080:8080 \
  --env-file .env \
  mcp-ultra-wasm:latest
```

### Kubernetes

```bash
# Deploy completo (app + postgres + redis + nats)
kubectl apply -f deploy/k8s/

# Verificar status
kubectl get pods -n mcp-ultra-wasm

# Logs
kubectl logs -f deployment/mcp-ultra-wasm -n mcp-ultra-wasm

# Scaling
kubectl scale deployment/mcp-ultra-wasm --replicas=3 -n mcp-ultra-wasm
```

**Componentes Kubernetes**:
- `deployment.yaml` - Aplicação principal
- `service.yaml` - ClusterIP + LoadBalancer
- `configmap.yaml` - Configurações não-secretas
- `secret.yaml` - Credenciais (usar Sealed Secrets em prod)
- `ingress.yaml` - Roteamento HTTP/HTTPS
- `hpa.yaml` - Horizontal Pod Autoscaler

### CI/CD

Pipeline GitHub Actions (`.github/workflows/ci.yml`):

```yaml
- Build & Test
- Linting
- Security Scan (Grype)
- Docker Build
- Deploy to Staging
- Deploy to Production (manual approval)
```

Documentação: [`docs/documentacao-full/DEPLOY.md`](docs/documentacao-full/DEPLOY.md)

---

## 📊 Observabilidade

### Métricas (Prometheus)

Endpoint: `http://localhost:9655/metrics`

**Métricas Disponíveis**:
- `http_requests_total` - Total de requests por rota/status
- `http_request_duration_seconds` - Latência por percentil (p50, p95, p99)
- `mcp_agent_cycles_total` - Ciclos de agentes MCP
- `mcp_agent_errors_total` - Erros por agente
- `db_connections_active` - Conexões ativas com PostgreSQL
- `nats_messages_published_total` - Mensagens publicadas no NATS

**Configuração**:
```yaml
# deploy/monitoring/prometheus.yml
scrape_configs:
  - job_name: 'mcp-ultra-wasm'
    static_configs:
      - targets: ['mcp-ultra-wasm:9655']
```

### Dashboards (Grafana)

Acesse: `http://localhost:3000` (admin/admin)

**Dashboards Pré-configurados**:
- **Overview** - Métricas gerais (requests, latency, errors)
- **MCP Agents** - Ciclos, performance, health dos agentes
- **Database** - Queries, connections, slow queries
- **NATS** - Throughput, lag, consumer health

Import: `deploy/monitoring/grafana/dashboards/*.json`

### Tracing (Jaeger)

Acesse: `http://localhost:16686`

**Features**:
- Distributed tracing entre serviços
- Latency breakdown por operação
- Dependency graph
- Error tracking

### Logs

**Formato**: JSON estruturado

```json
{
  "level": "info",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "mcp-ultra-wasm",
  "trace_id": "abc123",
  "tenant_key": "tenant-123",
  "message": "Request processed",
  "duration_ms": 45,
  "status_code": 200
}
```

**Níveis**:
- `debug` - Desenvolvimento
- `info` - Produção (default)
- `warn` - Avisos
- `error` - Erros

Documentação: [`docs/documentacao-full/OBSERVABILIDADE.md`](docs/documentacao-full/OBSERVABILIDADE.md)

---

## 🔒 Segurança e Compliance

### Autenticação & Autorização

- **JWT** com RS256 ou HS256
- **Refresh tokens** com rotação
- **API Keys** para integrações M2M
- **RBAC** - Roles: admin, manager, analyst, user

### Proteção de Dados

- **Encryption at Rest** - AES-256 para dados sensíveis
- **Encryption in Transit** - TLS 1.2+ obrigatório
- **PII Scanning** - Detecta e marca dados pessoais
- **Data Retention** - Políticas configuráveis por tenant

### LGPD/GDPR

- ✅ **Consent Tracking** - Log de consentimentos
- ✅ **Data Portability** - Export completo de dados do usuário
- ✅ **Right to Erasure** - Deleção completa (hard delete)
- ✅ **Audit Trail** - Logs imutáveis de acesso a dados sensíveis

### Segurança da Aplicação

- Rate limiting por tenant e IP
- Circuit breaker para dependências externas
- Input validation (JSON schemas)
- SQL injection protection (parameterized queries)
- XSS protection (sanitização de output)
- CORS configurável

### Security Scanning

```bash
# Vulnerability scan com Grype
grype dir:. --config grype.yaml

# SAST com gosec
gosec ./...

# Dependency check
go list -json -m all | nancy sleuth
```

Documentação: [`docs/documentacao-full/SEGURANCA.md`](docs/documentacao-full/SEGURANCA.md)

---

## 🏢 Multi-tenancy

### Modelo de Isolamento

**Row Level Security (RLS)** no PostgreSQL:

```sql
-- Todas as tabelas têm tenant_key
CREATE TABLE resources (
    id UUID PRIMARY KEY,
    tenant_key VARCHAR(64) NOT NULL,
    name VARCHAR(255),
    ...
);

-- RLS ativado
ALTER TABLE resources ENABLE ROW LEVEL SECURITY;

-- Policy: usuário só vê dados do seu tenant
CREATE POLICY tenant_isolation ON resources
    FOR ALL
    TO application_user
    USING (tenant_key = current_setting('app.current_tenant')::VARCHAR);
```

### Propagação do Tenant

**HTTP**:
```bash
X-Tenant-Key: tenant-abc-123
```

**NATS Events**:
```json
{
  "tenant_key": "tenant-abc-123",
  "event_type": "resource.created",
  ...
}
```

**Database**:
```go
// Setter do tenant no contexto da sessão
db.Exec("SET app.current_tenant = $1", tenantKey)
```

### Limites por Tenant

Configurado em `business_config.yaml`:

```yaml
plans:
  - id: "pro"
    limits:
      agents: 25
      tenants: 5
      requests_per_minute: 600
```

---

## 💰 Planos e Billing

### Planos Disponíveis

| Plano | Preço | Agents | Tenants | Req/min | Features |
|-------|-------|--------|---------|---------|----------|
| **Free** | R$ 0 (30 dias) | 2 | 1 | 60 | Básico |
| **Pro** | R$ 299/mês | 25 | 5 | 600 | Completo + Observability |
| **Enterprise** | R$ 1.499/mês | 200 | 50 | 6000 | SLO 99.9% + Suporte 24/7 |

### KPIs de Negócio

Meta (configurado em `business_config.yaml`):

```yaml
kpis:
  mrr_target: 100000           # R$ 100k MRR
  cac_ltv_ratio_min: 4.0       # LTV:CAC ≥ 4:1
  payback_months_max: 3        # Payback ≤ 3 meses
  churn_monthly_max_pct: 2.5   # Churn ≤ 2.5%
```

### SLOs (Service Level Objectives)

| Métrica | Alvo | Plano |
|---------|------|-------|
| Latência p95 | ≤ 120 ms | Todos |
| Error rate | ≤ 0.5% | Todos |
| Uptime | ≥ 99.9% | Enterprise |
| Uptime | ≥ 99.5% | Pro |
| Cobertura testes | ≥ 80% | - |

---

## 🧭 Roadmap

### Q1 2025

- [ ] Implementar compliance v2 (`ScanForPII`, `RecordConsent`)
- [ ] Finalizar métricas de latência p95 por tenant
- [ ] Painel SaaS de billing integrado
- [ ] Multi-região (replicação cross-region)

### Q2 2025

- [ ] Suporte a webhooks configuráveis
- [ ] API GraphQL (além de REST)
- [ ] Mobile SDK (iOS/Android)
- [ ] Marketplace de agentes MCP customizados

### Q3 2025

- [ ] AI-powered analytics (insights automáticos)
- [ ] Self-service onboarding
- [ ] White-label customization
- [ ] Advanced RBAC com custom roles

### Futuro

- [ ] Suporte a outros bancos (MySQL, MongoDB)
- [ ] Edge computing (Cloudflare Workers)
- [ ] Blockchain audit trail (opcional)
- [ ] Real-time collaboration features

---

## 🤝 Contribuindo

**Este é um projeto open-source e contribuições são muito bem-vindas!** 🎉

Seja você desenvolvedor iniciante ou experiente, há várias formas de contribuir:
- 🐛 Reportar bugs via [Issues](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/issues)
- 💡 Sugerir novas features
- 📝 Melhorar a documentação
- 🔧 Submeter Pull Requests
- ⭐ Dar uma estrela no projeto

### Processo de Contribuição

1. **Fork** o repositório
2. **Clone** seu fork: `git clone https://github.com/seu-usuario/mcp-ultra-wasm.git`
3. **Crie branch**: `git checkout -b feature/minha-feature`
4. **Faça suas mudanças** e teste localmente
5. **Commit**: `git commit -m "feat: adiciona minha feature"`
6. **Push**: `git push origin feature/minha-feature`
7. **Pull Request** para `main` com descrição detalhada

### Convenções de Código

Seguimos [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: nova funcionalidade
fix: correção de bug
docs: documentação
style: formatação (sem mudança de lógica)
refactor: refatoração de código
test: adicionar ou corrigir testes
chore: tarefas de manutenção (deps, config, etc)
```

### Checklist do Pull Request

- ✅ Código compila sem erros (`go build ./...`)
- ✅ Testes passando (`make test`)
- ✅ Linting sem erros (`make lint`)
- ✅ Cobertura mantida ≥ 80%
- ✅ Documentação atualizada (README, godoc, etc)
- ✅ CHANGELOG atualizado (se aplicável)

### Primeiras Contribuições

Procurando por onde começar? Veja issues marcadas com:
- [`good first issue`](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/labels/good%20first%20issue) - Ideal para iniciantes
- [`help wanted`](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/labels/help%20wanted) - Precisamos de ajuda
- [`documentation`](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/labels/documentation) - Melhorias na documentação

### Código de Conduta

Este projeto segue o [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). Ao participar, você concorda em respeitar este código.

---

## 📜 Licença

**MIT License** - © 2025 Vertikon Labs

Este projeto é open-source e está disponível sob a licença MIT. Você é livre para:
- ✅ Usar comercialmente
- ✅ Modificar
- ✅ Distribuir
- ✅ Uso privado

Ver [`LICENSE`](LICENSE) para detalhes completos.

---

## 🆘 Suporte

### Documentação

- **Arquitetura**: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- **API**: [`docs/documentacao-full/API.md`](docs/documentacao-full/API.md)
- **Deploy**: [`docs/documentacao-full/DEPLOY.md`](docs/documentacao-full/DEPLOY.md)
- **Operação**: [`docs/documentacao-full/OPERACAO.md`](docs/documentacao-full/OPERACAO.md)
- **Catálogo de Erros**: [`docs/CATALOGO-ERROS-E-SOLUCOES.md`](docs/CATALOGO-ERROS-E-SOLUCOES.md)

### Comunidade e Suporte

- **Issues**: [GitHub Issues](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/issues) - Reporte bugs ou sugira features
- **Discussions**: [GitHub Discussions](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/discussions) - Tire dúvidas e compartilhe ideias
- **Pull Requests**: Contribuições são bem-vindas!
- **Email**: rogeriofesta@gmail.com (mantenedor principal)
- **Repositório**: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm

### Contribuidores

Este projeto é mantido pela comunidade. Agradecemos a todos que contribuem! 🙏

<a href="https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=vertikon/mcp-ultra-wasm" />
</a>

### FAQ

**P: Como faço para adicionar um novo agente MCP?**
R: Ver [`docs/documentacao-full/MANUAL-DE-USO.md`](docs/documentacao-full/MANUAL-DE-USO.md) seção "Agentes MCP".

**P: Como configurar multi-região?**
R: Ver [`docs/documentacao-full/DEPLOY.md`](docs/documentacao-full/DEPLOY.md) seção "Multi-região".

**P: Como funciona o billing?**
R: Ver `business_config.yaml` e integração com Vertikon Billing API.

---

### ✅ Status de Validação

| Item | Status |
|------|--------|
| Compilação | ✅ 100% |
| Testes | ✅ 100% |
| Linting | ✅ 100% |
| Cobertura (≥80%) | ✅ 100% |
| Security Scan | ✅ 100% |
| Documentação | ✅ 100% |
| **Score Total** | **✅ 20/20 (100%)** |

Template pronto para produção! 🚀

---

<div align="center">

**[⬆ Voltar ao topo](#-vertikon-mcp-ultra-wasm)**

Made with ❤️ by [Vertikon Labs](https://github.com/vertikon) and [Contributors](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/graphs/contributors)

⭐ **Se este projeto foi útil, considere dar uma estrela!** ⭐

</div>
