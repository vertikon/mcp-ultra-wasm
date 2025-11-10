# 🚀 Quick Start - wasm Platform

## 3 comandos para rodar tudo

```bash
# 1. Clone o repositório
git clone https://github.com/vertikon/mcp-ultra-wasm.git && cd mcp-ultra-wasm

# 2. Configure variáveis de ambiente
cp mcp/mcp-ultra-wasm/.env.example mcp/mcp-ultra-wasm/.env && docker-compose -f mcp/mcp-ultra-wasm/docker-compose.yml up -d

# 3. Verifique o health check
curl http://localhost:8080/health  # ✅ Pronto!
```

## 📋 Estrutura do Projeto

```
mcp-ultra-wasm/
├── mcp/mcp-ultra-wasm/          # 🏗️ Servidor Go Principal
│   ├── cmd/wasm-server/     # Entrypoint do servidor wasm
│   ├── internal/wasm/       # Lógica de negócio
│   ├── wasm/                # Frontend + WASM
│   └── deploy/                  # Docker + K8s configs
├── sdk/sdk-ultra-wasm/          # 🔌 SDK para integração
└── templates/                   # 📋 Templates e blueprints
```

## 🎯 Acesso Rápido

**Após subir os serviços:**

- **🌐 Web Interface**: http://localhost:8080
- **📊 Health Check**: http://localhost:8080/health
- **🔌 API Endpoints**: http://localhost:8080/api/v1/
- **📡 WebSocket**: ws://localhost:8080/ws

## 📚 Documentação Completa

- **📖 README Principal**: [`mcp/mcp-ultra-wasm/README.md`](mcp/mcp-ultra-wasm/README.md)
- **🏗️ Arquitetura**: [`docs/ARCHITECTURE.md`](mcp/mcp-ultra-wasm/docs/ARCHITECTURE.md)
- **🔧 API Reference**: [`docs/API.md`](mcp/mcp-ultra-wasm/docs/API.md)
- **🐳 Docker Deploy**: [`docs/DEPLOY.md`](mcp/mcp-ultra-wasm/docs/DEPLOY.md)

## 🛠️ Stack Implementado

| Componente | Tecnologia | Status |
|------------|------------|--------|
| **Servidor** | Go + Gin | ✅ Completo |
| **WASM** | Go → JavaScript | ✅ Compilando |
| **Frontend** | HTML5 + CSS3 + JS | ✅ Responsivo |
| **WebSocket** | Gorilla WebSocket | ✅ Real-time |
| **Messaging** | NATS JetStream | ✅ Event-driven |
| **Auth** | JWT + RBAC | ✅ Seguro |
| **Observabilidade** | Prometheus + OpenTelemetry | ✅ Monitoramento |
| **Deploy** | Docker + Kubernetes | ✅ Production-ready |

## 🚀 Features Implementadas

### ✅ Core wasm Platform
- [x] **Servidor Go** com Gin framework
- [x] **Módulo WASM** compilado de Go para JavaScript
- [x] **Frontend completo** com HTML, CSS, JavaScript
- [x] **WebSocket real-time** para comunicação bidirecional

### ✅ Security & Authentication
- [x] **JWT authentication** com refresh tokens
- [x] **RBAC authorization** system
- [x] **CORS middleware** configuration
- [x] **Rate limiting** com token bucket algorithm

### ✅ Communication & Messaging
- [x] **NATS JetStream** para mensageria assíncrona
- [x] **Event-driven architecture** com publish/subscribe
- [x] **WebSocket handlers** para atualizações real-time
- [x] **API REST** endpoints completos

### ✅ Observability & Monitoring
- [x] **Prometheus metrics** collection
- [x] **OpenTelemetry tracing** distribuído
- [x] **Structured logging** com Zap
- [x] **Health check** endpoints

### ✅ Production Deployment
- [x] **Docker containerização** multi-stage
- [x] **Kubernetes manifests** completos
- [x] **Docker Compose** para desenvolvimento
- [x] **Security hardening** no container

## 🎯 Para Começar

### 1. Clonar e Configurar
```bash
git clone https://github.com/vertikon/mcp-ultra-wasm.git
cd mcp-ultra-wasm
cd mcp/mcp-ultra-wasm
cp .env.example .env
```

### 2. Desenvolvimento Local
```bash
# Instalar dependências Go
go mod download

# Compilar módulo WASM
$env:GOOS="js"; $env:GOARCH="wasm"; go build -o ../wasm/wasm/main.wasm ../wasm/wasm/main.go

# Rodar servidor
go run ./cmd/wasm-server
```

### 3. Produção com Docker
```bash
# Subir stack completo
docker-compose up -d

# Verificar status
curl http://localhost:8080/health
```

## 📝 Exemplos de Uso

### API REST
```bash
# Criar task via API
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"name":"analyze","config":{"project":"my-project"}}'

# Listar tasks
curl http://localhost:8080/api/v1/tasks
```

### WebSocket
```javascript
// Conectar WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => console.log('Received:', event.data);
```

### WASM no Browser
```javascript
// Carregar e executar módulo WASM
import { loadWasm } from './js/wasm-loader.js';
const wasmModule = await loadWasm();
const result = wasmModule.analyzeProject(config);
```

## 🧪 Testes

```bash
# Rodar todos os testes
go test ./...

# Com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📈 Status do Projeto

- ✅ **Build**: Compilando sem erros
- ✅ **Testes**: Cobertura completa
- ✅ **Linter**: 0 issues
- ✅ **Security**: Scan passed
- ✅ **Deploy**: Production ready

## 🤝 Contribuições

Contribuições são bem-vindas! Veja:
- [`CONTRIBUTING.md`](mcp/mcp-ultra-wasm/CONTRIBUTING.md)
- [Issues](https://github.com/vertikon/mcp-ultra-wasm/issues)
- [Pull Requests](https://github.com/vertikon/mcp-ultra-wasm/pulls)

---

**🎉 Parabéns! Você tem uma plataforma wasm completa rodando!**

Made with ❤️ by [Vertikon Labs](https://github.com/vertikon)