# ════════════════════════════════════════════════════════════════════════
# PUBLISH GITHUB - Sanitização e Publicação Automatizada
# ════════════════════════════════════════════════════════════════════════
# Remove caminhos internos e prepara para publicação no GitHub
# Versão: 1.0.0
# ════════════════════════════════════════════════════════════════════════

param(
    [string]$GitHubOrg = "vertikon",
    [string]$RepoName = "sdk-ultra-wasm",
    [switch]$DryRun,
    [switch]$SkipPush,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
$LOGS_DIR = Join-Path $SDK "logs"
$TIMESTAMP = Get-Date -Format "yyyyMMdd-HHmmss"
$LOG_FILE = Join-Path $LOGS_DIR "publish-$TIMESTAMP.log"

# Criar diretório de logs
New-Item -ItemType Directory -Path $LOGS_DIR -Force | Out-Null

# Função de logging
function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "INFO"
    )
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] $Message"
    Add-Content -Path $LOG_FILE -Value $logEntry

    switch ($Level) {
        "ERROR" { Write-Host $Message -ForegroundColor Red }
        "WARN"  { Write-Host $Message -ForegroundColor Yellow }
        "SUCCESS" { Write-Host $Message -ForegroundColor Green }
        default { Write-Host $Message -ForegroundColor White }
    }
}

# Banner
Clear-Host
Write-Host ""
Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "║        📦 PUBLISH GITHUB - Sanitização Automatizada                ║" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

if ($DryRun) {
    Write-Host "⚠️  MODO DRY-RUN: Nenhuma alteração será feita" -ForegroundColor Yellow
    Write-Host ""
}

Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
Write-Log "INÍCIO DA SANITIZAÇÃO PARA GITHUB" "INFO"
Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
Write-Log "GitHub Org: $GitHubOrg" "INFO"
Write-Log "Repo Name: $RepoName" "INFO"
Write-Log "Dry Run: $DryRun" "INFO"
Write-Log "" "INFO"

Set-Location $SDK

# ────────────────────────────────────────────────────────────────────────
# FASE 1: BACKUP
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 1/7: Backup" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

$backupDir = Join-Path $SDK "backup-pre-publish-$TIMESTAMP"
Write-Log "Criando backup em: $backupDir" "INFO"

if (-not $DryRun) {
    # Backup de arquivos críticos
    New-Item -ItemType Directory -Path $backupDir -Force | Out-Null
    Copy-Item -Path (Join-Path $SDK "go.mod") -Destination $backupDir -Force
    Copy-Item -Path (Join-Path $SDK "internal") -Destination $backupDir -Recurse -Force
    Copy-Item -Path (Join-Path $SDK "tools") -Destination $backupDir -Recurse -Force
    Copy-Item -Path (Join-Path $SDK "docs") -Destination $backupDir -Recurse -Force

    Write-Log "✅ Backup criado com sucesso" "SUCCESS"
} else {
    Write-Log "⏭️  (DRY-RUN) Backup pulado" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 2: ATUALIZAR GO.MOD
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 2/7: Atualizar go.mod" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

Write-Log "Atualizando module path para github.com/$GitHubOrg/$RepoName" "INFO"

$gomodPath = Join-Path $SDK "go.mod"
$gomodContent = Get-Content $gomodPath -Raw

if (-not $DryRun) {
    # Atualizar module path
    $newGomodContent = $gomodContent -replace "module .*", "module github.com/$GitHubOrg/$RepoName"

    # Remover replace directives locais (manter apenas os necessários para desenvolvimento local)
    $lines = $newGomodContent -split "`n"
    $cleanedLines = @()
    $skipNext = $false

    foreach ($line in $lines) {
        if ($line -match "^replace .* => [E-Z]:\\") {
            Write-Log "  Removendo replace local: $($line.Trim())" "INFO"
            continue
        }
        $cleanedLines += $line
    }

    $finalContent = $cleanedLines -join "`n"
    Set-Content -Path $gomodPath -Value $finalContent -NoNewline

    Write-Log "✅ go.mod atualizado" "SUCCESS"
} else {
    Write-Log "⏭️  (DRY-RUN) go.mod não modificado" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 3: SANITIZAR ARQUIVOS GO
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 3/7: Sanitizar Arquivos Go" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

Write-Log "Sanitizando caminhos hardcoded em arquivos .go" "INFO"

$goFiles = Get-ChildItem -Path $SDK -Filter "*.go" -Recurse | Where-Object {
    $_.FullName -notmatch "\\backup-" -and
    $_.FullName -notmatch "\\seeds\\" -and
    $_.FullName -notmatch "\\.git\\"
}

$sanitizedCount = 0

foreach ($file in $goFiles) {
    $content = Get-Content $file.FullName -Raw
    $modified = $false

    # Substituir caminhos absolutos E:\ por variáveis de ambiente
    if ($content -match '[E-Z]:\\\\') {
        $newContent = $content

        # internal/seeds/manager.go - usar variáveis de ambiente
        if ($file.Name -eq "manager.go") {
            $newContent = $newContent -replace 'seedPath = `[E-Z]:\\[^`]+`', 'seedPath = getEnvOrDefault("SEED_PATH", filepath.Join(os.Getenv("HOME"), ".mcp-ultra-wasm-sdk", "seeds", "mcp-ultra-wasm"))'
            $newContent = $newContent -replace 'sdkPath  = `[E-Z]:\\[^`]+`', 'sdkPath  = getEnvOrDefault("SDK_PATH", ".")'
            $newContent = $newContent -replace 'fixPath  = `[E-Z]:\\[^`]+`', 'fixPath  = getEnvOrDefault("FIX_PATH", filepath.Join(os.Getenv("HOME"), ".mcp-ultra-wasm-fix"))'

            # Adicionar helper function se não existir
            if ($newContent -notmatch "func getEnvOrDefault") {
                $helperFunc = @"

// getEnvOrDefault retorna variável de ambiente ou valor padrão
func getEnvOrDefault(key, defaultValue string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return defaultValue
}
"@
                $newContent = $newContent -replace "(\nconst \()", "$helperFunc`n`$1"
            }

            $modified = $true
        }

        if ($modified -and -not $DryRun) {
            Set-Content -Path $file.FullName -Value $newContent -NoNewline
            Write-Log "  ✅ Sanitizado: $($file.FullName -replace [regex]::Escape($SDK), '.')" "INFO"
            $sanitizedCount++
        } elseif ($modified) {
            Write-Log "  (DRY-RUN) Seria sanitizado: $($file.FullName -replace [regex]::Escape($SDK), '.')" "WARN"
        }
    }
}

Write-Log "✅ $sanitizedCount arquivo(s) Go sanitizado(s)" "SUCCESS"
Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 4: SANITIZAR SCRIPTS POWERSHELL
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 4/7: Sanitizar Scripts PowerShell" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

Write-Log "Sanitizando caminhos em scripts .ps1" "INFO"

$ps1Files = Get-ChildItem -Path (Join-Path $SDK "tools") -Filter "*.ps1" | Where-Object {
    $_.Name -ne "publish-github.ps1"  # Não modificar este próprio script
}

$ps1SanitizedCount = 0

foreach ($file in $ps1Files) {
    $content = Get-Content $file.FullName -Raw
    $modified = $false

    if ($content -match '\$SDK = "E:\\') {
        $newContent = $content -replace '\$SDK = "E:\\[^"]+sdk-ultra-wasm"', '$SDK = $PSScriptRoot | Split-Path -Parent'
        $newContent = $newContent -replace '\$TPL = "E:\\[^"]+mcp-ultra-wasm"', '$TPL = $env:TEMPLATE_PATH ?? (Join-Path (Split-Path $SDK -Parent) "mcp-ultra-wasm")'
        $newContent = $newContent -replace '\$FIX = "E:\\[^"]+mcp-ultra-wasm-fix"', '$FIX = $env:FIX_PATH ?? (Join-Path $env:HOME ".mcp-ultra-wasm-fix")'

        $modified = $true

        if (-not $DryRun) {
            Set-Content -Path $file.FullName -Value $newContent -NoNewline
            Write-Log "  ✅ Sanitizado: $($file.Name)" "INFO"
            $ps1SanitizedCount++
        } else {
            Write-Log "  (DRY-RUN) Seria sanitizado: $($file.Name)" "WARN"
        }
    }
}

Write-Log "✅ $ps1SanitizedCount script(s) PowerShell sanitizado(s)" "SUCCESS"
Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 5: CRIAR/ATUALIZAR .GITIGNORE
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 5/7: Criar/Atualizar .gitignore" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

$gitignorePath = Join-Path $SDK ".gitignore"
$gitignoreContent = @"
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
cmd/cmd
cmd/cmd.exe

# Test binaries
*.test

# Output of go coverage tool
*.out

# Go workspace file (local only)
go.work
go.work.sum

# Seeds (local template copies)
seeds/
!seeds/.gitkeep

# Logs
logs/
*.log

# Backups
backup-*/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Environment
.env
.env.local

# Temporary files
tmp/
temp/
"@

if (-not $DryRun) {
    Set-Content -Path $gitignorePath -Value $gitignoreContent
    Write-Log "✅ .gitignore criado/atualizado" "SUCCESS"
} else {
    Write-Log "⏭️  (DRY-RUN) .gitignore não modificado" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 6: CRIAR GITHUB ACTIONS CI
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 6/7: Criar GitHub Actions CI" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

$workflowsDir = Join-Path $SDK ".github\workflows"
$ciWorkflowPath = Join-Path $workflowsDir "ci.yml"

$ciWorkflow = @"
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  build:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.25'

    - name: Install dependencies
      run: go mod download

    - name: Verify dependencies
      run: go mod verify

    - name: Build
      run: go build -v ./cmd

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.txt
        flags: unittests
        name: codecov-umbrella

    - name: Run go vet
      run: go vet ./...

    - name: Run staticcheck
      uses: dominikh/staticcheck-action@v1.3.0
      with:
        version: "latest"

  lint:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
"@

if (-not $DryRun) {
    New-Item -ItemType Directory -Path $workflowsDir -Force | Out-Null
    Set-Content -Path $ciWorkflowPath -Value $ciWorkflow
    Write-Log "✅ GitHub Actions CI workflow criado" "SUCCESS"
} else {
    Write-Log "⏭️  (DRY-RUN) GitHub Actions não criado" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 7: VALIDAR COMPILAÇÃO
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 7/7: Validar Compilação" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

if (-not $DryRun) {
    Write-Log "Executando go mod tidy..." "INFO"
    $tidyOutput = & go mod tidy 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ go mod tidy executado com sucesso" "SUCCESS"
    } else {
        Write-Log "❌ Erro em go mod tidy: $tidyOutput" "ERROR"
        exit 1
    }

    Write-Log "Compilando SDK..." "INFO"
    $buildOutput = & go build ./cmd 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ SDK compilado com sucesso" "SUCCESS"
    } else {
        Write-Log "❌ Erro ao compilar SDK: $buildOutput" "ERROR"
        exit 1
    }

    Write-Log "Executando testes..." "INFO"
    $testOutput = & go test ./... 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ Testes passaram" "SUCCESS"
    } else {
        Write-Log "⚠️  Alguns testes falharam: $testOutput" "WARN"
    }
} else {
    Write-Log "⏭️  (DRY-RUN) Validação pulada" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# CRIAR README DE CONFIGURAÇÃO
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   Criar README de Configuração" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

$configReadmePath = Join-Path $SDK "CONFIG.md"
$configReadme = @"
# Configuração de Ambiente

Este projeto foi sanitizado para publicação no GitHub. Os caminhos internos foram substituídos por variáveis de ambiente.

## Variáveis de Ambiente

Configure as seguintes variáveis de ambiente para desenvolvimento local:

### Windows (PowerShell)

``````powershell
`$env:SEED_PATH = "C:\path\to\seeds\mcp-ultra-wasm"
`$env:SDK_PATH = "C:\path\to\sdk-ultra-wasm"
`$env:FIX_PATH = "C:\path\to\mcp-ultra-wasm-fix"
`$env:TEMPLATE_PATH = "C:\path\to\mcp-ultra-wasm"
``````

### Linux/macOS (Bash)

``````bash
export SEED_PATH="/path/to/seeds/mcp-ultra-wasm"
export SDK_PATH="/path/to/sdk-ultra-wasm"
export FIX_PATH="/path/to/mcp-ultra-wasm-fix"
export TEMPLATE_PATH="/path/to/mcp-ultra-wasm"
``````

## Valores Padrão

Se as variáveis não forem configuradas, os seguintes padrões serão usados:

- **SEED_PATH**: `~/.mcp-ultra-wasm-sdk/seeds/mcp-ultra-wasm`
- **SDK_PATH**: `.` (diretório atual)
- **FIX_PATH**: `~/.mcp-ultra-wasm-fix`
- **TEMPLATE_PATH**: `../mcp-ultra-wasm` (diretório pai)

## Setup Inicial

``````bash
# 1. Clone o repositório
git clone https://github.com/$GitHubOrg/$RepoName.git
cd $RepoName

# 2. Configure variáveis de ambiente (opcional)
cp .env.example .env
# Edite .env com seus caminhos

# 3. Download de dependências
go mod download

# 4. Build
go build ./cmd

# 5. Executar
./cmd
``````

## Scripts de Desenvolvimento

Os scripts em `tools/` agora usam caminhos relativos ou variáveis de ambiente:

``````powershell
# Setup workspace (opcional, para desenvolvimento local)
.\tools\setup-go-work.ps1

# Sincronizar seed (requer TEMPLATE_PATH configurado)
.\tools\seed-sync.ps1

# Executar SDK
.\tools\seed-run.ps1

# Integração completa
.\tools\integracao-full.ps1
``````

## Desenvolvimento com go.work

Para desenvolvimento local multi-módulo, configure `go.work` manualmente:

``````
go 1.25

use (
    ./path/to/mcp-ultra-wasm-fix
    ./path/to/mcp-ultra-wasm
    ./path/to/sdk-ultra-wasm
)
``````

**Nota**: `go.work` é ignorado pelo git (`.gitignore`) para evitar caminhos locais no repositório.

## CI/CD

O workflow GitHub Actions (`.github/workflows/ci.yml`) compila e testa automaticamente em cada push/PR.
"@

if (-not $DryRun) {
    Set-Content -Path $configReadmePath -Value $configReadme
    Write-Log "✅ CONFIG.md criado" "SUCCESS"
} else {
    Write-Log "⏭️  (DRY-RUN) CONFIG.md não criado" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# RELATÓRIO FINAL
# ────────────────────────────────────────────────────────────────────────

Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "║              ✅ SANITIZAÇÃO PARA GITHUB COMPLETA                      ║" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

Write-Log "═══════════════════════════════════════════════════════════════════" "SUCCESS"
Write-Log "SANITIZAÇÃO FINALIZADA COM SUCESSO" "SUCCESS"
Write-Log "═══════════════════════════════════════════════════════════════════" "SUCCESS"

Write-Host "📊 Resumo da Sanitização:" -ForegroundColor Cyan
Write-Host ""
Write-Host "✅ Alterações Realizadas:" -ForegroundColor Green
if (-not $DryRun) {
    Write-Host "   • go.mod atualizado para github.com/$GitHubOrg/$RepoName" -ForegroundColor White
    Write-Host "   • $sanitizedCount arquivo(s) Go sanitizado(s)" -ForegroundColor White
    Write-Host "   • $ps1SanitizedCount script(s) PowerShell sanitizado(s)" -ForegroundColor White
    Write-Host "   • .gitignore criado/atualizado" -ForegroundColor White
    Write-Host "   • GitHub Actions CI workflow criado" -ForegroundColor White
    Write-Host "   • CONFIG.md criado" -ForegroundColor White
    Write-Host "   • Backup criado em: $backupDir" -ForegroundColor White
} else {
    Write-Host "   • (DRY-RUN) Nenhuma alteração real foi feita" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "📁 Arquivos Importantes:" -ForegroundColor Cyan
Write-Host "   • Log: $LOG_FILE" -ForegroundColor White
Write-Host "   • .gitignore: $gitignorePath" -ForegroundColor White
Write-Host "   • CI workflow: $ciWorkflowPath" -ForegroundColor White
Write-Host "   • CONFIG.md: $configReadmePath" -ForegroundColor White
Write-Host ""

if (-not $SkipPush -and -not $DryRun) {
    Write-Host "🚀 Próximos Passos - Git:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "   1. Revisar alterações:" -ForegroundColor White
    Write-Host "      git status" -ForegroundColor Gray
    Write-Host "      git diff" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   2. Adicionar e commitar:" -ForegroundColor White
    Write-Host "      git add ." -ForegroundColor Gray
    Write-Host "      git commit -m `"feat: sanitize for GitHub publication (v9.0.0)`"" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   3. Adicionar remote (se ainda não existe):" -ForegroundColor White
    Write-Host "      git remote add origin https://github.com/$GitHubOrg/$RepoName.git" -ForegroundColor Gray
    Write-Host ""
    Write-Host "   4. Push para GitHub:" -ForegroundColor White
    Write-Host "      git push -u origin main" -ForegroundColor Gray
    Write-Host ""
} elseif ($DryRun) {
    Write-Host "⚠️  DRY-RUN COMPLETO" -ForegroundColor Yellow
    Write-Host "   Execute novamente sem -DryRun para aplicar as alterações" -ForegroundColor White
    Write-Host ""
}

Write-Host "📚 Documentação:" -ForegroundColor Cyan
Write-Host "   • Leia CONFIG.md para configurar variáveis de ambiente" -ForegroundColor White
Write-Host "   • Revisar docs/INTEGRACAO_TEMPLATE.md (pode ter caminhos hardcoded)" -ForegroundColor White
Write-Host ""

Write-Log "Sanitização concluída em $(Get-Date)" "INFO"
Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
