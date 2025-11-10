# 🧠 Vertikon MCP-Ultra WASM

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/vertikon/mcp-ultra-wasm/actions)
[![Coverage](https://img.shields.io/badge/Coverage-%E2%89%A580%25-brightgreen)](docs/melhorias/ENHANCED_VALIDATION_REPORT.md)

### 🚀 WebAssembly Platform for Model Context Protocol (MCP)

O **MCP-Ultra WASM** é uma plataforma inovadora que combina **WebAssembly** com **Model Context Protocol** para criar aplicações web inteligentes com processamento de alto desempenho diretamente no navegador.

> 🎯 **Por que MCP-Ultra WASM?**  
> Execute código Go compilado como WebAssembly no navegador, com integração NATS real-time, segurança enterprise-grade, e observabilidade completa. A próxima geração de aplicações web inteligentes!

```bash
# Quick Start - 3 comandos para rodar tudo
git clone https://github.com/vertikon/mcp-ultra-wasm.git && cd mcp-ultra-wasm
cp mcp/mcp-ultra-wasm/.env.example mcp/mcp-ultra-wasm/.env && docker-compose -f mcp/mcp-ultra-wasm/docker-compose.yml up -d
curl http://localhost:8080/health  # ✅ Pronto!
```

---

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Características Principais](#-características-principais)
- [Arquitetura Web-WASM](#-arquitetura-web-wasm)
- [Stack Tecnológica](#-stack-tecnológica)
- [Pré-requisitos](#-pré-requisitos)
- [Instalação](#-instalação)
- [Uso](#-uso)
- [API](#-api)
- [Desenvolvimento](#-desenvolvimento)
- [Testes](#-testes)
- [Deployment](#-deployment)
- [SDK](#-sdk)
- [Roadmap](#-roadmap)
- [Contribuindo](#-contribuindo)

---

## 🚀 Visão Geral

O **MCP-Ultra WASM** é uma plataforma completa que une:

- **🌐 WebAssembly**: Código Go compilado para executar no navegador
- **📡 Model Context Protocol**: Comunicação inteligente entre agentes de IA
- **⚡ Real-time Processing**: WebSocket + NATS para comunicação instantânea
- **🔒 Enterprise Security**: JWT + RBAC + Rate Limiting
- **📊 Full Observability**: Prometheus + OpenTelemetry + Logging

### 🎯 Casos de Uso Ideais

- 🧠 **AI-powered Web Applications** - Processamento inteligente no cliente
- 📊 **Real-time Analytics** - Dashboards com WASM performance
- 🤖 **Smart Forms** - Validação e processamento inteligente
- 🔄 **Event-driven Interfaces** - UIs reativas a eventos MCP
- 🎯 **Interactive Code Editors** - Execução segura no navegador

### 💡 O que você ganha

- ✅ **Performance Nativa** - WASM roda ~90% da velocidade de código nativo
- ✅ **Segurança** - Sandbox do navegador + auth server-side
- ✅ **Real-time** - WebSocket + NATS para comunicação instantânea
- ✅ **Type Safety** - Go → WASM com verificação de tipos
- ✅ **Cross-platform** - Roda em qualquer browser moderno
- ✅ **Enterprise Ready** - Observabilidade, monitoring, security

---

## ✨ Características Principais

### 🎯 Core WebAssembly Features

- ✅ **Go → WASM Compilation** - Build automático de código Go para WebAssembly
- ✅ **Browser Execution** - Execução segura e performática no cliente
- ✅ **JS Interop** - Comunicação bidirecional entre JavaScript e Go WASM
- ✅ **Memory Management** - Gerenciamento automático de memória no WASM
- ✅ **Module Loading** - Carregamento dinâmico de módulos WASM

### 🤖 MCP Integration

- **Smart Task Processing** - Agentes MCP processando tarefas no navegador
- **Context Sharing** - Compartilhamento de contexto entre frontend e backend
- **Event-driven Updates** - Atualizações automáticas via NATS + WebSocket
- **Intelligent Caching** - Cache inteligente de resultados WASM

### 📡 Real-time Communication

- **WebSocket Handlers** - Comunicação bidirecional server ↔ browser
- **NATS JetStream** - Messaging system enterprise-grade
- **Event Streaming** - Stream de eventos em tempo real
- **Connection Management** - Pooling e reconexão automática

### 🔒 Security & Performance

- **JWT Authentication** - Auth server-side com tokens JWT
- **RBAC Authorization** - Controle de acesso granular
- **Rate Limiting** - Proteção contra滥用
- **CORS Configuration** - Cross-origin seguro
- **Input Validation** - Validação rigorosa de dados

### 📊 Observability

- **Prometheus Metrics** - Métricas de performance do WASM
- **OpenTelemetry Tracing** - Distributed tracing end-to-end
- **Structured Logging** - Logs JSON com contexto completo
- **Health Monitoring** - Health checks em tempo real

---

## 🏗️ Arquitetura Web-WASM

```
┌─────────────────────────────────────────────────────┐
│                 Browser Frontend                    │
├─────────────────────┬───────────────────────────────┤
│   WebAssembly (Go)  │     JavaScript Client         │
│   ┌─────────────┐   │   ┌─────────────────────┐      │
│   │ Functions   │◄──┼──►│ WebSocket Client    │      │
│   │ Analysis    │   │   │ Event Handlers      │      │
│   │ Generation  │   │   │ UI Updates          │      │
│   │ Validation  │   │   └─────────────────────┘      │
│   └─────────────┘   │              │                  │
└─────────────────────┴──────────────┼──────────────────┘
                           ▼          │
                    ┌─────────────┐   │
                    │   WebSocket │◄──┘
                    │   Server    │
                    └─────────────┘
                           │
                    ┌─────────────┐
                    │  Go Server  │
                    │  (Gin)      │
                    └─────────────┘
                           │
                    ┌─────────────┐
                    │  NATS Jet   │
                    │  Stream     │
                    └─────────────┘
                           │
                    ┌─────────────┐
                    │  MCP Agents │
                    │  & Events   │
                    └─────────────┘
```

**Fluxo de Dados Web-WASM**:
1. **Browser** carrega módulo WASM compilado de Go
2. **JavaScript** invoca funções WASM via JS interop
3. **WASM** processa dados localmente (alta performance)
4. **WebSocket** envia eventos para o servidor Go
5. **NATS** distribui eventos para MCP agents
6. **Real-time updates** voltam via WebSocket

---

## 🛠️ Stack Tecnológica

| Camada | Tecnologia | Versão | Status |
|--------|------------|--------|--------|
| **Frontend** | HTML5 + CSS3 + JavaScript | Latest | ✅ Completo |
| **WASM Runtime** | Go → WebAssembly | 1.24+ | ✅ Compilando |
| **WebSocket** | Gorilla WebSocket | Latest | ✅ Real-time |
| **Servidor** | Go + Gin | 1.24+ | ✅ Production |
| **Messaging** | NATS JetStream | 2.10+ | ✅ Event-driven |
| **Auth** | JWT + RBAC | Latest | ✅ Secure |
| **Metrics** | Prometheus + OpenTelemetry | Latest | ✅ Monitoring |
| **Container** | Docker + K8s | Latest | ✅ Deploy-ready |

---

## 📋 Pré-requisitos

### Obrigatórios

- **Go** ≥ 1.24.0 ([download](https://go.dev/dl/))
- **Docker** + **Docker Compose** ([download](https://docs.docker.com/get-docker/))
- **Git** ([download](https://git-scm.com/downloads))

### Opcionais (Recomendados)

- **Node.js** ≥ 18 (para desenvolvimento frontend)
- **Make** - Para automação de tasks
- **kubectl** - Para deploy em Kubernetes

---

## ⚙️ Instalação

### 1. Clone o Repositório

```bash
git clone https://github.com/vertikon/mcp-ultra-wasm.git
cd mcp-ultra-wasm
```

### 2. Setup do Ambiente

```bash
# Entrar no diretório do projeto principal
cd mcp/mcp-ultra-wasm

# Configurar variáveis de ambiente
cp .env.example .env
# Edite .env com suas configurações

# Instalar dependências Go
go mod download
```

### 3. Compilar Módulo WASM

```bash
# Compilar Go para WebAssembly
$env:GOOS="js"; $env:GOARCH="wasm"; go build -o ../web-wasm/wasm/main.wasm ../web-wasm/wasm/main.go

# Verificar se foi criado
Test-Path "../web-wasm/wasm/main.wasm"  # Deve retornar True
```

### 4. Iniciar Serviços

```bash
# Via Docker (recomendado para produção)
docker-compose -f docker-compose.yml up -d

# Ou via Go (desenvolvimento)
go run ./cmd/web-wasm-server
```

### 5. Verificar Instalação

```bash
# Health check
curl http://localhost:8080/health

# Resposta esperada:
# {"status":"ok","timestamp":"2025-01-15T10:30:00Z","service":"web-wasm-server","version":"1.0.0"}

# Acessar interface web
open http://localhost:8080
```

---

## 🎮 Uso

### Interface Web

Acesse **http://localhost:8080** para usar a interface completa com:

- 📊 **Dashboard** com métricas em tempo real
- 🧠 **WASM Task Runner** para executar análises
- 📡 **WebSocket Monitor** para ver eventos em tempo real
- 🔧 **Configuration Panel** para ajustar parâmetros

### API REST

```bash
# Criar nova task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "analyze_project",
    "config": {
      "project_path": "/path/to/project",
      "analysis_type": "security"
    }
  }'

# Listar tasks
curl http://localhost:8080/api/v1/tasks

# Obter task específica
curl http://localhost:8080/api/v1/tasks/{task_id}

# Cancelar task
curl -X DELETE http://localhost:8080/api/v1/tasks/{task_id}
```

### WebSocket

```javascript
// Conectar ao WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

// Enviar comando para executar função WASM
ws.send(JSON.stringify({
  type: 'execute_wasm',
  data: {
    function: 'analyzeProject',
    config: { projectPath: '/my/project' }
  }
}));

// Receber resultados
ws.onmessage = (event) => {
  const result = JSON.parse(event.data);
  console.log('WASM Result:', result);
};
```

### Módulo WASM no Browser

```javascript
// Carregar módulo WASM
import { loadWasm } from './js/wasm-loader.js';

// Inicializar
const wasmModule = await loadWasm();

// Executar funções Go no navegador
const analysisResult = wasmModule.analyzeProject({
  projectPath: './my-project',
  includeTests: true
});

const generatedCode = wasmModule.generateCode({
  language: 'go',
  pattern: 'crud-api'
});

const validationResult = wasmModule.validateConfig({
  configFile: './app.yaml',
  schema: 'v2'
});
```

---

## 🔌 API Reference

### Endpoints

#### Health & Status
```
GET  /health              # Health check básico
GET  /api/v1/tasks        # Listar tasks
POST /api/v1/tasks        # Criar task
GET  /api/v1/tasks/:id    # Obter task
DELETE /api/v1/tasks/:id  # Cancelar task
GET  /ws                  # WebSocket endpoint
```

#### Respostas

**Sucesso - Task Creation**:
```json
{
  "id": "task-abc123",
  "name": "analyze_project",
  "status": "pending",
  "config": {...},
  "created_at": "2025-01-15T10:30:00Z"
}
```

**Erro - Validação**:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": [
      {
        "field": "project_path",
        "message": "Path is required"
      }
    ]
  }
}
```

---

## 💻 Desenvolvimento

### Estrutura do Projeto

```
mcp/mcp-ultra-wasm/
├── cmd/
│   └── web-wasm-server/         # Servidor principal
├── internal/web-wasm/
│   ├── handlers/                # HTTP & WebSocket handlers
│   │   ├── api_handler.go      # API REST endpoints
│   │   ├── websocket_handler.go # WebSocket real-time
│   │   └── ui_handler.go       # Static files
│   ├── nats/                    # NATS integration
│   │   ├── client.go           # NATS client
│   │   └── publisher.go        # Event publishing
│   ├── observability/          # Monitoring
│   │   ├── logger.go           # Structured logging
│   │   ├── metrics.go          # Prometheus metrics
│   │   └── tracing.go          # OpenTelemetry
│   ├── security/               # Security middleware
│   │   ├── auth.go             # JWT authentication
│   │   ├── cors.go             # CORS handling
│   │   └── ratelimit.go        # Rate limiting
│   └── sdk/                    # SDK integration
│       ├── client.go           # MCP client
│       ├── registry.go         # Plugin registry
│       └── contracts.go        # Type definitions
├── web-wasm/
│   ├── wasm/                    # Módulo WebAssembly
│   │   ├── main.go             # Go code para WASM
│   │   ├── functions/          # Funções exportadas
│   │   └── internal/          # Lógica interna WASM
│   └── static/                  # Frontend assets
│       ├── index.html          # Main interface
│       ├── css/main.css        # Estilos
│       └── js/                 # JavaScript
│           ├── main.js         # Application logic
│           ├── wasm-loader.js  # WASM loader
│           └── websocket-client.js # WebSocket client
├── deploy/
│   ├── docker/web-wasm/        # Docker configs
│   └── k8s/web-wasm/           # Kubernetes manifests
└── test/web-wasm/              # Testes
```

### Comandos de Desenvolvimento

```bash
# Compilar WASM
$env:GOOS="js"; $env:GOARCH="wasm"; go build -o ../web-wasm/wasm/main.wasm ../web-wasm/wasm/main.go

# Rodar servidor em modo dev
go run ./cmd/web-wasm-server -log-level=debug

# Rodar testes
go test ./...

# Formatar código
gofmt -w .

# Linting
golangci-lint run

# Build para produção
go build -o bin/web-wasm-server ./cmd/web-wasm-server
```

### Hot Reload no Desenvolvimento

Para desenvolvimento com hot reload:

```bash
# Instalar air para hot reload
go install github.com/cosmtrek/air@latest

# Criar .air.toml
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/web-wasm-server"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
EOF

# Rodar com hot reload
air
```

---

## 🧪 Testes

### Tipos de Testes

```bash
# Testes unitários
go test ./internal/web-wasm/handlers/...

# Testes de integração
go test ./test/web-wasm/...

# Testes WASM
go test ./web-wasm/wasm/...

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Testes de integração com Docker
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Testes de WebSocket

```go
func TestWebSocketConnection(t *testing.T) {
    // Test server
    server := httptest.NewServer(setupRouter())
    defer server.Close()
    
    // Convert HTTP to WebSocket
    wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
    
    // Connect
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    require.NoError(t, err)
    defer conn.Close()
    
    // Send message
    err = conn.WriteMessage(websocket.TextMessage, []byte("test"))
    require.NoError(t, err)
    
    // Read response
    _, message, err := conn.ReadMessage()
    require.NoError(t, err)
    assert.Contains(t, string(message), "response")
}
```

---

## 🚀 Deployment

### Docker

```bash
# Build imagem
docker build -f deploy/docker/web-wasm/Dockerfile -t mcp-ultra-wasm:latest .

# Rodar container
docker run -p 8080:8080 \
  --env-file .env \
  mcp-ultra-wasm:latest
```

### Docker Compose

```bash
# Subir stack completo
docker-compose -f docker-compose.yml up -d

# Verificar status
docker-compose ps

# Logs
docker-compose logs -f web-wasm-server

# Parar
docker-compose down
```

### Kubernetes

```bash
# Deploy completo
kubectl apply -f deploy/k8s/web-wasm/

# Verificar deployment
kubectl get pods -l app=web-wasm-server

# Acessar serviço
kubectl port-forward svc/web-wasm-service 8080:80
```

### Variáveis de Ambiente Produção

```bash
# Server
SERVER_PORT=8080
LOG_LEVEL=info

# NATS
NATS_URL=nats://nats:4222
NATS_USERNAME=mcp_ultra_wasm
NATS_PASSWORD=secure_password

# Security
JWT_SECRET=your_jwt_secret_here
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Observability
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

---

## 🔌 SDK

O SDK MCP Ultra WASM permite integração com outras aplicações:

### Instalação

```bash
go get github.com/vertikon/mcp-ultra-wasm/sdk/sdk-ultra-wasm@latest
```

### Uso Básico

```go
package main

import (
    "context"
    "log"
    
    "github.com/vertikon/mcp-ultra-wasm/sdk/sdk-ultra-wasm"
)

func main() {
    // Criar cliente
    client, err := sdk.NewClient(&sdk.Config{
        ServerURL: "http://localhost:8080",
        APIKey:    "your-api-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Executar análise WASM
    result, err := client.ExecuteWasmFunction(context.Background(), &sdk.WasmRequest{
        Function: "analyzeProject",
        Config: map[string]interface{}{
            "projectPath": "/path/to/project",
            "includeTests": true,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Analysis result: %+v", result)
}
```

### Features do SDK

- ✅ **Client SDK** - Integração fácil com Go
- ✅ **Type Safety** - Interfaces tipadas
- ✅ **Context Support** - Timeout e cancelamento
- ✅ **Error Handling** - Erros detalhados
- ✅ **Async Support** - Suporte a operações assíncronas

---

## 📊 Monitoramento

### Métricas Prometheus

Acesse: `http://localhost:8080/metrics`

**Métricas Disponíveis**:
- `wasm_functions_total` - Total de execuções WASM
- `wasm_execution_duration_seconds` - Tempo de execução
- `websocket_connections_active` - Conexões ativas
- `nats_messages_published_total` - Mensagens NATS
- `http_requests_total` - Requests HTTP por endpoint

### Health Checks

```bash
# Health básico
curl http://localhost:8080/health

# Ready check (com dependências)
curl http://localhost:8080/ready

# Metrics endpoint
curl http://localhost:8080/metrics
```

### Logging

Logs estruturados em formato JSON:

```json
{
  "level": "info",
  "timestamp": "2025-01-15T10:30:00Z",
  "service": "web-wasm-server",
  "trace_id": "abc123",
  "message": "WASM function executed",
  "function": "analyzeProject",
  "duration_ms": 150,
  "status": "success"
}
```

---

## 🧭 Roadmap

### Q1 2025

- [x] ✅ **WebAssembly Core** - Go → WASM compilation
- [x] ✅ **Real-time Communication** - WebSocket + NATS
- [x] ✅ **Security Layer** - JWT + RBAC
- [x] ✅ **Observability** - Prometheus + OpenTelemetry
- [ ] **Advanced WASM Features** - Shared memory, multi-threading
- [ ] **Plugin System** - Dynamic WASM plugin loading

### Q2 2025

- [ ] **WASM Optimization** - Code splitting, lazy loading
- [ ] **Advanced MCP Integration** - Multi-agent orchestration
- [ ] **Performance Monitoring** - APM específico para WASM
- [ ] **Security Hardening** - WASM sandboxing avançado

### Future

- [ ] **Mobile Support** - React Native + WASM
- [ ] **Edge Computing** - Cloudflare Workers + WASM
- [ ] **AI-powered Optimization** - Auto-tuning WASM performance
- [ ] **Marketplace** - WASM module marketplace

---

## 🤝 Contribuindo

Contribuições são bem-vindas! 🎉

### Como Contribuir

1. **Fork** o repositório
2. **Clone** seu fork
3. **Crie branch**: `git checkout -b feature/nova-feature`
4. **Faça mudanças** e teste
5. **Commit**: `git commit -m "feat: adiciona nova feature"`
6. **Push**: `git push origin feature/nova-feature`
7. **Pull Request** com descrição detalhada

### Áreas de Contribuição

- 🧠 **Novas funções WASM** - Expandir capacidades
- 🎨 **UI/UX** - Melhorar interface web
- 📊 **Monitoring** - Novas métricas e dashboards
- 🧪 **Testes** - Aumentar cobertura
- 📚 **Documentação** - Melhorar docs

### Convenções

Seguimos [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: nova funcionalidade
fix: correção de bug
docs: documentação
test: testes
refactor: refatoração
perf: performance
chore: manutenção
```

---

## 📜 Licença

**MIT License** - © 2025 Vertikon Labs

Este projeto é open-source e disponível sob a licença MIT.

---

## 🆘 Suporte

### Documentação

- 📖 **Quick Start**: [`QUICK_START.md`](QUICK_START.md)
- 🏗️ **Arquitetura**: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
- 🔧 **API Reference**: [`docs/API.md`](docs/API.md)
- 🐳 **Deployment**: [`docs/DEPLOY.md`](docs/DEPLOY.md)

### Comunidade

- 🐛 **Issues**: [Reportar bugs](https://github.com/vertikon/mcp-ultra-wasm/issues)
- 💡 **Discussions**: [Tirar dúvidas](https://github.com/vertikon/mcp-ultra-wasm/discussions)
- 📧 **Email**: rogeriofesta@gmail.com

### Status do Projeto

| Item | Status |
|------|--------|
| Build | ✅ Passing |
| Tests | ✅ 85%+ Coverage |
| Lint | ✅ 0 Issues |
| Security | ✅ Scanned |
| Documentation | ✅ Complete |

---

<div align="center">

**🚀 MCP Ultra WASM - A próxima geração de aplicações web inteligentes!**

Made with ❤️ by [Vertikon Labs](https://github.com/vertikon)

⭐ **Se este projeto foi útil, considere dar uma estrela!** ⭐

</div>