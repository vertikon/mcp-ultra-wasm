# MCP Ultra SDK - Custom Extension Framework

[![CI](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/actions/workflows/ci.yml/badge.svg)](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm.svg)](https://pkg.go.dev/github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm)
[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/vertikon/sdk-ultra-wasm)](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/releases)
[![Ultra Verified](https://img.shields.io/badge/Ultra%20Verified-100%25-success)](docs/CERTIFICADO_VALIDACAO_V9.md)

**Versão:** 9.0.0
**Status:** ✅ ULTRA VERIFIED CERTIFIED
**Licença:** MIT

SDK de customização para o ecossistema **mcp-ultra-wasm**, permitindo estender funcionalidades através de plugins sem modificar o template original.

---

## 🎯 Visão Geral

O **sdk-ultra-wasm** é um framework de extensão que permite:

- ✅ **Core Imutável** - Template original permanece intocado
- ✅ **Extensões Isoladas** - Plugins customizados em camada separada
- ✅ **Contratos Estáveis** - Interfaces versionadas com SemVer
- ✅ **Auto-registro** - Plugins registrados automaticamente via `init()`
- ✅ **Type-safe** - Registry com tipos segregados
- ✅ **Pronto para produção** - Health checks, middlewares, policies

---

## 📦 Estrutura

```
sdk-ultra-wasm/
├── pkg/
│   ├── contracts/          # Extension points (v1.0.0)
│   │   ├── route.go        # RouteInjector interface
│   │   ├── middleware.go   # MiddlewareInjector interface
│   │   ├── job.go          # Job interface
│   │   ├── service.go      # Service interface
│   │   └── version.go      # SemVer compatibility
│   ├── registry/           # Plugin registry
│   │   └── registry.go     # Type-safe registration
│   ├── router/             # HTTP abstractions
│   │   ├── mux.go          # Gorilla Mux wrapper
│   │   └── middleware/     # Built-in middlewares
│   ├── policies/           # Auth & RBAC
│   │   ├── jwt.go          # JWT authentication
│   │   ├── rbac.go         # Role-based access control
│   │   └── context.go      # Identity context
│   └── bootstrap/          # SDK initialization
│       ├── bootstrap.go    # Main bootstrap
│       └── health.go       # Health endpoints
├── cmd/
│   └── ultra-sdk-cli/      # CLI scaffolding tool
└── seed-examples/
    └── waba/               # WhatsApp Business API example
```

---

## 🚀 Quick Start

### 📦 Instalação

**Pré-requisitos:**
- Go 1.21 ou superior
- Git

**Instalar o SDK:**

```bash
go get github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm@v9.0.0
```

### ✅ Validação da Instalação

Após instalar o SDK, valide a configuração:

```bash
# 1. Verificar dependências
go mod download
go mod verify

# 2. Executar testes
go test ./pkg/... -v -cover

# 3. Compilar projeto
go build ./cmd/...

# 4. Verificar linter (recomendado)
golangci-lint run ./...

# 5. Verificar formatação
gofmt -l .
```

**Saída esperada:**
- ✅ Todos os testes passando (PASS)
- ✅ Coverage >= 70%
- ✅ Zero erros de compilação
- ✅ Zero warnings do linter

### 2. Criar um Novo Plugin

Use a CLI para gerar scaffold:

```bash
# Compilar CLI
cd cmd/ultra-sdk-cli
go build -o ../../bin/ultra-sdk-cli.exe

# Gerar plugin
./bin/ultra-sdk-cli.exe --name marketing --kind marketing
```

Ou crie manualmente:

```go
// internal/plugins/marketing/plugin.go
package marketing

import (
    "encoding/json"
    "net/http"

    "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
    "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
)

func init() {
    _ = registry.Register("marketing", &Plugin{})
}

type Plugin struct{}

func (p *Plugin) Name() string    { return "marketing" }
func (p *Plugin) Version() string { return "1.0.0" }

func (p *Plugin) Routes() []contracts.Route {
    return []contracts.Route{
        {
            Method:  "POST",
            Path:    "/marketing/campaign",
            Handler: p.createCampaign,
        },
    }
}

func (p *Plugin) createCampaign(w http.ResponseWriter, r *http.Request) {
    // Implementação
    json.NewEncoder(w).Encode(map[string]string{
        "status": "created",
    })
}
```

## 🚦 Execução

### Inicializar no main.go

```go
// cmd/main.go
package main

import (
    "log"
    "net/http"

    "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/bootstrap"

    // Import side-effect para auto-registro
    _ "github.com/vertikon/seed/internal/plugins/marketing"
    _ "github.com/vertikon/seed/internal/plugins/waba"
)

func main() {
    // Bootstrap SDK
    mux := bootstrap.Bootstrap(bootstrap.Config{
        EnableRecovery: true,
        EnableLogger:   true,
        CORSOrigins:    []string{"*"},
    })

    // Servidor HTTP
    log.Println("🚀 Servidor iniciando na porta 8080")
    http.ListenAndServe(":8080", mux)
}
```

### Executar Servidor

```bash
# Desenvolvimento
go run ./cmd/main.go

# Produção (compilado)
go build -o bin/server ./cmd/main.go
./bin/server
```

### Validar Servidor em Execução

```bash
# Health check
curl http://localhost:8080/healthz
# Esperado: 200 OK

# Readiness
curl http://localhost:8080/readyz
# Esperado: 200 OK (ou 503 se não pronto)

# Listar rotas (debug)
curl http://localhost:8080/debug/routes  # Se habilitado
```

---

## 📋 Extension Points (Contratos v1.0.0)

### 1. RouteInjector

Permite plugins registrarem rotas HTTP:

```go
type RouteInjector interface {
    Name() string
    Version() string
    Routes() []Route
}
```

### 2. MiddlewareInjector

Permite plugins registrarem middlewares com prioridade:

```go
type MiddlewareInjector interface {
    Name() string
    Priority() int  // Menor = primeiro
    Middleware() func(http.Handler) http.Handler
}
```

### 3. Job

Permite plugins registrarem jobs agendados:

```go
type Job interface {
    Name() string
    Schedule() string  // Expressão cron
    Run(ctx context.Context) error
}
```

### 4. Service

Permite plugins registrarem serviços customizados:

```go
type Service interface {
    Name() string
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() error
}
```

---

## 🔐 Policies (JWT + RBAC)

### Autenticação JWT

```go
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/policies"

// Implementar TokenValidator
type MyValidator struct{}

func (v *MyValidator) Validate(token string) (subject string, roles []string, err error) {
    // Validar JWT e extrair claims
    return "user123", []string{"admin", "editor"}, nil
}

// Aplicar middleware
mux.Use(policies.Auth(&MyValidator{}))
```

### Controle de Acesso (RBAC)

```go
// Exigir papel específico
protectedHandler := policies.RequireRole("admin")(myHandler)

// Exigir qualquer um dos papéis
protectedHandler := policies.RequireAnyRole("admin", "editor")(myHandler)
```

### Acessar Identidade

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    identity := policies.FromIdentity(r.Context())
    if identity != nil {
        log.Printf("User: %s, Roles: %v", identity.Subject, identity.Roles)
    }
}
```

---

## ⚙️ Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão | Obrigatório |
|----------|-----------|--------|-------------|
| `PORT` | Porta HTTP do servidor | `8080` | Não |
| `LOG_LEVEL` | Nível de log (debug\|info\|warn\|error) | `info` | Não |
| `GOMEMLIMIT` | Limite de memória Go (alinhado ao pod limit) | - | Recomendado |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endpoint OpenTelemetry | - | Não |
| `OTEL_RESOURCE_ATTRIBUTES` | Atributos do recurso OTEL | - | Não |
| `NATS_URL` | URL do servidor NATS | `nats://localhost:4222` | Não |
| `NATS_CLUSTER_ID` | ID do cluster NATS | `mcp-cluster` | Não |
| `JWT_SECRET` | Secret para validação JWT | - | Se usar auth |

### Exemplo .env

```bash
# Server
PORT=8080
LOG_LEVEL=info

# Observability
GOMEMLIMIT=512MiB
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_RESOURCE_ATTRIBUTES=service.name=sdk-ultra-wasm,service.version=9.0.0

# Messaging
NATS_URL=nats://nats:4222
NATS_CLUSTER_ID=mcp-prod

# Security
JWT_SECRET=your-secret-key-here
```

---

## 🏥 Health Endpoints

Automaticamente disponíveis após `bootstrap.Bootstrap()`:

- `GET /healthz` - Liveness probe (sempre retorna 200)
- `GET /readyz` - Readiness probe (503 se não pronto)
- `GET /health` - Alias de `/healthz`
- `GET /ping` - Alias de `/healthz`
- `GET /metrics` - Métricas Prometheus (se habilitado)

### Controlar Readiness

```go
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/bootstrap"

// Marcar como pronto
bootstrap.MarkReady()

// Marcar como não-pronto
bootstrap.MarkNotReady()
```

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## 🧪 Testes

### Executar Testes

```bash
go test ./pkg/... -v
```

### Exemplo de Teste

```go
func TestMyPlugin(t *testing.T) {
    registry.Reset()

    plugin := &MyPlugin{}
    err := registry.Register("test", plugin)

    if err != nil {
        t.Fatalf("Erro ao registrar: %v", err)
    }

    injectors := registry.RouteInjectors()
    if len(injectors) != 1 {
        t.Errorf("Esperava 1 plugin, obteve %d", len(injectors))
    }
}
```

---

## 📊 Exemplo Completo: WABA Plugin

Veja `seed-examples/waba/` para exemplo completo de plugin WhatsApp Business API com:

- ✅ Verificação de webhook (GET /waba/webhook)
- ✅ Recebimento de mensagens (POST /waba/webhook com HMAC)
- ✅ Envio de templates (POST /waba/send)
- ✅ Listagem de templates (GET /waba/templates)

### Executar Exemplo WABA

```bash
cd seed-examples/waba

# Configurar variáveis de ambiente
cp .env.example .env
# Editar .env com suas credenciais

# Rodar servidor
go run ./cmd/main.go
```

### Testar Endpoints

```bash
# Health check
curl http://localhost:8080/healthz

# Verificar webhook (Meta)
curl "http://localhost:8080/waba/webhook?hub.mode=subscribe&hub.verify_token=SEU_TOKEN&hub.challenge=123"

# Enviar template
curl -X POST http://localhost:8080/waba/send \
  -H "Content-Type: application/json" \
  -d '{"to":"5511999999999","template":"welcome","params":["João"]}'
```

---

## 🔧 Middlewares Built-in

### Recovery

Captura panics e retorna 500:

```go
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/router/middleware"

mux.Use(middleware.Recovery())
```

### Logger

Registra todas as requests:

```go
mux.Use(middleware.Logger())
```

### CORS

Adiciona headers CORS:

```go
mux.Use(middleware.CORS([]string{"*"}))
// ou
mux.Use(middleware.CORS([]string{"https://example.com"}))
```

---

## 📚 SemVer & Compatibilidade

O SDK segue [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.x.x) - Mudanças incompatíveis nas interfaces
- **MINOR** (x.1.x) - Novas funcionalidades compatíveis
- **PATCH** (x.x.1) - Correções de bugs compatíveis

### Verificar Compatibilidade

```go
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"

if !contracts.CompatibleWith("1.2.0") {
    log.Fatal("Plugin incompatível com SDK")
}
```

---

## 🛠️ CLI de Scaffolding

### Compilar CLI

```bash
cd cmd/ultra-sdk-cli
go build -o ../../bin/ultra-sdk-cli.exe
```

### Uso

```bash
# Gerar plugin genérico
./bin/ultra-sdk-cli.exe --name my-plugin

# Gerar plugin específico
./bin/ultra-sdk-cli.exe --name campaigns --kind marketing

# Especificar diretório de saída
./bin/ultra-sdk-cli.exe --name payments --output custom/path
```

---

## 📋 Checklist de Produção

Antes de fazer deploy:

- [ ] Todos os testes passando (`go test ./...`)
- [ ] Plugin registrado via `init()`
- [ ] Versão SemVer definida
- [ ] Health checks respondendo
- [ ] Logs estruturados configurados
- [ ] CORS configurado corretamente
- [ ] Secrets em variáveis de ambiente (não hardcoded)
- [ ] Graceful shutdown implementado
- [ ] Métricas expostas (se aplicável)

---

## 🤝 Contribuindo

1. Fork o repositório
2. Crie uma branch de feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'feat: adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

---

## 📝 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---

## 🆘 Suporte

- 📧 Email: dev@vertikon.com
- 📚 Docs: https://docs.vertikon.com/mcp-ultra-wasm-sdk
- 🐛 Issues: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/issues

---

## 🔌 Integração NATS

O SDK suporta comunicação via **NATS** para mensageria assíncrona entre plugins e serviços MCP.

**Documentação completa:** [docs/NATS_SUBJECTS.md](docs/NATS_SUBJECTS.md)

**Subjects documentados:**
- `mcp.ultra.sdk.custom.health.ping` - Health check via NATS
- `mcp.ultra.sdk.custom.seed.validate` - Validação de seeds
- `mcp.ultra.sdk.custom.template.sync` - Sincronização de templates
- `mcp.ultra.sdk.custom.sdk.check` - Verificação de compatibilidade

**Configuração:**
```bash
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=mcp-cluster
NATS_CLIENT_ID=sdk-ultra-wasm
```

---

## 🎯 Roadmap

- [ ] Adapter Meta Graph API (WABA completo)
- [ ] Observability (OpenTelemetry)
- [ ] Job Scheduler (robfig/cron)
- [ ] Plugin Marketplace
- [ ] Hot Reload de Plugins
- [ ] CLI flags avançados (--with-auth, --with-jobs)
- [ ] NATS Streaming completo

---

**Desenvolvido com ❤️ pela equipe Vertikon**
