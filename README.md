# MCP Ultra Templates

Projeto para transformar as bases MCP (`mcp`, `sdk` e `mcp-wasm`) em templates reutilizáveis via CLI, com observabilidade, pipeline CI e suporte a execução containerizada.

## Sumário

- [Requisitos](#requisitos)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Build & Testes](#build--testes)
- [Execução da CLI](#execução-da-cli)
- [Trabalhando com Templates](#trabalhando-com-templates)
- [Modo Interativo](#modo-interativo)
- [Containerização & Docker Compose](#containerização--docker-compose)
- [Observabilidade](#observabilidade)
- [CI/CD](#cicd)

## Requisitos

- Go 1.22+
- Docker 24+ (opcional, para containerização)
- Docker Compose v2 (opcional)

## Estrutura do Projeto

```
.
├── cmd/                       # CLI principal
├── internal/                  # Config, handlers, serviços, repositórios
├── pkg/                       # Pacotes compartilhados (log, metrics, template, etc.)
├── templates/                 # Templates disponíveis (mcp, sdk, mcp-wasm)
├── deploy/prometheus/         # Configuração Prometheus para docker-compose
├── tools/coverage             # Utilitário interno para cálculo de cobertura
├── Dockerfile                 # Build multi-stage da CLI
├── docker-compose.yaml        # Stack local com Prometheus + Jaeger
└── .github/workflows/ci.yml   # Pipeline de lint + testes
```

## Build & Testes

```bash
# baixar dependências
go mod download

# build do binário CLI
go build -o bin/mcp-templates ./cmd

# execução de todos os testes
go test ./...

# testes com cobertura
go test ./... -coverprofile=coverage.out
go tool cover -func coverage.out
```

## Execução da CLI

```bash
# listar templates disponíveis
go run ./cmd -- list

# renderizar template informando valores via --set
go run ./cmd -- render \
  --template mcp \
  --output ./out/mcp-service \
  --set module_name=github.com/example/mcp-service
```

### Flags principais

| Flag            | Descrição                                                           |
|----------------|----------------------------------------------------------------------|
| `--template`    | Nome do template (`mcp`, `sdk`, `mcp-wasm`).                         |
| `--output`      | Diretório para gerar o projeto.                                      |
| `--values`      | Arquivo YAML com variáveis.                                          |
| `--set`         | Define variáveis no formato `chave=valor` (pode ser usado múltiplas vezes). |
| `--overwrite`   | Permite limpar o diretório de destino caso não esteja vazio.         |
| `--interactive` | Solicita interativamente variáveis obrigatórias ausentes.            |

## Trabalhando com Templates

| Template   | Diretório base             | Uso típico                                                  |
|------------|---------------------------|-------------------------------------------------------------|
| `mcp`      | `templates/mcp`           | Serviço MCP completo com APIs, observabilidade e deploys.   |
| `sdk`      | `templates/sdk`           | SDK do cliente com CLI auxiliar e exemplos de seeds.        |
| `mcp-wasm` | `templates/mcp-wasm`      | Gateway WASM com servidor Go + assets estáticos.            |

Exemplo para cada template:

```bash
# MCP
go run ./cmd -- render --template mcp --output out/mcp --set module_name=github.com/acme/mcp

# SDK
go run ./cmd -- render --template sdk --output out/sdk --set module_name=github.com/acme/sdk

# MCP WASM
go run ./cmd -- render --template mcp-wasm --output out/mcp-wasm --set module_name=github.com/acme/mcp-wasm
```

## Modo Interativo

Use `--interactive` para preencher variáveis obrigatórias que ainda não possuam valor (via defaults, `--set` ou arquivo YAML). Exemplo:

```bash
go run ./cmd -- render \
  --template mcp \
  --output ./out/mcp-interactive \
  --interactive
```

Ao executar, a CLI exibirá prompts para cada variável obrigatória pendente, aplicando os `defaults` definidos em `template.yaml` sempre que possível.

## Containerização & Docker Compose

### Build do container

```bash
docker build -t mcp-templates-cli .
```

### Stack local com observabilidade

```bash
docker compose up -d

# acessar shell do container para gerar projetos
docker compose exec cli mcp-templates list
docker compose exec cli mcp-templates render --template mcp ...
```

Volumes montados:

- `./templates` → `/workspace/templates` (somente leitura)
- `./out` → `/workspace/out` (saída dos projetos gerados)

## Observabilidade

- **Prometheus UI**: <http://localhost:9090>  
- **Jaeger UI**: <http://localhost:16686>  
- **Endpoint de métricas da CLI**: <http://localhost:2112/metrics>

Variáveis de ambiente relevantes (configuradas no container):

| Variável               | Descrição                               | Default               |
|------------------------|------------------------------------------|-----------------------|
| `TEMPLATES_PATH`       | Caminho dos templates                    | `/workspace/templates`|
| `OBS_ENABLE_METRICS`   | Habilita exposição Prometheus            | `true`                |
| `OBS_METRICS_ADDRESS`  | Endereço de bind para métricas           | `:2112`               |
| `OBS_ENABLE_TRACING`   | Habilita envio de traces OTLP            | `true` (no compose)   |
| `OBS_OTLP_ENDPOINT`    | Endpoint OTLP/Jaeger                     | `jaeger:4317`         |

## CI/CD

Pipeline GitHub Actions (`.github/workflows/ci.yml`) executa:

1. `golangci-lint` (config em `.golangci.yml`)
2. `go test ./... -coverprofile=coverage.out`
3. Upload do arquivo de cobertura como artifact

Para replicar localmente o lint:

```bash
golangci-lint run ./...
```

---

Ficou com dúvida ou encontrou oportunidade de melhoria? Abra uma issue ou contribua diretamente! :rocket:

