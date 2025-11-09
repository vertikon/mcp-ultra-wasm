# MCP Ultra WASM Web Interface

## Visão Geral

O componente **Web-WASM** fornece uma interface web moderna e interativa para a plataforma MCP Ultra WASM, permitindo que usuários executem tarefas WASM diretamente no navegador com performance nativa e sandboxing seguro.

## 🚀 Funcionalidades

### Core Features
- **Interface Web Moderna**: UI responsiva com design intuitivo
- **Execução WASM Nativa**: Módulos Go compilados para WebAssembly executados diretamente no browser
- **Integração NATS**: Comunicação assíncrona com o ecossistema MCP via eventos
- **WebSocket Real-time**: Atualizações em tempo real do status das tarefas
- **SDK Ultra WASM**: Integração completa com o SDK existente da plataforma

### Funcionalidades Específicas
- **Análise de Projetos**: Analise projetos Go diretamente na web
- **Geração de Código**: Gere código baseado em especificações
- **Validação de Configurações**: Valide configurações de projetos e deployments
- **Execução de Tasks**: Execute tarefas personalizadas via interface web

## 🏗️ Arquitetura

### Componentes Principais

```
web-wasm/
├── cmd/web-wasm-server/     # Servidor web Go
├── internal/web-wasm/
│   ├── handlers/            # Handlers HTTP
│   ├── nats/               # Cliente NATS
│   ├── sdk/                # Cliente SDK Ultra WASM
│   ├── observability/     # Métricas, tracing, logging
│   └── security/          # Autenticação, CORS, rate limiting
├── web-wasm/
│   ├── wasm/               # Código WASM compilado
│   ├── static/             # Arquivos estáticos (CSS, JS)
│   └── templates/          # Templates HTML
└── README.md               # Este arquivo
```

### Fluxo de Comunicação

1. **Frontend → Servidor**: Requisições HTTP/HTTPS
2. **Servidor → NATS**: Publicação de eventos de tarefas
3. **NATS → Workers MCP**: Entrega de eventos para processamento
4. **Workers → SDK**: Execução via SDK Ultra WASM
5. **Resultados → NATS**: Publicação de resultados
6. **NATS → Servidor → Frontend**: Entrega via WebSocket

## 🛠️ Tecnologias Utilizadas

### Backend
- **Go 1.24.0**: Linguagem principal
- **Gin**: Framework web HTTP
- **NATS JetStream**: Messaging e streaming
- **WebSocket**: Comunicação real-time
- **Prometheus**: Métricas e monitoring
- **OpenTelemetry**: Tracing distribuído
- **Zap**: Structured logging

### Frontend
- **HTML5/CSS3/JavaScript**: Interface web moderna
- **WebAssembly**: Execução de código Go no browser
- **WebSocket**: Comunicação real-time
- **Responsive Design**: Adaptação a diferentes dispositivos

### Segurança
- **JWT**: Autenticação e autorização
- **RBAC**: Controle de acesso baseado em roles
- **CORS**: Cross-Origin Resource Sharing
- **Rate Limiting**: Limitação de requisições
- **HTTPS**: Comunicação segura

## 📋 Pré-requisitos

### Desenvolvimento
- Go 1.24.0 ou superior
- Node.js 18+ (para ferramentas)
- NATS Server 2.10+

### Produção
- Kubernetes 1.24+ (ou Docker Compose)
- PostgreSQL 17+ (opcional)
- Redis 7+ (opcional, para cache)
- Prometheus + Grafana (monitoring)

## 🚀 Instalação e Configuração

### 1. Clone o Repositório
```bash
git clone https://github.com/vertikon/mcp-ultra-wasm.git
cd mcp-ultra-wasm/mcp/mcp-ultra-wasm
```

### 2. Build da Aplicação
```bash
# Build do servidor
go build -o bin/web-wasm-server ./cmd/web-wasm-server

# Build do módulo WASM
GOOS=js GOARCH=wasm go build -o web-wasm/wasm/main.wasm ./web-wasm/wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./web-wasm/wasm/
```

### 3. Execução Local
```bash
# Iniciar o servidor
./bin/web-wasm-server

# Ou com make
make run
```

### 4. Deploy com Docker
```bash
# Build da imagem
docker build -f deploy/docker/web-wasm/Dockerfile -t web-wasm-server .

# Executar com Docker Compose
cd deploy/docker/web-wasm
docker-compose up -d
```

### 5. Deploy com Kubernetes
```bash
# Aplicar manifests
kubectl apply -f deploy/k8s/web-wasm/

# Verificar status
kubectl get pods -n web-wasm
kubectl get services -n web-wasm
```

## ⚙️ Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-------------|---------|
| `PORT` | Porta do servidor | `8080` |
| `NATS_URL` | URL do NATS | `nats://localhost:4222` |
| `SDK_ADDRESS` | Endereço do SDK | `localhost:9091` |
| `LOG_LEVEL` | Nível de log | `info` |
| `JWT_SECRET` | Secret JWT | `change-me` |
| `ENV` | Ambiente | `development` |

### Arquivo de Configuração

Veja `deploy/docker/web-wasm/config.yaml` para configuração detalhada.

## 🧪 Testes

### Executar Testes Unitários
```bash
go test ./...
```

### Executar Testes de Integração
```bash
go test -tags=integration ./...
```

### Cobertura de Testes
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📊 Monitoramento e Observabilidade

### Métricas
- **HTTP Requests**: Taxa de requisições, tempo de resposta, status codes
- **WASM Operations**: Tempo de execução, uso de memória
- **SDK Operations**: Chamadas ao SDK, taxas de sucesso/erro
- **WebSocket**: Conexões ativas, mensagens trocadas

### Endpoints
- `/metrics`: Métricas Prometheus
- `/health`: Health check
- `/security/info`: Informações de segurança

### Dashboard Grafana
O dashboard está disponível em `http://localhost:3001` (Docker Compose) ou configure seu próprio dashboard.

## 🔧 Desenvolvimento

### Estrutura de Projetos
O projeto segue a estrutura padrão Go:

```
mcp-ultra-wasm/
├── cmd/                    # Pontos de entrada
├── internal/              # Código interno
│   ├── web-wasm/        # Componente web-wasm
│   ├── handlers/        # Handlers HTTP
│   ├── nats/           # Cliente NATS
│   ├── sdk/            # Cliente SDK
│   ├── observability/  # Observabilidade
│   └── security/       # Segurança
├── web-wasm/             # Frontend e WASM
├── test/                  # Testes
├── deploy/                # Configurações de deploy
└── docs/                  # Documentação
```

### Adicionando Novas Funcionalidades WASM

1. **Criar nova função em `web-wasm/wasm/functions/`**
```go
func NewFeature(this js.Value, args []js.Value) interface{} {
    callback := args[len(args)-1]
    // Implementar lógica
    go func() {
        result := processNewFeature()
        callback.Invoke(result)
    }()
    return nil
}
```

2. **Registrar função em `web-wasm/wasm/main.go`**
```go
js.Global().Set("newFeature", js.FuncOf(functions.NewFeature))
```

3. **Atualizar frontend em `web-wasm/static/js/main.js`**
```javascript
case 'new_feature':
    this.wasmModule.newFeature(config, callback);
    break;
```

### Desenvolvimento com Hot Reload

```bash
# Usar Air para recarregar automaticamente
air -c 'go build -o bin/web-wasm-server ./cmd/web-wasm-server && ./bin/web-wasm-server'

# Ou com make
make dev
```

## 🔐 Segurança

### Autenticação
- **JWT Tokens**: Tokens assinados com expiração configurável
- **Refresh Tokens**: Renovação automática de tokens
- **Blacklist**: Revogação de tokens comprometidos

### Autorização
- **RBAC**: Controle de acesso baseado em roles
- **Permissões Granulares**: Controle fino de operações
- **Rate Limiting**: Proteção contra abuso

### Segurança Web
- **CORS**: Controle de acesso cross-origin
- **Headers de Segurança**: HSTS, XSS Protection, Content Security Policy
- **Rate Limiting**: Limitação de requisições por IP/usuário

## 📈 Performance

### Otimizações
- **WebAssembly**: Execução nativa no browser
- **Cache de Módulos**: Cache inteligente de módulos WASM
- **Connection Pooling**: Reuso de conexões NATS
- **Streaming**: Processamento streaming de dados grandes

### Métricas de Performance
- **Latency**: < 100ms para operações comuns
- **Throughput**: 1000+ requisições/segundo
- **Memory**: < 128MB por instância
- **CPU**: < 500m por instância

## 🤝 Contribuição

### Como Contribuir
1. Fork o repositório
2. Crie uma branch (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

### Guidelines de Código
- Seguir padrões Go (`gofmt`, `golint`)
- Escrever testes para novas funcionalidades
- Documentar código publico
- Seguir convenções do projeto

## 📝 Licença

Este projeto está licenciado sob a Licença MIT. Veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

### Documentação
- [Documentação Técnica](../MCP-ULTRA-WASM-DOCUMENTACAO-TECNICA.md)
- [API Reference](../docs/api.md)
- [Guia de Desenvolvimento](../docs/development.md)

### Comunidade
- Issues: [GitHub Issues](https://github.com/vertikon/mcp-ultra-wasm/issues)
- Discussões: [GitHub Discussions](https://github.com/vertikon/mcp-ultra-wasm/discussions)

### Contato
- Email: support@vertikon.com
- Website: https://vertikon.com

## 🗺️ Roadmap

### v1.1 (Planejado)
- [ ] Plugin system para extensões
- [ ] Multi-tenant completo
- [ ] Advanced analytics
- [ ] CI/CD integration

### v1.2 (Planejado)
- [ ] Machine Learning integration
- [ ] Custom dashboards
- [ ] Advanced security features
- [ ] GraphQL API

## 📋 Changelog

### v1.0.0 (Atual)
- Release inicial do Web-WASM
- Interface web completa
- Integração NATS e WebSocket
- Suporte básico ao SDK Ultra WASM
- Sistema de autenticação e autorização
- Monitoramento e observabilidade completos

---

**MCP Ultra WASM Web Interface** - Interface web moderna e performática para a plataforma MCP Ultra WASM.