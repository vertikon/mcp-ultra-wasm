# ============================================================================
# prepare-template.ps1
# ============================================================================
# Prepara o template mcp-ultra-wasm com placeholders {{MODULE_PATH}}
# RODAR 1x NO TEMPLATE - Depois commitar no repo
# ============================================================================

Param(
  [string]$TemplateRoot = "E:\vertikon\business\SaaS\templates\mcp-ultra-wasm"
)

Write-Host "🔧 Preparando template mcp-ultra-wasm com placeholders..." -ForegroundColor Cyan
Write-Host "📁 Template: $TemplateRoot`n" -ForegroundColor Gray

Set-Location $TemplateRoot

# ============================================================================
# 1) Garantir estrutura de diretórios
# ============================================================================
Write-Host "📂 [1/5] Criando estrutura de diretórios..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path ".\scripts" | Out-Null
New-Item -ItemType Directory -Force -Path ".\test\mocks" | Out-Null
Write-Host "   ✅ Diretórios criados`n" -ForegroundColor Green

# ============================================================================
# 2) Substituir module no go.mod
# ============================================================================
Write-Host "📝 [2/5] Parametrizando go.mod..." -ForegroundColor Yellow

if (Test-Path ".\go.mod") {
    $gomod = Get-Content .\go.mod -Raw

    # Detectar módulo atual
    if ($gomod -match 'module\s+([^\s]+)') {
        $currentModule = $matches[1]
        Write-Host "   📦 Módulo atual: $currentModule" -ForegroundColor Gray
    }

    # Substituir por placeholder
    $gomod = $gomod -replace '^module\s+.+', 'module {{MODULE_PATH}}'
    Set-Content .\go.mod $gomod -NoNewline

    Write-Host "   ✅ go.mod parametrizado: module {{MODULE_PATH}}`n" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  go.mod não encontrado!" -ForegroundColor Red
}

# ============================================================================
# 3) Substituir imports nos arquivos .go
# ============================================================================
Write-Host "🔄 [3/5] Substituindo imports nos arquivos .go..." -ForegroundColor Yellow

$files = Get-ChildItem -Recurse -Include *.go -File | Where-Object {
  $_.FullName -notmatch '\\vendor\\' -and
  $_.FullName -notmatch '\\.git\\' -and
  $_.FullName -notmatch '\\node_modules\\'
}

$totalFiles = 0
$totalReplacements = 0

foreach ($f in $files) {
  $c = Get-Content $f.FullName -Raw
  $original = $c

  # Substituir imports base
  $c = $c -replace 'github\.com/vertikon/mcp-ultra-wasm/', '{{MODULE_PATH}}/'

  # Garantir que test/mocks aponte pro módulo local
  $c = $c -replace 'github\.com/vertikon/mcp-ultra-wasm/test/mocks', '{{MODULE_PATH}}/test/mocks'

  if ($c -ne $original) {
    Set-Content $f.FullName $c -NoNewline
    $totalFiles++
    $totalReplacements++
    Write-Host "   📝 $($f.FullName -replace [regex]::Escape($TemplateRoot), '.')" -ForegroundColor Cyan
  }
}

Write-Host "   ✅ $totalFiles arquivos corrigidos`n" -ForegroundColor Green

# ============================================================================
# 4) Criar esqueleto de test/mocks
# ============================================================================
Write-Host "🧪 [4/5] Configurando test/mocks..." -ForegroundColor Yellow

if (-not (Test-Path ".\test\mocks\README.md")) {
  $mockReadme = @'
# Test Mocks

Mocks locais para testes deste projeto.

## Uso com testify

```go
package mocks

import "github.com/stretchr/testify/mock"

type ExampleService struct {
    mock.Mock
}

func (m *ExampleService) DoSomething(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
```

## Uso com gomock

```bash
go install github.com/golang/mock/mockgen@latest
mockgen -source=internal/services/example.go -destination=test/mocks/example_mock.go -package=mocks
```

## Gerando mocks automaticamente

Adicione ao arquivo de interface:

```go
//go:generate mockgen -destination=../../test/mocks/example_mock.go -package=mocks . ExampleService
```

Depois rode:

```bash
go generate ./...
```
'@
  Set-Content ".\test\mocks\README.md" $mockReadme

  Write-Host "   ✅ README.md criado em test/mocks" -ForegroundColor Green
} else {
  Write-Host "   ℹ️  test/mocks/README.md já existe" -ForegroundColor Gray
}

# Criar .gitkeep para garantir que a pasta seja versionada
if (-not (Test-Path ".\test\mocks\.gitkeep")) {
    New-Item -ItemType File -Path ".\test\mocks\.gitkeep" -Force | Out-Null
}

Write-Host ""

# ============================================================================
# 5) Criar arquivo de documentação
# ============================================================================
Write-Host "📚 [5/5] Criando documentação do template..." -ForegroundColor Yellow

$docContent = @"
# MCP Ultra Template - Guia de Uso

## 🎯 Propósito

Este é o **template base** para criação de novos microserviços na arquitetura Vertikon.
Ele usa placeholders `{{MODULE_PATH}}` que são substituídos durante o bootstrap.

## 🚀 Criando um novo serviço (semente)

### 1. Clone o template

\`\`\`powershell
Copy-Item -Recurse -Force \`
  "E:\vertikon\business\SaaS\templates\mcp-ultra-wasm" \`
  "E:\vertikon\.ecosistema-vertikon\NeuraLead\waba\meu-servico"
\`\`\`

### 2. Execute o bootstrap

\`\`\`powershell
cd "E:\vertikon\.ecosistema-vertikon\NeuraLead\waba\meu-servico"
.\scripts\bootstrap.ps1 github.com/vertikon/meu-servico
\`\`\`

### 3. Valide

\`\`\`powershell
go mod tidy
go build ./...
go test ./...
\`\`\`

## 📦 Dependências

### ✅ Permitidas (via mcp-ultra-wasm-fix)

\`\`\`go
import (
  "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/logger"
  "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/version"
  "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/config"
  // ... outros pacotes do fix
)
\`\`\`

### ❌ Proibidas (privadas)

\`\`\`go
// NUNCA use:
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/..."
import "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/test/mocks"
\`\`\`

### ✅ Imports internos (após bootstrap)

\`\`\`go
import (
  "github.com/vertikon/meu-servico/internal/config"
  "github.com/vertikon/meu-servico/internal/handlers"
  "github.com/vertikon/meu-servico/test/mocks"
)
\`\`\`

## 🧪 Mocks

Cada projeto mantém seus próprios mocks em `test/mocks/`.

**Opção 1: testify**
\`\`\`bash
go get github.com/stretchr/testify/mock
\`\`\`

**Opção 2: gomock**
\`\`\`bash
go install github.com/golang/mock/mockgen@latest
go generate ./...
\`\`\`

## 🔧 Manutenção do Template

### Atualizar placeholders (raro)

\`\`\`powershell
.\scripts\prepare-template.ps1
\`\`\`

### Adicionar nova dependência compartilhada

1. Adicione ao `mcp-ultra-wasm-fix` (não ao template!)
2. Publique nova versão do fix
3. Use no template via import do fix

## ✅ Checklist de Qualidade

- [ ] `go mod tidy` sem erros
- [ ] `go build ./...` compila
- [ ] `go test ./...` passa
- [ ] Nenhum import para `github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/...`
- [ ] Apenas imports `{{MODULE_PATH}}/...` ou `mcp-ultra-wasm-fix/pkg/...`
- [ ] Mocks em `test/mocks/` (local)

## 📞 Suporte

Dúvidas? Consulte o time de arquitetura.
"@

Set-Content ".\TEMPLATE_GUIDE.md" $docContent

Write-Host "   ✅ TEMPLATE_GUIDE.md criado`n" -ForegroundColor Green

# ============================================================================
# Resumo Final
# ============================================================================
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✅ TEMPLATE PREPARADO COM SUCESSO!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "📊 Estatísticas:" -ForegroundColor Yellow
Write-Host "   • Arquivos .go corrigidos: $totalFiles" -ForegroundColor White
Write-Host "   • Substituições realizadas: $totalReplacements" -ForegroundColor White
Write-Host ""
Write-Host "📋 Próximos passos:" -ForegroundColor Yellow
Write-Host "   1. Revise as mudanças: git diff" -ForegroundColor White
Write-Host "   2. Commit o template:" -ForegroundColor White
Write-Host "      git add ." -ForegroundColor Gray
Write-Host "      git commit -m 'feat: parametrizar template com {{MODULE_PATH}}'" -ForegroundColor Gray
Write-Host "   3. Use scripts/bootstrap.ps1 nas sementes" -ForegroundColor White
Write-Host ""
Write-Host "📚 Documentação: TEMPLATE_GUIDE.md" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
