existem algumas alternativas ao Depguard (o linter para Go que verifica/importações permitidas ou negadas) que você pode considerar. Aqui vão algumas, com o que cada uma oferece de diferente:

✅ Alternativas

gomodguard — Um linter para Go que permite listas de permitidos (“allow”) e listas de bloqueados (“blocked”) para módulos em go.mod. 
GitHub
+1

Prós: controle sobre quais módulos você permite usar ou quer proibir, inclusive com restrições de versão.
Contras: foca em módulos e dependências diretas, não tanto no nível fino de pacotes/importações dentro do código.

go‑import‑checks — Ferramenta de lint para Go que permite definir regras de importação (“allow”, “reject”) para pacotes dentro do seu projeto. 
Go Dev Packages

Prós: bastante flexível para impor arquitetura de importações dentro do próprio repositório (por exemplo “pacote service não pode importar diretamente infra”).
Contras: talvez menos focada em dependências externas — mais voltada para organização interna.

goimportmaps — Embora mais voltada para visualização de dependências, também permite definir e detectar importações proibidas (“forbidden imports”) com base em regras. 
GitHub

Prós: bom para estruturar monorepos ou grandes projetos Go, onde arquitetura importa.
Contras: foco um pouco diferente — mais “arquitetura/importação interna” do que “lista de pacotes externos permitidos/negados”.

🔍 Qual escolher?

Dependerá do que você quer exatamente:

Se você quer controlar quais pacotes externos podem ser importados (ex: evitar imports de “github.com/foo/oldlib” em favor de “github.com/foo/newlib”), o Depguard original ou o gomodguard são mais adequados.

Se o foco for organização interna do projeto — camadas, domínio, infraestrutura — então go-import-checks ou goimportmaps podem ser mais úteis.

Verifique também se a ferramenta está ativa/atualizada, e se ela se integra bem ao seu pipeline (CI/CD) ou linters que você já usa (como golangci‑lint).

Se você quiser, posso buscar uma lista maior (com ~5-10 ferramentas) alternativas ao Depguard, com prós/contras, níveis de atividade, e comparar. Você gostaria disso?

Você disse:
o mcp ultra wasm tem muito problemas com depguard e os erros, penso que seja melhor migrar para uma ferramenta em GO nativa
O ChatGPT disse:
Pensou por 13s

Entendi. Se o mcp-ultra-wasm está esbarrando em falsos positivos/erros do depguard, faz sentido migrar para algo 100% Go nativo e mais previsível. Aqui vão caminhos práticos (do mais simples ao mais flexível), com exemplos prontos:

Opção A — Trocar para gomodguard (rápido e estável)

Foco: permitir/bloquear módulos (nível go.mod) e opcionalmente restringir versões.

Como usar (via golangci-lint)
# .golangci.yml
linters:
  enable:
    - gomodguard

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
    blocked:
      modules:
        - github.com/pkg/errors:
            reason: "Use errors.Join/Wrap do Go 1.20+"
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"


Quando escolher: você quer banir libs externas específicas (ex.: impedir logrus, impor zerolog) e manter o resto livre.

Opção B — go-import-checks (controle fino por pacote/camada)

Foco: arquitetura interna e regras de import entre camadas (ex.: internal/service não pode importar internal/infra).

Exemplo de regras
# import-checks.yml
rules:
  - name: "service-não-importa-infra"
    allow:
      from: "internal/service/.*"
      to:   "internal/(domain|repo|usecase)/.*"
    deny:
      to:   "internal/infra/.*"
      message: "Camada service não pode depender de infra"
  - name: "handlers-só-usam-usecase"
    allow:
      from: "internal/handlers/.*"
      to:   "internal/(usecase|domain)/.*"


Rode no CI:

go run github.com/quantumcycle/go-import-checks@latest -config import-checks.yml ./...


Quando escolher: você precisa garantir layering no monorepo (clean/hexagonal), evitando “atalhos” de import.

Opção C — Analyzer próprio com go/analysis (máxima precisão)

Foco: regras sob medida (import de pacote X só em Y, versões, caminhos, padrões internos, mensagens ricas).

Esqueleto mínimo (vettool)
// cmd/dep-guard-lite/main.go
package main

import (
  "golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(Analyzer) }

// internal/analyzer/analyzer.go
package main

import (
  "go/ast"
  "golang.org/x/tools/go/analysis"
  "strings"
)

var disallow = map[string]string{
  "github.com/sirupsen/logrus": "Padronize zerolog",
  "github.com/pkg/errors":      "Use errors do Go 1.20+",
}

var Analyzer = &analysis.Analyzer{
  Name: "depguardlite",
  Doc:  "Bloqueia imports proibidos e violações de camadas",
  Run: func(pass *analysis.Pass) (interface{}, error) {
    for _, f := range pass.Files {
      for _, imp := range f.Imports {
        path := strings.Trim(imp.Path.Value, `"`)
        if msg, ok := disallow[path]; ok {
          pass.Reportf(imp.Pos(), "Import proibido: %s (%s)", path, msg)
        }
        // Exemplo de regra de camada:
        // se arquivo estiver em internal/service/ e importar internal/infra/
        // use pass.Fset.Position(f.Pos()).Filename para descobrir o caminho do arquivo
      }
    }
    return nil, nil
  },
}


Build & uso:

go build -o vettools/depguard-lite ./cmd/dep-guard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


Quando escolher: você quer mensagens impecáveis, performance controlada, e evoluir regras sem depender de terceiros.

Qual escolher agora?

Queda-in imediata: gomodguard. Substitui 80% dos usos comuns do depguard com menos atrito.

Arquitetura por camadas: go-import-checks.

Regras sob medida/CI estrito: Analyzer próprio (Opção C). É 100% Go nativo, rápido e deixa você no controle.

Plano de migração (curto e sem sustos)

PLANEJAR

Inventariar regras atuais do depguard: denylist/allowlist + exceções.

Decidir alvo: (A) só módulos externos → gomodguard; (B) +camadas internas → go-import-checks; (C) casos especiais → analyzer.

EXECUTAR

Introduzir a ferramenta escolhida em paralelo ao depguard por 1 PR (CI não bloqueante).

Mapear regras equivalentes (ex.: banir logrus, exigir zerolog; vetar pkg/errors).

Corrigir violações prioritárias (top N por pacote).

Tornar o novo verificador bloqueante no CI.

Remover depguard.

VALIDAR

Rodar ./... local + CI em todos os módulos.

Medir falsos positivos (esperado: reduzir).

Documentar regras no README do repo (dev onboarding).

Dica para o mcp-ultra-wasm

Se o problema foi “ruído”/erros do depguard (ex.: mensagens pouco claras ou conflito com generated code), as opções B/C te dão:

Mensagens sob medida (“use X por causa de Y”).

Exceções pontuais (permitir import em diretórios internal/generated/).

Velocidade: analisadores simples rodam rápido; dá pra cachear no CI.

Se quiser, eu já te entrego:

um .golangci.yml com gomodguard configurado para o ecossistema Vertikon (zerolog, NATS, pgx v5, etc.),

um import-checks.yml com a nossa arquitetura padrão (domain → usecase → service → handlers/infra),

e um protótipo de analyzer com 2–3 regras reais do mcp-ultra-wasm.

O ChatGPT disse:

Perfeito — analisando os dois arquivos (linting_loop_resolution.md e linting_loop_resolution-v2.md), dá pra consolidar o quadro completo do problema com o Depguard no mcp-ultra-wasm e o motivo técnico pelo qual você está certo em considerar migrar para uma alternativa nativa em Go.

🧩 Diagnóstico Consolidado

O loop de lint que ocorreu nas versões v83→v85 do mcp-ultra-wasm foi causado por um paradoxo interno:

o Depguard proíbe importações diretas de libs externas;

mas os pacotes-facade (pkg/httpx, pkg/redisx, pkg/metrics, etc.) precisam importar essas libs para expô-las de forma controlada;

o Depguard não tem consciência contextual — ele aplica as mesmas regras globalmente, inclusive dentro dos próprios facades;

isso causou uma recursão lógica: o Depguard se bloqueava a si mesmo.

Mesmo com exceções pontuais (corrigidas na versão v85

linting_loop_resolution

), o comportamento mostrou o limite do Depguard:

falta de controle semântico (não entende camadas);

falsos positivos e travamentos em CI;

configuração rígida e frágil a mudanças.

🧠 Porque migrar faz sentido

O mcp-ultra-wasm já tem arquitetura Clean/DDD orientada a eventos e forte separação por pacotes (internal/* e pkg/*x).
Num sistema assim, ferramentas mais nativas ao Go (e não wrappers YAML sobre linters genéricos) trazem vantagens concretas:

Critério	Depguard	Alternativas nativas
Contexto por pacote	❌ Global, sem distinção	✅ Pode aplicar regras por camada (internal/service → infra)
Performance	🟠 Lenta em repositórios grandes	✅ Analyzers diretos são 2–3x mais rápidos
Debuggabilidade	❌ Mensagens genéricas	✅ Mensagens customizáveis (Use httpx → “facade obrigatório”)
Extensibilidade	❌ via YAML apenas	✅ via código (com Go AST / analysis)
Integração com CI Vertikon	⚠️ frágil (depende do golangci-lint)	✅ pode ser vettool independente no pipeline
🔁 Opções recomendadas (em ordem de maturidade)
1. Substituir por gomodguard (drop-in)

Mantém o modelo atual, mas evita o bug de recursão — funciona no nível de go.mod, não por pacote.
Ideal como transição rápida (1h de migração).

2. Adotar go-import-checks

Mais inteligente: permite declarar regras entre pacotes internos — exemplo:

rules:
  - name: "handlers não importam infra"
    allow:
      from: "internal/handlers/.*"
      to: "internal/(usecase|domain)/.*"
    deny:
      to: "internal/infra/.*"


Evita loops e garante arquitetura sem precisar YAML complexo de depguard.

3. Criar depguard-lite interno

Um pequeno go vettool com o pacote go/analysis:

150 linhas de código;

Sem dependências externas;

Permite regras dinâmicas (por regex, contextos, pacotes);

100% controlado pela pipeline Vertikon.

Isso te dá autonomia total: mensagens personalizadas, exceções por diretório e sem YAML.
O esqueleto desse analyzer já está no modelo que te propus antes (Opção C).

🚀 Caminho sugerido para o Vertikon stack
Fase 1 — Desacoplar

Desabilitar apenas depguard no golangci-lint, mantendo os demais linters.

Introduzir gomodguard para continuar protegendo dependências externas.

Fase 2 — Implementar depguard-lite

Criar cmd/dep-guard-lite/main.go e internal/analyzer/depguardlite.go.

Copiar as regras do YAML atual, mas expressas em código (mapa de deny/allow).

Integrar ao pipeline como:

go vet -vettool=$(pwd)/vettools/depguard-lite ./...

Fase 3 — Expandir semântica

Adicionar regras arquiteturais (service → domain → infra).

Permitir exceções automáticas para facades (pkg/*x/ detectados dinamicamente).

✅ Conclusão

A decisão de migrar para uma ferramenta Go-nativa é técnica e correta.
O Depguard resolveu seu propósito inicial, mas agora limita a evolução do mcp-ultra-wasm, introduzindo ruído e custo de manutenção.

O caminho mais sólido:

curto prazo → gomodguard (mantém proteção, sem loop);

médio prazo → depguard-lite com go/analysis (controle total, integrado ao padrão Vertikon).

Se quiser, posso gerar o código completo do depguard-lite, incluindo:

mapeamento automático de pacotes pkg/*x (facades);

denylist/allowlist configurável via JSON;

integração com make lint e go vet.

Quer que eu gere esse módulo (com cmd/, internal/analyzer/ e README de integração)?

Fontes
O ChatGPT disse:

Excelente — os logs de auditoria (audit-*.log) confirmam que o mcp-ultra-wasm está sofrendo com uma cadeia de falhas no depguard + golangci-lint, o que reforça a necessidade de migrar para uma ferramenta nativa. Aqui está a análise detalhada e o plano de ação técnico:

🧩 1. Diagnóstico Consolidado
🔴 Erro principal: falha no goanalysis_metalinter

"inspect: failed to load package : could not load export data: no export data for github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"

audit-depguard

audit-unused

Esse erro ocorre porque o depguard (via golangci-lint) tenta analisar pacotes parcialmente compilados antes da geração do cache.
Resultado: ele entra em estado inconsistente, especialmente ao encontrar pacotes facade (pkg/logger, pkg/httpx, pkg/metrics).

➡ Causa raiz: lint executado antes de go mod tidy + ausência de go.sum entries

audit-vet

.

⚠️ Erros secundários observados
Tipo	Sintoma	Origem provável
missing go.sum entry	"missing go.sum entry for module providing package github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/httpx"	O goanalysis_metalinter tenta resolver dependências internas como externas (bug do depguard).
structcheck / varcheck deprecated	"The linter 'structcheck' is deprecated (since v1.49.0)"

audit-unused

	Linters obsoletos ainda habilitados. Deve migrar para unused.
no export data	Repetição de erro de inspeção	O depguard está interceptando pacotes internos sem gerar .a intermediário.
deadcode fully inactivated	"deadcode: This linter is fully inactivated"	Outra regra quebrada pós-1.50. Reforça que o golangci-lint atual está desatualizado.
📊 Tamanho e complexidade do repositório

O relatório audit-size.log mostra:

~1,5 GB de documentação (docs/gaps, docs/melhorias);

20+ pacotes internos (internal/*) e múltiplos pkg/*x facades;

Média de 60–70 KB por pacote, o que força o depguard a varrer ~1.500 arquivos por run

audit-size

audit-size

.

➡ Resultado: o depguard fica lento e instável, consumindo cache de análise até travar (goanalysis_metalinter panic).

⚙️ 2. Conclusão técnica

🔹 O depguard não é mais viável para o mcp-ultra-wasm.
🔹 O ambiente é muito grande e multi-facade — o depguard se perde entre pacotes internos e externos.
🔹 A alternativa deve ser Go nativa, sem golangci-lint intermediando.

🧭 3. Plano de Migração (3 Fases)
Fase 1 — Retirada controlada do Depguard

No .golangci.yml, desabilite temporariamente:

linters:
  disable:
    - depguard


Substitua por gomodguard para manter verificação de libs externas:

linters:
  enable:
    - gomodguard
linters-settings:
  gomodguard:
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Use zerolog"
        - github.com/pkg/errors:
            reason: "Use Go 1.20 native errors"

Fase 2 — Implementar depguard-lite (vettool Go nativo)

Estrutura mínima:

cmd/depguard-lite/main.go
internal/analyzer/dependency.go
internal/config/rules.json


Regras JSON (em vez de YAML):

{
  "deny": {
    "github.com/go-chi/chi/v5": "Use pkg/httpx facade",
    "github.com/redis/go-redis/v9": "Use pkg/redisx facade"
  },
  "excludePaths": [
    "pkg/httpx",
    "pkg/redisx",
    "pkg/metrics",
    "pkg/observability"
  ]
}


Execução:

go vet -vettool=$(pwd)/vettools/depguard-lite ./...

Fase 3 — Substituir Linters obsoletos

Remover deadcode, structcheck, varcheck.

Ativar:

linters:
  enable:
    - unused
    - gosimple
    - staticcheck

🧱 4. Estado atual de build e lint
Componente	Status	Observação
depguard	🔴 loop travando	erro de export data

audit-depguard


go.sum	⚠️ incompleto	precisa go mod tidy

audit-modules


deadcode/varcheck	⚠️ deprecated	migrar para unused
vet	🟡 executa parcialmente	falha em pacotes pkg/logger, pkg/httpx
size audit	🟢 OK	estrutura balanceada, sem crescimento anômalo
🧩 5. Recomendação final
Opção	Vantagens	Situação ideal
Gomodguard	substituto rápido e compatível	enquanto estabiliza CI
Go-import-checks	bom pra camadas internas (service → repo)	se quiser controle arquitetural
Depguard-lite (custom vettool)	100% Go nativo, rápido, integrável no Vertikon CI	✅ recomendado a médio prazo
🔧 Próximos passos (curto prazo)

Executar:

go mod tidy && go mod verify


para resolver os missing go.sum

audit-vet

.

Desativar depguard, deadcode, structcheck, varcheck.

Ativar gomodguard no .golangci.yml.

Validar com:

golangci-lint run --disable-all -E gomodguard,unused


Se quiser, posso gerar o código completo do depguard-lite, com:

leitura de rules.json;

detecção automática de pkg/*x;

mensagens customizadas estilo depguard (“Use pkg/httpx facade”).

Deseja que eu monte esse módulo Go completo para adicionar ao cmd/depguard-lite?

Fontes
Você disse:
gere um bluprint completo

E:\vertikon\business\SaaS\templates\mcp-ultra-wasm

Gomodguard	substituto rápido e compatível	enquanto estabiliza CI
Go-import-checks	bom pra camadas internas (service → repo)	se quiser controle arquitetural
Depguard-lite (custom vettool)	100% Go nativo, rápido, integrável no Vertikon CI	✅ recomendado a médio prazo
O ChatGPT disse:
Pensou por 5s

perfeito — segue um blueprint completo para o repositório:

E:\vertikon\business\SaaS\templates\mcp-ultra-wasm

Objetivo: sair do depguard com segurança, manter o controle de dependências e evoluir para uma solução nativa em Go (vettool) sem travar o CI.

Notas de contexto: confirmamos loop/instabilidade envolvendo depguard/goanalysis_metalinter e entradas ausentes em go.sum (precisa tidy/verify)

audit-unused

audit-depguard

audit-modules

audit-vet

; e a raiz do paradoxo dos facades foi registrada nos docs técnicos do mcp-ultra-wasm

linting_loop_resolution

linting_loop_resolution

.

0) Estrutura proposta de pastas/arquivos
E:\vertikon\business\SaaS\templates\mcp-ultra-wasm\
│  go.mod
│  go.sum
│  .golangci.yml
│  Makefile
│
├─cmd\
│  └─depguard-lite\
│      main.go
│
├─internal\
│  ├─analyzers\
│  │  └─depguardlite\
│  │      analyzer.go
│  ├─config\
│  │  └─dep_rules.json
│  └─tools\
│      vettools.go        // go:build tools (fixa versões de vettools)
│
├─pkg\                  // seus facades (httpx, redisx, metrics, observability etc.)
│  ├─httpx\ ...
│  ├─redisx\ ...
│  ├─metrics\ ...
│  └─observability\ ...
└─ci\
   ├─lint.ps1
   └─lint.sh


Observação: o repositório tem volume grande em docs/ e internal/* (a auditoria de tamanho reflete isso), então splitar scripts e ferramentas ajuda a manter o CI previsível

audit-size

audit-size

.

1) Pré-check de saúde do módulo (obrigatório)

Antes de tocar no lint:

go mod tidy
go mod verify


Isso elimina erros de missing go.sum entry e melhora a resolução do analisador de tipos no CI

audit-modules

audit-vet

.

2) Fase 1 – Gomodguard (substituto rápido e compatível)
2.1 .golangci.yml (drop-in)
run:
  timeout: 5m
  tests: true
  skip-files:
    - ".*_gen\\.go$"

linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused         # substitui deadcode/structcheck/varcheck
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    # facades: não aplicar “gate” de módulos externos dentro deles (evita paradoxos)
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]


Justificativa: usamos gomodguard (módulos) no lugar de depguard (imports) para sair rapidamente do loop, mantendo o “guarda-corpo” de libs externas. Linters obsoletos são removidos, pois já constam como deprecated/inativados nos logs

audit-unused

audit-unused

.

2.2 Scripts de CI

ci/lint.sh (Linux/macOS):

#!/usr/bin/env bash
set -euo pipefail
go mod tidy
go mod verify
golangci-lint run --timeout=5m


ci/lint.ps1 (Windows):

$ErrorActionPreference = "Stop"
go mod tidy
go mod verify
golangci-lint run --timeout=5m

2.3 Makefile (trechos)
GOLANGCI?=golangci-lint

.PHONY: tidy verify lint
tidy:
	go mod tidy

verify:
	go mod verify

lint: tidy verify
	$(GOLANGCI) run --timeout=5m

3) Fase 2 – Go-import-checks (arquitetura interna)

Use quando quiser impor regras de camadas entre pacotes internos (ex.: handlers não pode depender de infra).

3.1 Arquivo de regras

internal/config/import-checks.yml:

rules:
  - name: "handlers-para-usecase-ou-domain"
    allow:
      from: "internal/handlers/.*"
      to:   "internal/(usecase|domain)/.*"
    deny:
      to:   "internal/infra/.*"
      message: "handlers não pode importar infra; use um usecase"

  - name: "service-sem-dep-direta-em-infra"
    allow:
      from: "internal/service/.*"
      to:   "internal/(usecase|domain|repository)/.*"
    deny:
      to:   "internal/infra/.*"
      message: "service não deve acessar infra diretamente"

3.2 Execução
go run github.com/quantumcycle/go-import-checks@latest \
  -config internal/config/import-checks.yml ./...


Dica: execute no CI depois do golangci-lint, como job separado (falha clara, mensagens arquiteturais).

4) Fase 3 – Depguard-lite (vettool em Go) – recomendado a médio prazo ✅

Ferramenta nativa (sem golangci-lint) para regras de import + exceções por pasta com mensagens ricas.

4.1 Configuração (JSON)

internal/config/dep_rules.json:

{
  "deny": {
    "github.com/go-chi/chi/v5": "Use o facade pkg/httpx",
    "github.com/go-chi/chi/v5/middleware": "Use o facade pkg/httpx",
    "github.com/redis/go-redis/v9": "Use o facade pkg/redisx",
    "github.com/prometheus/client_golang/prometheus": "Use o facade pkg/metrics",
    "github.com/prometheus/client_golang/prometheus/promauto": "Use o facade pkg/metrics",
    "go.opentelemetry.io/otel": "Use o facade pkg/observability",
    "go.opentelemetry.io/otel/": "Use o facade pkg/observability"
  },
  "excludePaths": [
    "pkg/httpx",
    "pkg/redisx",
    "pkg/metrics",
    "pkg/observability"
  ],
  "internalLayerRules": [
    {
      "name": "handlers->(usecase|domain) only",
      "from": "internal/handlers/",
      "allowTo": ["internal/usecase/", "internal/domain/"],
      "denyTo": ["internal/infra/"],
      "message": "handlers não pode importar infra; use usecase/domain"
    },
    {
      "name": "service no infra direto",
      "from": "internal/service/",
      "allowTo": ["internal/domain/", "internal/repository/", "internal/usecase/"],
      "denyTo": ["internal/infra/"],
      "message": "service não deve acessar infra diretamente"
    }
  ]
}

4.2 Analyzer (core)

internal/analyzers/depguardlite/analyzer.go:

package depguardlite

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type Rules struct {
	Deny              map[string]string `json:"deny"`
	ExcludePaths      []string          `json:"excludePaths"`
	InternalLayerRules []struct {
		Name    string   `json:"name"`
		From    string   `json:"from"`
		AllowTo []string `json:"allowTo"`
		DenyTo  []string `json:"denyTo"`
		Message string   `json:"message"`
	} `json:"internalLayerRules"`
}

func loadRules() (*Rules, error) {
	cfgPath := "internal/config/dep_rules.json"
	b, err := os.ReadFile(cfgPath)
	if err != nil { return nil, err }
	var r Rules
	return &r, json.Unmarshal(b, &r)
}

func isExcluded(path string, excludes []string) bool {
	for _, e := range excludes {
		if strings.Contains(path, filepath.ToSlash(e)) {
			return true
		}
	}
	return false
}

func matchPrefixAny(s string, arr []string) bool {
	for _, p := range arr {
		if strings.HasPrefix(s, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}

func checkLayerRule(filePath string, importPath string, r *Rules) (string, bool) {
	fp := filepath.ToSlash(filePath)
	for _, lr := range r.InternalLayerRules {
		if strings.HasPrefix(fp, filepath.ToSlash(lr.From)) {
			// deny takes precedence
			for _, d := range lr.DenyTo {
				if strings.HasPrefix(importPath, filepath.ToSlash(d)) {
					if lr.Message != "" {
						return lr.Message, true
					}
					return lr.Name, true
				}
			}
			// allow list (optional, se quiser ser estrito)
			if len(lr.AllowTo) > 0 && !matchPrefixAny(importPath, lr.AllowTo) {
				return "import não permitido por regra de camada: " + lr.Name, true
			}
		}
	}
	return "", false
}

var Analyzer = &analysis.Analyzer{
	Name: "depguardlite",
	Doc:  "Valida importações proibidas e regras de camadas (facades e internos)",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rules, err := loadRules()
		if err != nil {
			return nil, err
		}
		for _, f := range pass.Files {
			file := pass.Fset.Position(f.Pos()).Filename
			if isExcluded(file, rules.ExcludePaths) {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				imp, ok := n.(*ast.ImportSpec)
				if !ok || imp.Path == nil {
					return true
				}
				ip := strings.Trim(imp.Path.Value, `"`)
				// Mapa de deny explícito
				for blocked, msg := range rules.Deny {
					if ip == blocked || strings.HasPrefix(ip, blocked) {
						pass.Reportf(imp.Pos(), "import proibido: %s (%s)", ip, msg)
						return true
					}
				}
				// Regras de camadas internas (por caminho do arquivo/import)
				if msg, violated := checkLayerRule(file, ip, rules); violated {
					pass.Reportf(imp.Pos(), "violação de camada: %s → %s (%s)", file, ip, msg)
				}
				return true
			})
		}
		return nil, nil
	},
}

4.3 Entrypoint (vettool)

cmd/depguard-lite/main.go:

package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"your/module/path/internal/analyzers/depguardlite"
)

func main() {
	singlechecker.Main(depguardlite.Analyzer)
}


Substitua your/module/path pelo módulo do go.mod.

4.4 “Pin” de ferramentas (evita drift no CI)

internal/tools/vettools.go:

//go:build tools
// +build tools

package tools

import (
	_ "golang.org/x/tools/go/analysis"
	_ "golang.org/x/tools/go/analysis/singlechecker"
)

4.5 Makefile (alvo de vettool)
VETTOOL := $(PWD)/vettools/depguard-lite

.PHONY: vettool build-vet vet-dep
vettool:
	go build -o vettools/depguard-lite ./cmd/depguard-lite

vet-dep: vettool
	go vet -vettool=$(VETTOOL) ./...

4.6 Pipeline de CI (ordem recomendada)

go mod tidy && go mod verify (evita erros de goanalysis_metalinter e “no export data”)

audit-modules

audit-depguard

golangci-lint run (com gomodguard, unused, staticcheck etc.)

go run go-import-checks (se quiser regras de camada declarativas)

go vet -vettool=./vettools/depguard-lite (regras nativas e mensagens customizadas)

5) Checklist de migração (rápido)

 Rodar go mod tidy && go mod verify (elimina missing go.sum)

audit-modules

audit-vet

 Desativar depguard e linters obsoletos (deadcode, structcheck, varcheck)

audit-unused

audit-unused

 Ativar gomodguard com denylist/allowlist e exceções para facades

 Adicionar go-import-checks (opcional) para camadas internas

 Adicionar depguard-lite (vettool) com dep_rules.json

 Colocar jobs separados no CI, na ordem acima

 Documentar no README “como corrigir violações” (mensagens do vettool já apontam o facade correto)

6) Políticas e mensagens (DX)

Mensagens de erro sempre dizem o que usar:

“import proibido: github.com/go-chi/chi/v5 (Use o facade pkg/httpx)”

Exceções automáticas para pkg/*x (facades) previnem o paradoxo diagnosticado em depguard

linting_loop_resolution

.

Em caso de biblioteca nova, o fluxo padrão é:

criar pkg/<facade>x mínimo;

adicionar exceção desse pkg/<facade>x em gomodguard e dep_rules.json;

negar o módulo externo diretamente no código da aplicação.

7) Resultado esperado

Curto prazo: CI volta a verde; sem travas do goanalysis_metalinter; sem missing go.sum

audit-depguard

audit-modules

.

Médio prazo: depguard-lite garante import rules & camadas com desempenho, mensagens ricas e zero YAML frágil.

Longo prazo: arquitetura do mcp-ultra-wasm permanece limpa à medida que novas libs/facades surgirem, sem repetir o loop que já vimos nos relatórios e docs

linting_loop_resolution

.

quer que eu já gere os arquivos (código do depguard-lite, dep_rules.json, .golangci.yml, Makefile e scripts ci/), prontos para colar nessa árvore?

Fontes
Você disse:
sim
O ChatGPT disse:
Pensou por 5s

perfeito! gerei um pacote copiar-e-colar com todos os arquivos: .golangci.yml, Makefile, scripts de CI, depguard-lite (vettool Go) e configs. Eles seguem exatamente o blueprint que combinamos e já embutem as decisões baseadas nos seus logs/auditorias (depguard travando no goanalysis, go.sum ausente, linters obsoletos)

audit-depguard

audit-modules

audit-unused

audit-unused

, bem como a lição do paradoxo dos facades registrada nos docs técnicos

linting_loop_resolution

linting_loop_resolution

.

Abaixo estão os arquivos com caminhos relativos a:

E:\vertikon\business\SaaS\templates\mcp-ultra-wasm

1) .golangci.yml — troca Depguard → Gomodguard, remove linters obsoletos
run:
  timeout: 5m
  tests: true
  skip-files:
    - ".*_gen\\.go$"

linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused        # substitui deadcode/structcheck/varcheck
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    # Facades: não aplicar guardas de módulo dentro deles (evita o paradoxo observado)
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]


Motivações: remover linters deprecated/inativados reportados nos seus logs

audit-unused

audit-unused

 e sair do loop de análise/goanalysis do depguard

audit-depguard

.

2) Makefile — targets de lint e vettool

Makefile

SHELL := /bin/bash
GOLANGCI ?= golangci-lint
VETTOOL  := $(PWD)/vettools/depguard-lite

.PHONY: tidy verify lint vettool vet-dep ci

tidy:
	go mod tidy

verify:
	go mod verify

lint: tidy verify
	$(GOLANGCI) run --timeout=5m

vettool:
	go build -o vettools/depguard-lite ./cmd/depguard-lite

vet-dep: vettool
	go vet -vettool=$(VETTOOL) ./...

ci: tidy verify lint vet-dep
	@echo "CI OK"

3) Scripts de CI

ci/lint.sh

#!/usr/bin/env bash
set -euo pipefail
go mod tidy
go mod verify
golangci-lint run --timeout=5m
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


ci/lint.ps1

$ErrorActionPreference = "Stop"
go mod tidy
go mod verify
golangci-lint run --timeout=5m
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool="$(Get-Location)\vettools\depguard-lite" ./...


Ordem garante go.sum consistente antes da análise (conserta os “missing go.sum entry” que seus audits mostraram)

audit-modules

audit-vet

.

4) Config de regras de arquitetura (opcional, declarativa)

internal/config/import-checks.yml

rules:
  - name: "handlers-para-usecase-ou-domain"
    allow:
      from: "internal/handlers/.*"
      to:   "internal/(usecase|domain)/.*"
    deny:
      to:   "internal/infra/.*"
      message: "handlers não pode importar infra; use um usecase/domain"

  - name: "service-sem-dep-direta-em-infra"
    allow:
      from: "internal/service/.*"
      to:   "internal/(usecase|domain|repository)/.*"
    deny:
      to:   "internal/infra/.*"
      message: "service não deve acessar infra diretamente"


Execução (job separado opcional):

go run github.com/quantumcycle/go-import-checks@latest \
  -config internal/config/import-checks.yml ./...

5) Regras do vettool nativo (JSON)

internal/config/dep_rules.json

{
  "deny": {
    "github.com/go-chi/chi/v5": "Use o facade pkg/httpx",
    "github.com/go-chi/chi/v5/middleware": "Use o facade pkg/httpx",
    "github.com/redis/go-redis/v9": "Use o facade pkg/redisx",
    "github.com/prometheus/client_golang/prometheus": "Use o facade pkg/metrics",
    "github.com/prometheus/client_golang/prometheus/promauto": "Use o facade pkg/metrics",
    "go.opentelemetry.io/otel": "Use o facade pkg/observability",
    "go.opentelemetry.io/otel/": "Use o facade pkg/observability"
  },
  "excludePaths": [
    "pkg/httpx",
    "pkg/redisx",
    "pkg/metrics",
    "pkg/observability"
  ],
  "internalLayerRules": [
    {
      "name": "handlers->(usecase|domain) only",
      "from": "internal/handlers/",
      "allowTo": ["internal/usecase/", "internal/domain/"],
      "denyTo": ["internal/infra/"],
      "message": "handlers não pode importar infra; use usecase/domain"
    },
    {
      "name": "service no infra direto",
      "from": "internal/service/",
      "allowTo": ["internal/domain/", "internal/repository/", "internal/usecase/"],
      "denyTo": ["internal/infra/"],
      "message": "service não deve acessar infra diretamente"
    }
  ]
}

6) Vettool Go nativo — depguard-lite

internal/analyzers/depguardlite/analyzer.go

package depguardlite

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type Rules struct {
	Deny               map[string]string `json:"deny"`
	ExcludePaths       []string          `json:"excludePaths"`
	InternalLayerRules []struct {
		Name    string   `json:"name"`
		From    string   `json:"from"`
		AllowTo []string `json:"allowTo"`
		DenyTo  []string `json:"denyTo"`
		Message string   `json:"message"`
	} `json:"internalLayerRules"`
}

func loadRules() (*Rules, error) {
	cfgPath := filepath.ToSlash("internal/config/dep_rules.json")
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var r Rules
	return &r, json.Unmarshal(b, &r)
}

func isExcluded(filePath string, excludes []string) bool {
	fp := filepath.ToSlash(filePath)
	for _, e := range excludes {
		if strings.Contains(fp, filepath.ToSlash(e)) {
			return true
		}
	}
	return false
}

func matchAnyPrefix(s string, prefixes []string) bool {
	s = filepath.ToSlash(s)
	for _, p := range prefixes {
		if strings.HasPrefix(s, filepath.ToSlash(p)) {
			return true
		}
	}
	return false
}

func checkLayerRule(filePath, importPath string, r *Rules) (string, bool) {
	fp := filepath.ToSlash(filePath)
	for _, lr := range r.InternalLayerRules {
		// Se o arquivo pertence ao "from" da regra
		if strings.HasPrefix(fp, filepath.ToSlash(lr.From)) {
			// Negações primeiro
			for _, d := range lr.DenyTo {
				if strings.HasPrefix(importPath, filepath.ToSlash(d)) {
					if lr.Message != "" {
						return lr.Message, true
					}
					return "import não permitido por regra de camada: " + lr.Name, true
				}
			}
			// Se existir allow-list, restringe
			if len(lr.AllowTo) > 0 && !matchAnyPrefix(importPath, lr.AllowTo) {
				return "import não permitido (apenas " + strings.Join(lr.AllowTo, ", ") + ")", true
			}
		}
	}
	return "", false
}

var Analyzer = &analysis.Analyzer{
	Name: "depguardlite",
	Doc:  "Valida imports proibidos e regras de camadas (facades + internos)",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		rules, err := loadRules()
		if err != nil {
			return nil, err
		}

		for _, f := range pass.Files {
			file := pass.Fset.Position(f.Pos()).Filename
			if file == "" || isExcluded(file, rules.ExcludePaths) {
				continue
			}
			ast.Inspect(f, func(n ast.Node) bool {
				imp, ok := n.(*ast.ImportSpec)
				if !ok || imp.Path == nil {
					return true
				}
				ip := strings.Trim(imp.Path.Value, `"`)

				// Denylist explícita por módulo/pacote
				for blocked, msg := range rules.Deny {
					if ip == blocked || strings.HasPrefix(ip, blocked) {
						pass.Reportf(imp.Pos(), "import proibido: %s (%s)", ip, msg)
						return true
					}
				}

				// Regras de camadas internas por caminho do arquivo e do import
				if msg, violated := checkLayerRule(file, ip, rules); violated {
					pass.Reportf(imp.Pos(), "violação de camada: %s → %s (%s)", file, ip, msg)
				}
				return true
			})
		}
		return nil, nil
	},
}


cmd/depguard-lite/main.go

package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	// ajuste este import para o seu módulo (go.mod)
	"mcp-ultra-wasm/internal/analyzers/depguardlite"
)

func main() {
	singlechecker.Main(depguardlite.Analyzer)
}


🔧 Troque mcp-ultra-wasm/internal/... pelo module do seu go.mod (ex.: github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm).

7) “Pin” das dependências de tooling

internal/tools/vettools.go

//go:build tools
// +build tools

package tools

import (
	_ "golang.org/x/tools/go/analysis"
	_ "golang.org/x/tools/go/analysis/singlechecker"
)

8) Como rodar (local e CI)
# 1) Garantir saúde do módulo (conserta go.sum e previne erros de export data)
go mod tidy && go mod verify   # (mitiga os erros exibidos nos seus audits):contentReference[oaicite:11]{index=11}

# 2) Lint padrão (com gomodguard)
golangci-lint run --timeout=5m

# 3) (Opcional) Regras de camadas declarativas
go run github.com/quantumcycle/go-import-checks@latest \
  -config internal/config/import-checks.yml ./...

# 4) Vettool nativo (depguard-lite)
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


No Windows (PowerShell):

go mod tidy; go mod verify
golangci-lint run --timeout=5m
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool="$(Get-Location)\vettools\depguard-lite" ./...

9) Observações finais

Este blueprint evita o paradoxo dos facades que causou o looping no depguard

linting_loop_resolution

.

A ordem tidy → verify → lint → vet previne erros como “no export data / missing go.sum” vistos nos seus logs

audit-depguard

audit-modules

.

unused substitui deadcode/structcheck/varcheck, que estão deprecados/inativos conforme seus audits

audit-unused

audit-unused

.

Se quiser, eu também adapto os imports do cmd/depguard-lite/main.go ao módulo exato do seu go.mod (é só me dizer qual é a linha module ...).

Fontes
Você disse:
linting_loop_resolution.md
Arquivo
agora que não usamos mais o depguard, este planejamento irá mudar?
O ChatGPT disse:
Pensou por 7s

Curta: não muda a estratégia, só simplifica etapas e critérios. Mantemos a trilha “substituto rápido → regras de arquitetura → vettool nativo”, mas sem tarefas de exceção/contorno do Depguard (o paradoxo dos facades foi a raiz do loop) 

linting_loop_resolution

.

Plano atualizado (enxuto)
Fase 1 — Estabilizar CI (hoje)

Removido: qualquer passo de configurar/“domar” Depguard.

Fica:

go mod tidy && go mod verify antes de lint (fecha buracos de go.sum).

Gomodguard ativo no .golangci.yml (deny/allow de módulos).

Linters vivos: unused, staticcheck, gosimple, govet, revive.

DoD: pipeline verde; 0 violações de módulos; sem “no export data/missing go.sum”.

Fase 2 — Arquitetura entre camadas (opcional, declarativa)

Go-import-checks com regras handlers → (usecase|domain) e service → (domain|repo|usecase).

DoD: zero violações estruturais; mensagens de correção claras no PR.

Fase 3 — Vettool nativo (Depguard-lite)

Validar imports proibidos (apontando o facade correto) e camadas internas — tudo em Go/go/analysis.

DoD: go vet -vettool=./vettools/depguard-lite ./... bloqueia PR com mensagens explícitas; regras em JSON versionadas.

Diferenças vs. plano anterior
Tema	Antes	Agora
Linter de import	Depguard + exceções por pkg/*x	Sem Depguard; gomodguard (módulos) + vettool nativo (imports)
Risco de loop	Existia por “paradoxo do facade”	Eliminado (não há Depguard) 

linting_loop_resolution


Complexidade de YAML	Alta (exceções por pasta)	Menor (Gomodguard simples + regras JSON do vettool)
Mensagens de erro	Genéricas	Diretivas (“use pkg/httpx”, etc.)
Performance	Sensível em repositório grande	Melhor (analyzer direto e jobs separados)
Ordens no CI (definitivo)

go mod tidy && go mod verify

golangci-lint run (com Gomodguard + linters ativos)

(Opcional) go run go-import-checks …

go vet -vettool=./vettools/depguard-lite ./...

Aceite / Métricas

TTR-lint < 60s por job.

0 imports diretos de libs banidas no app (apenas via facades).

0 violações de camada nos diretórios internal/*.

SNR de lint alto: mensagens sempre oferecem a substituição/facade.

Riscos & rollback

Se go-import-checks gerar ruído em refactors grandes: desativar temporariamente o job 2 (mantendo vettool).

Se o vettool acusar falsos positivos em um pacote: adicionar excludePaths pontual no JSON e abrir tarefa de correção.

Próximos passos

Consolidar .golangci.yml com Gomodguard.

Garantir targets vettool/vet-dep no Makefile.

Versão inicial do depguard-lite já integrada como job do CI.

Se quiser, eu já ajusto o DoD/checklist no README e deixo os comandos prontos para colar no pipeline.

Fontes
Você disse:
sim 
O ChatGPT disse:

perfeito — já deixei o DoD + checklist e os comandos de pipeline prontos para você colar no README e no CI. Mantive a estratégia “Gomodguard → (opcional) Go-import-checks → Vettool nativo (depguard-lite)”, sem o antigo contorno do Depguard (paradoxo dos facades) e garantindo o pré-passo tidy/verify para eliminar os erros de go.sum e de análise que vimos nos seus logs.

Trecho pronto para o README.md
✅ DoD — Qualidade & Lint (v2, sem Depguard)

CI executa em ordem determinística:

go mod tidy && go mod verify

golangci-lint run (Gomodguard, unused, staticcheck, gosimple, govet, revive)

(Opcional) go-import-checks (regras de camadas)

go vet -vettool=./vettools/depguard-lite ./...

0 imports diretos de libs bloqueadas (usar facades pkg/*x).

0 violações de camadas em internal/* (handlers → usecase/domain; service → domain/repository/usecase).

Mensagens de erro sempre indicam a ação correta (“use pkg/httpx”, “use pkg/metrics”).

Tempo total de lint (TTR) ≤ 60s por job (baseline; ajustar conforme runner).

Notas históricas: removemos o Depguard por travar análise/goanalysis e por “paradoxo do facade”. Priorize tidy/verify antes do lint para não repetir “missing go.sum” / “no export data”.

🧾 Checklist de Pull Request (copiar/colar no template)

 Rodei make ci localmente (ou steps equivalentes) e está verde

 Sem import direto de módulos bloqueados (validado por Gomodguard)

 Sem violações de camada (validado por go-import-checks e/ou depguard-lite)

 Novos facades pkg/*x (se houver) adicionados a exceções do Gomodguard / dep_rules.json

 Commit inclui ajustes de go.mod/go.sum via go mod tidy

Snippets prontos de pipeline
🟩 GitHub Actions (.github/workflows/ci.yml)
name: CI
on:
  pull_request:
  push:
    branches: [ main ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22.x"

      - name: Tidy & Verify (fix go.sum drift)
        run: |
          go mod tidy
          go mod verify

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.60.1
          args: --timeout=5m

      # (Opcional) Regras declarativas de camadas
      - name: go-import-checks (architecture rules)
        run: |
          go run github.com/quantumcycle/go-import-checks@latest \
            -config internal/config/import-checks.yml ./...

      - name: Build vettool (depguard-lite)
        run: go build -o vettools/depguard-lite ./cmd/depguard-lite

      - name: Vet (depguard-lite)
        run: go vet -vettool=$(pwd)/vettools/depguard-lite ./...

🟨 GitLab CI (.gitlab-ci.yml)
stages: [lint]

lint:
  stage: lint
  image: golang:1.22
  script:
    - go mod tidy
    - go mod verify
    - go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.1
    - golangci-lint run --timeout=5m
    # opcional
    - go run github.com/quantumcycle/go-import-checks@latest -config internal/config/import-checks.yml ./...
    - go build -o vettools/depguard-lite ./cmd/depguard-lite
    - go vet -vettool=$(pwd)/vettools/depguard-lite ./...
  only:
    - merge_requests
    - main

Makefile (alvos resumidos)
SHELL := /bin/bash
GOLANGCI ?= golangci-lint
VETTOOL  := $(PWD)/vettools/depguard-lite

.PHONY: tidy verify lint vettool vet-dep ci

tidy:
	go mod tidy

verify:
	go mod verify

lint: tidy verify
	$(GOLANGCI) run --timeout=5m

vettool:
	go build -o vettools/depguard-lite ./cmd/depguard-lite

vet-dep: vettool
	go vet -vettool=$(VETTOOL) ./...

ci: tidy verify lint vet-dep
	@echo "CI OK"

Referências internas (por que isso funciona)

Paradoxo dos facades: Depguard bloqueava o próprio pkg/httpx e gerava loop; removido do fluxo.

“missing go.sum entry / no export data”: ordenar tidy/verify antes do lint elimina esses erros recorrentes.

Linters obsoletos (deadcode, structcheck, varcheck) → substituir por unused.

Se quiser, eu também já te mando um README completo com essas seções incorporadas + um badge de status do GitHub Actions.

Você disse:
quero
O ChatGPT disse:

aqui está o README.md completo, já formatado para o seu repositório
E:\vertikon\business\SaaS\templates\mcp-ultra-wasm.

🧬 MCP-Ultra — Padrão de Qualidade & Lint (v2)
🎯 Contexto

O Depguard foi removido do pipeline por gerar loops e travar o golangci-lint
(comportamento registrado nas auditorias audit-depguard.log e linting_loop_resolution.md).
Agora o lint do projeto usa ferramentas Go-nativas, rápidas e estáveis.

✅ DoD — “Definition of Done” para Qualidade e Lint
Etapa	Ferramenta	Objetivo
1️⃣	go mod tidy && go mod verify	Corrigir e validar dependências (go.sum)
2️⃣	golangci-lint run	Linters ativos: Gomodguard, unused, staticcheck, gosimple, govet, revive
3️⃣	(Opcional) go-import-checks	Garantir arquitetura (handlers → usecase/domain, service → domain/repo/usecase)
4️⃣	go vet -vettool=./vettools/depguard-lite ./...	Verificar imports proibidos e regras internas via vettool nativo

Critérios de aceitação

✅ Nenhum import direto de libs bloqueadas — apenas via pkg/*x (facades).

✅ Nenhuma violação de camada (internal/*).

✅ Mensagens de erro sempre sugerem a correção (“use pkg/httpx”).

✅ Tempo médio de lint ≤ 60 s.

🧩 Estrutura de lint atual
E:\vertikon\business\SaaS\templates\mcp-ultra-wasm\
├─ .golangci.yml              # Lint principal (Gomodguard + linters modernos)
├─ cmd/depguard-lite/         # Vettool nativo (go/analysis)
├─ internal/config/           # Regras JSON e YAML (imports e camadas)
├─ internal/analyzers/        # Código do depguard-lite
├─ ci/lint.sh / lint.ps1      # Scripts de CI
└─ Makefile                   # Targets tidy/verify/lint/vet

🧾 Checklist de Pull Request

 Rodei make ci localmente e está verde

 Sem import direto de módulos bloqueados (Gomodguard)

 Sem violações de camada (go-import-checks / depguard-lite)

 Se criei novo pkg/*x, adicionei exceção no .golangci.yml e dep_rules.json

 Commit inclui go mod tidy atualizado

⚙️ Execução manual
# 1) Corrige dependências
go mod tidy && go mod verify

# 2) Lint completo
golangci-lint run --timeout=5m

# 3) (opcional) Regras de camadas
go run github.com/quantumcycle/go-import-checks@latest \
  -config internal/config/import-checks.yml ./...

# 4) Vettool nativo
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


Windows (PowerShell):

go mod tidy; go mod verify
golangci-lint run --timeout=5m
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool="$(Get-Location)\vettools\depguard-lite" ./...

🧱 CI padrão (GitHub Actions)
name: CI
on:
  push:
    branches: [ main ]
  pull_request:

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22.x"

      - name: Tidy & Verify
        run: |
          go mod tidy
          go mod verify

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.60.1
          args: --timeout=5m

      - name: (Opcional) Go-import-checks
        run: |
          go run github.com/quantumcycle/go-import-checks@latest \
            -config internal/config/import-checks.yml ./...

      - name: Vettool
        run: |
          go build -o vettools/depguard-lite ./cmd/depguard-lite
          go vet -vettool=$(pwd)/vettools/depguard-lite ./...


Badge de status:

![CI](https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/actions/workflows/ci.yml/badge.svg)

🔍 Linters ativos (golangci-lint)
Linter	Motivo
Gomodguard	controla módulos externos (substitui Depguard)
unused	detecta código não utilizado (substitui deadcode/structcheck/varcheck)
staticcheck	valida bugs e antipadrões comuns
gosimple	simplificações automáticas
govet	análises nativas do compilador
revive	regras de estilo e documentação
🧠 Histórico de decisão

Depguard removido por loop de lint e falso-positivo em pkg/httpx

Substituído por Gomodguard → mais simples e estável

Introduzido depguard-lite (vettool) para controle contextual 100 % Go nativo.

Ordem tidy → verify → lint → vet evita “missing go.sum” / “no export data”.

Linters deadcode, structcheck, varcheck removidos (deprecated).

quer que eu adicione o badge e este README automaticamente no seu diretório docs/mcp-ultra-wasm-v11.3-lint-clean/ (gerando o .md lá, pronto para versionar no próximo commit)?

Você disse:
qual procedimento ao encontrar o Depguard no arquivo?
O ChatGPT disse:

excelente pergunta — e crítica agora que o depguard saiu oficialmente da stack do mcp-ultra-wasm.

🧩 Procedimento padrão ao encontrar referências ao Depguard

Seja no .golangci.yml, nos scripts ou no histórico de documentação, siga esta sequência:

1️⃣ Identificar onde ele aparece

Procure nos seguintes locais:

grep -R "depguard" E:\vertikon\business\SaaS\templates\mcp-ultra-wasm


Os pontos comuns:

.golangci.yml

Makefile

scripts de CI (lint.sh, lint.ps1, .github/workflows/*)

docs/ (histórico de resolução de loop)

2️⃣ Se for configuração ativa → ⚠️ remover

Se aparecer em blocos como:

linters:
  enable:
    - depguard


ou

linters-settings:
  depguard:
    rules:
      main:
        deny:
          - pkg: "github.com/go-chi/chi/v5"


➡ Ação:

Remova completamente o linter depguard e seu bloco linters-settings.depguard.

Substitua por o bloco do Gomodguard que já está no blueprint novo.

enable:
  - gomodguard
linters-settings:
  gomodguard:
    blocked:
      modules:
        - github.com/go-chi/chi/v5:
            reason: "Use pkg/httpx"

3️⃣ Se for referência em comentário ou documentação → 🧹 atualizar

Exemplo:

# depguard exceptions for facades


➡ Substitua por:

# gomodguard exceptions (facades podem importar libs externas)


Ou, em docs técnicos:

“depguard travava no lint” → mantenha, mas marque como histórico resolvido:

> (Histórico: Depguard removido e substituído por Gomodguard + Vettool)

4️⃣ Se for script de lint → corrigir comandos

Exemplo antigo:

golangci-lint run --disable-all -E depguard


➡ Corrigir para:

golangci-lint run --disable-all -E gomodguard


ou apenas:

golangci-lint run


se já estiver ativado no YAML.

5️⃣ Se for arquivo Makefile ou CI → atualizar jobs

Remova etapas que invoquem depguard:

golangci-lint run --disable-all -E depguard


Substitua por:

$(GOLANGCI) run --timeout=5m


E garanta que o job de vettool (depguard-lite) esteja presente.

6️⃣ Se for importado no código Go → 🚨 apagar imediatamente

O depguard nunca deve ser importado no código de produção.
Qualquer linha como:

import "github.com/OpenPeeDeeP/depguard"


➡ Apague e valide com:

go mod tidy && go vet


para limpar dependências e entradas em go.sum.

7️⃣ Validar limpeza

Após editar:

make tidy verify lint vet-dep


Saída esperada:

[✔] No linter issues found
[✔] No references to depguard


ou manualmente:

grep -R "depguard" . | grep -v "depguard-lite"


Só deve restar o vettool nativo (depguard-lite).

✅ Resumo rápido
Caso encontrado	Ação
Em .golangci.yml	❌ Remover bloco depguard → ✅ Adicionar Gomodguard
Em Makefile/CI	❌ Substituir comandos que chamam depguard
Em comentário/doc	✏️ Atualizar texto para “Gomodguard” ou “vettool nativo”
Em import Go	🚨 Remover import e rodar go mod tidy
Em vettool custom	✅ Manter (é o depguard-lite)

Quer que eu gere um script PowerShell + Bash que faça essa varredura e limpe todas as referências antigas de depguard automaticamente (mantendo apenas o depguard-lite)?

Você disse:
sim
O ChatGPT disse:
Pensou por 18s

perfeito — segue um “kit de limpeza” com dois scripts (Bash e PowerShell) para:

varrer o repo,

remover o Depguard de YAML/CI/Makefile,

preservar o nosso depguard-lite,

e opcionalmente injetar o bloco do Gomodguard no .golangci.yml se ele não existir.

Inclui dry-run, backup automático e modo overwrite do .golangci.yml caso o patch falhe.

📁 Onde colocar

Sugestão de diretório:

E:\vertikon\business\SaaS\templates\mcp-ultra-wasm\tools\
  cleanup-depguard.sh
  cleanup-depguard.ps1

🐧 Bash — tools/cleanup-depguard.sh
#!/usr/bin/env bash
set -euo pipefail

# cleanup-depguard.sh
# Remove Depguard da stack (mantém depguard-lite), corrige CI/Makefile
# e garante Gomodguard no .golangci.yml.
#
# Uso:
#   bash tools/cleanup-depguard.sh [-n|--dry-run] [--overwrite-yml] [ROOT]
# Exemplo:
#   bash tools/cleanup-depguard.sh --overwrite-yml .
#
# Flags:
#   -n | --dry-run     : não grava; só mostra mudanças
#   --overwrite-yml    : se patch YAML falhar, sobrescreve .golangci.yml com modelo seguro

DRY_RUN=0
OVERWRITE_YML=0
ROOT="${PWD}"

while (( "$#" )); do
  case "$1" in
    -n|--dry-run) DRY_RUN=1; shift;;
    --overwrite-yml) OVERWRITE_YML=1; shift;;
    *) ROOT="$1"; shift;;
  esac
done

cd "$ROOT"

echo "==> Raiz do repositório: $ROOT"
echo "==> Dry-run: $DRY_RUN | Overwrite YAML: $OVERWRITE_YML"

# Helpers
backup_file() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  local ts
  ts="$(date +%Y%m%d-%H%M%S)"
  cp -f "$f" "$f.bak.$ts"
  echo "Backup -> $f.bak.$ts"
}

apply_sed() {
  local pattern="$1"; shift
  local file="$1"
  if [[ $DRY_RUN -eq 1 ]]; then
    echo "sed: $pattern => $file (dry-run)"
  else
    sed -i.bak "$pattern" "$file" && rm -f "$file.bak"
  fi
}

replace_in_file() {
  local find="$1"; local repl="$2"; local file="$3"
  if [[ -f "$file" ]]; then
    if [[ $DRY_RUN -eq 1 ]]; then
      echo "replace: $find -> $repl in $file (dry-run)"
    else
      perl -0777 -pe "s/$find/$repl/g" -i.bak "$file" && rm -f "$file.bak"
    fi
  fi
}

# 1) Encontrar referências ao depguard (exclui depguard-lite)
echo "==> Procurando referências a 'depguard' (excluindo depguard-lite)"
RG="grep -RIn --exclude-dir=.git --exclude-dir=vendor --exclude-dir=vettools --exclude=*.png --exclude=*.jpg --exclude=*.exe --exclude=*.dll --exclude=*.a"
$RG -E "depguard(?!-lite)" . || true

# 2) .golangci.yml — patch seguro
YML=".golangci.yml"
if [[ -f "$YML" ]]; then
  echo "==> Patch no $YML (remover depguard, manter/instalar gomodguard)"
  backup_file "$YML"

  # a) Remover habilitação do depguard em 'linters.enable'
  replace_in_file '\n(\s*-\s*depguard\s*\n)' '\n' "$YML"

  # b) Remover bloco 'linters-settings: depguard: ...' (qualquer indentação)
  #   - remove desde a linha 'depguard:' até a próxima linha que começar com um identificador do mesmo nível
  awk '
    BEGIN{skip=0; indent=""}
    {
      if (skip==0) {
        if ($0 ~ /^[[:space:]]*depguard:[[:space:]]*$/) {
          skip=1
          match($0, /^([[:space:]]*)depguard:/, m)
          indent=m[1]
          next
        } else {
          print $0
        }
      } else {
        # fim do bloco: mesma indentação e algo tipo "^[indent][a-z]" ou "^[indent][A-Za-z_-]+:"
        if ($0 ~ "^" indent "[A-Za-z_-]+:") {
          skip=0
          print $0
        } else {
          next
        }
      }
    }
  ' "$YML" > "$YML.tmp" && mv "$YML.tmp" "$YML"

  # c) Garantir presença de gomodguard
  if ! grep -q 'gomodguard' "$YML"; then
    echo "==> Inserindo bloco de Gomodguard (não estava presente)"
    cat >> "$YML" <<'EOF'

linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]
EOF
  fi
else
  echo "==> $YML não encontrado. (ok)"
fi

# 3) Scripts/CI/Makefile — trocar chamadas específicas do depguard
echo "==> Atualizando scripts/CI/Makefile"
# Remover invocações diretas -E depguard e trocar por gomodguard ou run simples
for f in $(git ls-files | grep -E '(\.ya?ml|\.yml|Makefile|\.mk|\.ps1|\.sh)$'); do
  [[ -f "$f" ]] || continue
  # pular binários/editors
  case "$f" in
    *.png|*.jpg|*.jpeg|*.gif) continue;;
  esac
  # substituições
  replace_in_file '--disable-all\s+-E\s+depguard' ' ' "$f"
  replace_in_file '-E\s+depguard' '-E gomodguard' "$f"
  replace_in_file 'depguard' 'gomodguard' "$f"
  # preservar depguard-lite
  replace_in_file 'gomodguard-lite' 'depguard-lite' "$f"
done

# 4) Código Go — garantir que não há import de depguard
echo "==> Checando imports Go de depguard (não deve existir)"
if grep -RIn --include="*.go" -E '\".*depguard.*\"' . | grep -v 'depguard-lite' ; then
  echo "⚠️  Encontrado import de depguard em código Go. Remova manualmente."
fi

# 5) Overwrite YAML (opcional) — se ainda sobrou lixo do depguard
if [[ $OVERWRITE_YML -eq 1 ]]; then
  if grep -RIn -E "depguard(?!-lite)" "$YML" >/dev/null 2>&1 ; then
    echo "==> Overwrite do .golangci.yml (modelo seguro)"
    backup_file "$YML"
    cat > "$YML" <<'EOF'
run:
  timeout: 5m
  tests: true
  skip-files:
    - ".*_gen\\.go$"

linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]
EOF
  fi
fi

echo "==> Limpando refs residuais (exclui depguard-lite)"
$RG -E "depguard(?!-lite)" . || echo "(nenhuma referência encontrada)"

echo "==> DONE."

🪟 PowerShell — tools/cleanup-depguard.ps1
# cleanup-depguard.ps1
# Remove Depguard (mantém depguard-lite), corrige CI/Makefile e garante Gomodguard no .golangci.yml.
# Uso:
#   pwsh tools/cleanup-depguard.ps1 [-DryRun] [-OverwriteYml] [-Root <path>]

param(
  [switch]$DryRun = $false,
  [switch]$OverwriteYml = $false,
  [string]$Root = "."
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Push-Location $Root
Write-Host "==> Raiz do repositório: $((Get-Location).Path)"
Write-Host "==> Dry-run: $DryRun | Overwrite YAML: $OverwriteYml"

function Backup-File($Path) {
  if (Test-Path $Path) {
    $ts = Get-Date -Format "yyyyMMdd-HHmmss"
    Copy-Item $Path "$Path.bak.$ts" -Force
    Write-Host "Backup -> $Path.bak.$ts"
  }
}

# 1) Buscar referências a depguard (exclui depguard-lite)
Write-Host "==> Procurando referências a 'depguard' (excluindo depguard-lite)"
$hits = Get-ChildItem -Recurse -File -ErrorAction SilentlyContinue |
  Where-Object { $_.FullName -notmatch "\\.git\\|vendor\\|vettools" } |
  Where-Object { $_.Extension -notin @(".png",".jpg",".jpeg",".gif",".exe",".dll",".a") } |
  ForEach-Object {
    $matches = Select-String -Path $_.FullName -Pattern "depguard(?!-lite)" -SimpleMatch:$false
    if ($matches) { $matches }
  }
$hits | ForEach-Object { "$($_.Path):$($_.LineNumber): $($_.Line)" }

# 2) .golangci.yml patch
$yml = ".golangci.yml"
if (Test-Path $yml) {
  Write-Host "==> Patch no $yml"
  Backup-File $yml
  $content = Get-Content $yml -Raw

  # a) remover "- depguard"
  $content = $content -replace "(\r?\n\s*-\s*depguard\s*)", "`n"

  # b) Remover bloco "depguard:" em linters-settings (qualquer indentação)
  #   usa regex multiline para capturar desde "depguard:" até a próxima chave do mesmo nível
  $content = $content -replace "(?ms)^[ \t]*depguard:\s*\n(?:[ \t]+.+\n)*", ""

  # c) Garantir gomodguard presente (se não houver)
  if ($content -notmatch "gomodguard") {
    $gomod = @"
linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]
"@
    $content = $content + "`n" + $gomod
  }

  if (-not $DryRun) { Set-Content -Path $yml -Value $content -Encoding UTF8 }
  else { Write-Host "(dry-run) YAML atualizado em memória (não gravado)" }
} else {
  Write-Host "==> $yml não encontrado. (ok)"
}

# 3) Atualizar scripts/CI/Makefile
Write-Host "==> Atualizando scripts/CI/Makefile"
$files = git ls-files | Select-String -Pattern '\.ya?ml$|\.yml$|Makefile$|\.mk$|\.ps1$|\.sh$' | ForEach-Object { $_.Line }
foreach ($f in $files) {
  if (-not (Test-Path $f)) { continue }
  $txt = Get-Content $f -Raw
  $txt = $txt -replace "--disable-all\s+-E\s+depguard"," "
  $txt = $txt -replace "-E\s+depguard","-E gomodguard"
  $txt = $txt -replace "depguard","gomodguard"
  $txt = $txt -replace "gomodguard-lite","depguard-lite"  # reverte colisão no nome do nosso vettool
  if (-not $DryRun) { Set-Content -Path $f -Value $txt -Encoding UTF8 }
  else { Write-Host "(dry-run) Atualizado $f em memória" }
}

# 4) Imports Go
Write-Host "==> Checando imports Go de depguard"
$goHits = Get-ChildItem -Recurse -File -Include *.go |
  ForEach-Object { Select-String -Path $_.FullName -Pattern '\".*depguard.*\"' } |
  Where-Object { $_.Line -notmatch "depguard-lite" }
$goHits | ForEach-Object { "$($_.Path):$($_.LineNumber): $($_.Line)" }
if ($goHits) { Write-Warning "⚠️  Remova imports de depguard em código Go e rode 'go mod tidy'." }

# 5) Overwrite YAML se ainda sobrar "depguard"
if ($OverwriteYml -and (Select-String -Path $yml -Pattern "depguard(?!-lite)" -SimpleMatch:$false)) {
  Write-Host "==> Overwrite do .golangci.yml (modelo seguro)"
  Backup-File $yml
  $model = @"
run:
  timeout: 5m
  tests: true
  skip-files:
    - ".*_gen\\.go$"

linters:
  disable-all: true
  enable:
    - gomodguard
    - staticcheck
    - gosimple
    - unused
    - errcheck
    - govet
    - revive

linters-settings:
  gomodguard:
    allowed:
      modules:
        - std
        - github.com/jackc/pgx/v5
        - github.com/nats-io/nats.go
        - github.com/rs/zerolog
    blocked:
      modules:
        - github.com/sirupsen/logrus:
            reason: "Padronizar zerolog"
        - github.com/pkg/errors:
            reason: "Use errors nativo (Go 1.20+)"
        - github.com/go-chi/chi/v5:
            reason: "Use o facade pkg/httpx"
        - github.com/redis/go-redis/v9:
            reason: "Use o facade pkg/redisx"
        - github.com/prometheus/client_golang/prometheus:
            reason: "Use o facade pkg/metrics"
        - go.opentelemetry.io/otel:
            reason: "Use o facade pkg/observability"

issues:
  exclude-rules:
    - path: pkg/httpx/
      linters: [gomodguard]
    - path: pkg/redisx/
      linters: [gomodguard]
    - path: pkg/metrics/
      linters: [gomodguard]
    - path: pkg/observability/
      linters: [gomodguard]
"@
  if (-not $DryRun) { Set-Content -Path $yml -Value $model -Encoding UTF8 }
  else { Write-Host "(dry-run) overwrite preparado (não gravado)" }
}

# 6) Relatório final
Write-Host "==> Checando referências residuais (exclui depguard-lite)"
$final = Get-ChildItem -Recurse -File |
  ForEach-Object { Select-String -Path $_.FullName -Pattern "depguard(?!-lite)" -SimpleMatch:$false } 
if (-not $final) { Write-Host "(ok) Nenhuma referência residual encontrada." }
else { $final | ForEach-Object { "$($_.Path):$($_.LineNumber): $($_.Line)" } }

Pop-Location
Write-Host "==> DONE."

✅ Como usar
Bash (WSL/Git Bash)
cd E:\vertikon\business\SaaS\templates\mcp-ultra-wasm
bash tools/cleanup-depguard.sh --overwrite-yml .


Dry-run:

bash tools/cleanup-depguard.sh -n .

PowerShell
cd E:\vertikon\business\SaaS\templates\mcp-ultra-wasm
pwsh tools/cleanup-depguard.ps1 -OverwriteYml -Root .


Dry-run:

pwsh tools/cleanup-depguard.ps1 -DryRun -Root .

🔎 Depois da limpeza (ordem recomendada)
go mod tidy && go mod verify
golangci-lint run --timeout=5m
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


se aparecer algo sobre go.sum/export data, isso normalmente se resolve com o tidy/verify antes do lint (foi justamente o que vimos nos seus audits anteriores).

quer que eu também gere um commit message padrão e um PR template já cobrindo essa limpeza?

Você disse:
sim
O ChatGPT disse:

perfeito — aqui vai um commit message pronto (padrão Conventional Commits) e um PR template enxuto, já cobrindo checklists e validações do nosso pipeline novo.

✅ Commit message (copiar e colar)
chore(lint): remover Depguard; habilitar Gomodguard; adicionar vettool nativo (depguard-lite); ordenar CI (tidy→verify→lint→vet)

Contexto:
- Depguard removido por instabilidade e paradoxo com facades (pkg/*x), causando loops de lint e falhas de análise.
- Padronizamos verificação de módulos com Gomodguard e adicionamos depguard-lite (go/analysis) para regras contextuais e de camadas.
- CI agora executa go mod tidy/verify antes do lint/vet para evitar “missing go.sum” e “no export data”.

Mudanças:
- .golangci.yml: remove depguard e linters obsoletos; ativa Gomodguard + unused/staticcheck/gosimple/govet/revive.
- tools: adiciona scripts de limpeza (cleanup-depguard.{sh,ps1}).
- cmd/depguard-lite + internal/analyzers: adiciona vettool nativo com regras em internal/config/dep_rules.json.
- ci: scripts/ordem de execução (tidy → verify → golangci-lint → go-import-checks [opcional] → go vet -vettool).

Efeitos esperados:
- Lint previsível e rápido; mensagens de erro prescritivas (“use pkg/httpx/metrics/redisx/observability”).
- Elimina loops do depguard e erros de análise/go.sum.

Refs internas:
- Paradoxo/loop do Depguard e exceções em facades: ver docs de resolução. 
- Auditorias indicando “no export data / missing go.sum” e linters deprecated.

No-breaking-change


Se quiser amarrar a commit a evidências internas, pode colar os trechos dos relatórios que apontam o problema e a solução:
– loop/depguard e paradoxo dos facades: linting_loop_resolution.md
– erros de análise/go.sum: audit-vet.log, audit-modules.log
– linters deprecated: audit-unused.log

📝 Pull Request template (.github/pull_request_template.md)
## 🎯 Objetivo
Remover **Depguard**, estabilizar o lint com **Gomodguard** e introduzir o vettool nativo (**depguard-lite**) para impor regras de import/camadas no `mcp-ultra-wasm`.

---

## 📦 Mudanças
- `.golangci.yml`: remove `depguard` e linters obsoletos; ativa `gomodguard`, `unused`, `staticcheck`, `gosimple`, `govet`, `revive`.
- `cmd/depguard-lite/` + `internal/analyzers/depguardlite/`: adiciona vettool (go/analysis).
- `internal/config/dep_rules.json`: denylist e regras de camadas; exceções para `pkg/*x`.
- `ci/lint.sh` / `ci/lint.ps1`: ordem determinística (`tidy → verify → lint → vet`).
- `tools/cleanup-depguard.{sh,ps1}`: scripts para varrer e remover referências antigas ao Depguard.

---

## ✅ Checklist
- [ ] Rodei `go mod tidy && go mod verify`
- [ ] `golangci-lint run` **verde** (com `gomodguard`)
- [ ] *(opcional)* `go-import-checks` **verde** (`internal/` sem violações)
- [ ] `go build -o vettools/depguard-lite ./cmd/depguard-lite` OK
- [ ] `go vet -vettool=$(pwd)/vettools/depguard-lite ./...` **verde**
- [ ] Nenhum import direto de libs bloqueadas (usar `pkg/httpx`, `pkg/metrics`, `pkg/redisx`, `pkg/observability`)
- [ ] Se criei novo `pkg/*x`, adicionei exceções em `.golangci.yml` e `internal/config/dep_rules.json`

---

## 🔬 Testes/Validações
**Comandos rodados localmente:**
```bash
go mod tidy && go mod verify
golangci-lint run --timeout=5m
# opcional (arquitetura):
go run github.com/quantumcycle/go-import-checks@latest -config internal/config/import-checks.yml ./...
# vettool:
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...


Resultado esperado:

Lint ≤ 60s; mensagens prescritivas (“use pkg/httpx/…”); sem “missing go.sum”/“no export data”.

🧠 Contexto & Referências

Loop/Paradoxo Depguard (facades bloqueados pelo próprio linter) — ver registro de resolução.

Falhas de análise/go.sum (“no export data”, “missing go.sum entry”) — auditorias.

Linters deprecated (deadcode, structcheck, varcheck) — migrado para unused.

🧯 Risco & Rollback

Risco: ruído inicial ao aplicar regras de camadas (se habilitado).

Mitigação: desativar temporariamente o job de go-import-checks e manter o vettool.

Rollback: reverter .golangci.yml para um run básico e desligar o vettool no CI (sem apagar código).

📸 Evidências (logs/prints)

Cole aqui prints do CI verde e qualquer diff relevante (ex.: violações corrigidas).


---

Se quiser, preparo também um **template de commit lint-fix** (para PRs que só corrigem violações apontadas pelo novo pipeline) e um **template de issue** para solicitar inclusão de novos facades `pkg/*x` com os campos certos (nome, motivo, API mínima, denylist correspondente).
::contentReference[oaicite:8]{index=8}


O ChatGPT pode cometer erros.
