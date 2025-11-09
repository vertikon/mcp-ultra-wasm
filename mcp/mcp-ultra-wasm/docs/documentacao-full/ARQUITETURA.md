# 🏗️ Arquitetura - {{PROJECT_NAME}}

Documentação da arquitetura técnica do projeto **{{PROJECT_NAME}}**.

---

## 📌 Visão Geral
- **Linguagem**: {{LANGUAGE}} {{VERSION}}
- **Arquitetura**: Clean Architecture + Repository Pattern
- **Banco de Dados**: {{DATABASE}}
- **Cache**: {{CACHE_SYSTEM}}
- **Containerização**: Docker + Kubernetes
- **Observabilidade**: Prometheus, Grafana, Jaeger

---

## 🎯 Clean Architecture

```
┌─────────────────────────────────────┐
│           Presentation              │
│  (Controllers, Handlers, Routes)    │
├─────────────────────────────────────┤
│           Use Cases                 │
│    (Business Logic Layer)           │
├─────────────────────────────────────┤
│           Entities                  │
│      (Core Business Rules)          │
├─────────────────────────────────────┤
│        Infrastructure               │
│   (DB, External APIs, Cache)        │
└─────────────────────────────────────┘
```

---

## 📁 Estrutura de Pastas

```
{{PROJECT_NAME}}/
├── cmd/                    # Entrypoint da aplicação
│   └── main.go
├── internal/               # Código privado da aplicação
│   ├── controllers/        # Handlers HTTP
│   ├── usecases/          # Casos de uso (business logic)
│   ├── entities/          # Entidades de domínio
│   ├── repositories/      # Interfaces de repositório
│   └── infrastructure/    # Implementações concretas
│       ├── database/      # Conexões DB
│       ├── cache/         # Redis, etc.
│       └── external/      # APIs externas
├── pkg/                   # Código público reutilizável
│   ├── middleware/        # Middlewares HTTP
│   ├── utils/            # Utilitários
│   └── config/           # Configurações
├── deployments/          # Kubernetes, Docker configs
├── docs/                 # Documentação
└── tests/               # Testes integrados
```

---

## 🔄 Fluxo de Dados

```
HTTP Request
     ↓
[Controllers] → [Use Cases] → [Repositories] → [Database]
     ↑               ↓              ↑              ↓
[Response]  ← [Entities]  ← [Query Result] ← [SQL Query]
```

---

## 🗄️ Banco de Dados

### Principais Tabelas
- `{{table1}}` - {{Description}}
- `{{table2}}` - {{Description}}
- `{{table3}}` - {{Description}}

### Relacionamentos
```sql
{{table1}} (1) ←→ (N) {{table2}}
{{table2}} (1) ←→ (N) {{table3}}
```

---

## ⚡ Cache Strategy

### Redis Layers
- **L1**: Queries frequentes (TTL: 5min)
- **L2**: Dados de sessão (TTL: 30min)
- **L3**: Configurações (TTL: 1h)

---

## 🔐 Segurança

### Autenticação & Autorização
- **JWT RS256** tokens
- **RBAC** com roles: admin, manager, analyst, user
- **Middleware** de autenticação em todas as rotas protegidas

### Criptografia
- **TLS 1.3** obrigatório
- **AES-256** para dados sensíveis
- **bcrypt** para passwords

---

## 📊 Observabilidade

### Métricas (Prometheus)
- `http_requests_total`
- `http_request_duration_seconds`
- `database_connections_active`
- `{{business_metric}}_total`

### Tracing (Jaeger)
- Request tracing completo
- Latência por componente
- Análise de bottlenecks

### Logs (Structured JSON)
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "{{PROJECT_NAME}}",
  "trace_id": "abc123",
  "message": "Request processed",
  "duration_ms": 45
}
```

---

## 🚀 Escalabilidade

### Horizontal Scaling
- **Kubernetes HPA**: 3-20 pods
- **Load Balancing**: NGINX Ingress
- **Database**: Read replicas + Connection pooling

### Performance
- **Connection Pool**: Max 100 conexões
- **Cache Hit Ratio**: >95%
- **Response Time**: <200ms (P95)