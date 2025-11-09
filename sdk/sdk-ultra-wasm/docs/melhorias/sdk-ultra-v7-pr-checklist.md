# PR Checklist — mcp-ultra-wasm-sdk — Validação V7

**Gerado em:** 2025-10-12
**Fonte:** gaps-report-2025-10-12.json
**Score atual:** 65/100
**Total de GAPs:** 7 (3 críticos, 0 médios, 4 baixos)
**Auto-fixáveis:** 0 | **Requerem análise manual:** 7

## Resumo dos Bloqueantes (GAPs Críticos)
1. **Compilação falha** — `go mod tidy` necessário
2. **Testes falham** — dependem do item 1
3. **Nil pointer** (1 ocorrência) — `context.go:24` type assertion sem verificação `ok`


## Padrões Gerais
- Commits convencionais: `fix:`, `feat:`, `refactor:`, `chore:`, `docs:`, `test:`, `ci:`
- PLANEJAR → EXECUTAR → VALIDAR
- DoD: build e testes passando, linter/vet limpos, cobertura ≥ 80% onde indicado, validator V7 “Aprovado”

## Ambiente local (pré-requisitos)
```pwsh
# Windows + WSL/Ubuntu opcional
go version
golangci-lint version
go env GOPATH
```


## Passo-a-passo (para Claude Code)


### 1) Resolver **dependências e compilação** (BLOQUEANTE) ⛔
**GAP crítico detectado:** Compilação falha com `go: updates to go.mod needed`

**Ação**
1. Executar `go mod tidy` para sincronizar dependências
2. Verificar conflitos de versão no `go.mod`
3. Executar build completo para validar

**Comandos**
```powershell
# Windows (PowerShell)
cd E:\vertikon\business\SaaS\templates\sdk-ultra-wasm
go mod tidy
go build ./...
```

**DoD**
- ✅ `go build ./...` executa sem erros
- ✅ `go.mod` e `go.sum` atualizados e sincronizados
- ✅ Nenhum erro de import ou símbolo não encontrado

**Severidade:** CRÍTICO | **Esforço:** 5min | **Commit:** `fix: sincroniza dependências go.mod`


### 2) Executar **testes** (BLOQUEANTE) ⛔
**GAP crítico detectado:** Testes falham (depende do passo 1)

**Ação**
1. Garantir que passo 1 foi concluído com sucesso
2. Executar testes com race detector
3. Analisar falhas e corrigir

**Comandos**
```powershell
# Executar testes com race detection
go test ./... -race -count=1 -v

# Se falhar, executar por pacote para isolar problemas
go test -v ./pkg/...
go test -v ./internal/...
```

**DoD**
- ✅ Todos os testes passam sem erros
- ✅ Race detector não reporta data races
- ✅ Coverage pode ser calculado (meta: ≥70%)

**Severidade:** CRÍTICO | **Esforço:** 15-30min | **Commit:** `fix: corrige testes falhando`


### 3) Blindar **nil pointer** e **type assertions** (BLOQUEANTE) ⛔
**GAP crítico detectado:** 1 type assertion sem verificação em `context.go:24`

**Ação**
1. Localizar `context.go:24` e verificar a type assertion
2. Substituir padrão perigoso por verificação segura
3. Varredura adicional em todo o código

**Exemplo de correção**
```go
// ❌ ANTES (perigoso)
value := ctx.Value(key).(string)

// ✅ DEPOIS (seguro)
value, ok := ctx.Value(key).(string)
if !ok {
    return fmt.Errorf("tipo inválido para key %v", key)
}
```

**Comandos**
```powershell
# Buscar todas as type assertions no código
rg '\.\([A-Z].*\)' --glob '*.go' -n

# Executar linter com foco em nil safety
golangci-lint run --enable=staticcheck,nilnil
```

**DoD**
- ✅ `context.go:24` corrigido com verificação `ok`
- ✅ Nenhuma type assertion sem verificação no codebase
- ✅ Testes executam sem `panic: interface conversion`

**Severidade:** CRÍTICO | **Esforço:** 10-15min | **Commit:** `fix: adiciona verificação segura de type assertion em context.go`


### 4) Corrigir **configuração do linter** (GAP BAIXO) ⚠️
**GAP baixo detectado:** 10 warnings de configuração deprecated no golangci-lint

**Ação**
1. Atualizar `.golangci.yml` removendo opções descontinuadas
2. Substituir linters deprecated por equivalentes modernos

**Mapeamento de substituições**
```yaml
# ❌ Remover/Substituir:
- interfacer    → (removido, sem substituto)
- maligned      → (removido, sem substituto)
- goerr113      → err113
- gomnd         → mnd
- exportloopref → copyloopvar (Go 1.22+)
- deadcode      → unused
- varcheck      → unused

# ❌ Remover opções descontinuadas:
run.skip-files     → issues.exclude-files
run.skip-dirs      → issues.exclude-dirs
linters.govet.check-shadowing → habilitar linter 'shadow'
```

**Comandos**
```powershell
# Validar configuração após edição
golangci-lint run --config .golangci.yml
```

**DoD**
- ✅ Nenhum warning de deprecation ao rodar linter
- ✅ Linters modernos habilitados (err113, mnd, copyloopvar, unused)
- ✅ Pipeline CI passa com nova config

**Severidade:** BAIXO | **Esforço:** 10min | **Commit:** `chore: atualiza config golangci-lint removendo opções deprecated`

---

### 5) Implementar **logs estruturados** (GAP BAIXO) ⚠️
**GAP baixo detectado:** Logs estruturados não encontrados no código

**Ação**
1. Escolher biblioteca (recomendado: `slog` nativo Go 1.21+ ou `zap`)
2. Substituir `fmt.Println`, `log.Print` por logger estruturado
3. Adicionar campos de contexto (trace_id, user_id, etc.)

**Exemplo com slog (nativo)**
```go
import "log/slog"

// Setup no main
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
slog.SetDefault(logger)

// Uso
slog.Info("operação concluída",
    "user_id", userID,
    "duration_ms", elapsed.Milliseconds(),
)
```

**DoD**
- ✅ Biblioteca de log estruturado configurada
- ✅ Logs de produção em formato JSON
- ✅ Campos de contexto em todos os logs importantes

**Severidade:** BAIXO | **Esforço:** 20-30min | **Commit:** `feat: adiciona logs estruturados com slog`

---

### 6) Calcular **cobertura de testes** (GAP BAIXO) ⚠️
**GAP baixo detectado:** Coverage não pôde ser calculado (depende de testes passando)

**Ação**
1. Garantir que passos 1-3 foram concluídos
2. Executar testes com coverage profile
3. Analisar arquivos com baixa cobertura

**Comandos**
```powershell
# Gerar relatório de cobertura
go test ./... -coverprofile=coverage.out -covermode=atomic

# Ver cobertura por função
go tool cover -func=coverage.out

# Gerar HTML para análise visual
go tool cover -html=coverage.out -o coverage.html
```

**DoD**
- ✅ Coverage ≥ 70% (meta atual do validator)
- ✅ Pacotes críticos (internal/core, pkg/orchestrator) com ≥85%
- ✅ Relatório HTML gerado para revisão

**Severidade:** BAIXO | **Esforço:** 5min (execução) + variável (escrita de testes) | **Commit:** `test: adiciona testes para atingir 70% coverage`


### 7) Completar **README** (GAP BAIXO) ⚠️
**GAP baixo detectado:** README incompleto — faltam seções "descrição" e "instalação"

**Ação**
1. Adicionar descrição clara do propósito do SDK
2. Adicionar seção de instalação/setup
3. Incluir exemplo de uso básico

**Template mínimo**
```markdown
# sdk-ultra-wasm

## Descrição
SDK para [descrever propósito principal]...

## Instalação
\```bash
go get github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm
\```

## Uso
\```go
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/..."

// Exemplo básico
\```

## Configuração
- `GOMEMLIMIT`: limite de memória (alinhado ao pod limit)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: endpoint de telemetria
- `LOG_LEVEL`: debug|info|warn|error

## Observabilidade
- Métricas: `/metrics`
- Health: `/healthz`, `/readyz`
```

**DoD**
- ✅ README contém todas as seções obrigatórias
- ✅ Desenvolvedor consegue executar exemplo do zero
- ✅ Seções de configuração e observabilidade documentadas

**Severidade:** BAIXO | **Esforço:** 15min | **Commit:** `docs: completa README com instalação e uso`

---

## Validação Final (após todos os passos)

### Checklist de aprovação
Execute após completar os passos 1-7:

```powershell
# 1. Build limpo
go build ./...

# 2. Testes passando
go test ./... -race -count=1

# 3. Linter limpo
golangci-lint run

# 4. Coverage adequado
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | Select-String "total"

# 5. Validator V7 (se disponível)
go run ./tools/validator/enhanced_validator_v7.go .
```

**Meta:** Score ≥ 85/100 no Validator V7

---

## Ordem de Execução Recomendada

**Bloqueantes primeiro (ordem sequencial):**
1. Passo 1 → Passo 2 → Passo 3

**Melhorias em paralelo (podem ser feitas simultaneamente):**
- Passo 4 (linter config)
- Passo 5 (logs estruturados)
- Passo 7 (README)

**Validação final:**
- Passo 6 (coverage) — após passos 1-3 concluídos

**Estimativa total:** 1h30min - 2h30min (depende de quantos testes precisam ser escritos)



## Arquivos Impactados por GAP

| Arquivo | GAP | Severidade | Ação |
|---------|-----|------------|------|
| `go.mod` / `go.sum` | Dependências | CRÍTICO | go mod tidy |
| `context.go:24` | Nil pointer | CRÍTICO | Adicionar verificação `ok` |
| `.golangci.yml` | Config deprecated | BAIXO | Atualizar linters |
| `README.md` | Documentação | BAIXO | Adicionar seções |
| `internal/**/*.go` | Logs estruturados | BAIXO | Implementar slog |
| `**/*_test.go` | Coverage | BAIXO | Adicionar testes |

---

## Apêndice: Detalhes dos GAPs (do relatório)

<details>
<summary>GAP Crítico 1: Compilação falha</summary>

```json
{
  "Type": "Código compila",
  "Severity": "crítico",
  "Description": "go: updates to go.mod needed; to update it:\n\tgo mod tidy",
  "Fixability": {
    "Safe": false,
    "RequiresReview": true,
    "NonFixableReason": "BUSINESS_LOGIC"
  }
}
```
</details>

<details>
<summary>GAP Crítico 2: Testes falham</summary>

```json
{
  "Type": "Testes PASSAM",
  "Severity": "crítico",
  "Description": "Testes falharam (depende do GAP 1)",
  "Fixability": {
    "RequiresReview": true,
    "NonFixableReason": "BUSINESS_LOGIC"
  }
}
```
</details>

<details>
<summary>GAP Crítico 3: Nil pointer em context.go:24</summary>

```json
{
  "Type": "Nil Pointer Check",
  "Severity": "crítico",
  "Location": "context.go:24",
  "Description": "type assertion sem verificação",
  "ManualSteps": [
    "1. Para type assertions: value, ok := x.(Type)",
    "2. Sempre verifique if !ok",
    "3. Para pointer dereference: if ptr != nil { ptr.Field }",
    "4. Considere usar nilaway para análise estática"
  ]
}
```
</details>

---

## Recursos e Ferramentas

**Linters e Análise**
- golangci-lint: `winget install golangci.golangci-lint`
- nilaway: `go install go.uber.org/nilaway/cmd/nilaway@latest`

**Observabilidade**
- OTEL SDK: `go.opentelemetry.io/otel`
- Prometheus: `github.com/prometheus/client_golang`

**GC Tuner (já incluído no SDK)**
- Path: `./internal/gctuner` ou `./pkg/gctuner`
- Uso: inicializar no `main()` antes de cargas de trabalho

**Error Analyzer (já incluído no SDK)**
- Path: `./tools/analyzers/errorsan`
- Build: `go build -o ../../bin/errorsan.exe .`
- Uso: `go vet -vettool=./bin/errorsan ./...`

---

**Última atualização:** 2025-10-12
**Validator:** V7
**Target Score:** 85/100