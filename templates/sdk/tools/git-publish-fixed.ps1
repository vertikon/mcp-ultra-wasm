# ================================================================
# GIT PUBLISH - Publicacao no GitHub
# ================================================================
# Executa sanitizacao e faz push para GitHub
# Versao: 1.0.0
# ================================================================

param(
    [string]$GitHubOrg = "vertikon",
    [string]$RepoName = "sdk-ultra-wasm",
    [string]$Branch = "main",
    [switch]$SkipSanitize,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"

Clear-Host
Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "        GIT PUBLISH - Publicacao no GitHub                     " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

Set-Location $SDK

# ----------------------------------------------------------------
# FASE 1: SANITIZACAO (se nao for pulada)
# ----------------------------------------------------------------

if (-not $SkipSanitize) {
    Write-Host "================================================================" -ForegroundColor Yellow
    Write-Host "   FASE 1/5: Sanitizacao" -ForegroundColor Yellow
    Write-Host "================================================================" -ForegroundColor Yellow
    Write-Host ""

    Write-Host "Executando publish-github.ps1..." -ForegroundColor White
    & "$SDK\tools\publish-github.ps1" -GitHubOrg $GitHubOrg -RepoName $RepoName -SkipPush

    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERRO] Erro na sanitizacao" -ForegroundColor Red
        exit 1
    }

    Write-Host "[OK] Sanitizacao completa" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "[AVISO] Sanitizacao pulada (-SkipSanitize)" -ForegroundColor Yellow
    Write-Host ""
}

# ----------------------------------------------------------------
# FASE 2: VERIFICAR GIT
# ----------------------------------------------------------------

Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "   FASE 2/5: Verificar Git" -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""

# Verificar se e repositorio git
if (-not (Test-Path ".git")) {
    Write-Host "Inicializando repositorio Git..." -ForegroundColor White
    git init
    Write-Host "[OK] Git inicializado" -ForegroundColor Green
} else {
    Write-Host "[OK] Repositorio Git ja existe" -ForegroundColor Green
}

Write-Host ""

# ----------------------------------------------------------------
# FASE 3: VERIFICAR README.md
# ----------------------------------------------------------------

Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "   FASE 3/5: Verificar README.md" -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""

if (-not (Test-Path "README.md")) {
    Write-Host "README.md nao encontrado. Criando..." -ForegroundColor Yellow

    $readmeContent = @"
# sdk-ultra-wasm

**Versao:** 9.0.0
**Status:** ULTRA VERIFIED CERTIFIED

MCP-Ultra SDK Custom e um SDK Go completo para integracao com o ecossistema MCP-Ultra.

## Caracteristicas

- 100% validado (Enhanced Validator V4)
- Integracao automatizada com template mcp-ultra-wasm
- Endpoints HTTP para gerenciamento de seeds
- Scripts PowerShell para automacao
- Auditoria e logging estruturado
- Prometheus metrics integrado
- Health checks completos
- Preparado para integracao com orquestrador

## Quick Start

``````bash
# Clone o repositorio
git clone https://github.com/$GitHubOrg/$RepoName.git
cd $RepoName

# Configure variaveis de ambiente (opcional)
cp .env.example .env

# Download de dependencias
go mod download

# Build
go build ./cmd

# Executar
./cmd
``````

Servidor iniciara em http://localhost:8080

## Endpoints

- GET /health - Health check
- GET /healthz - Liveness probe
- GET /readyz - Readiness probe
- GET /metrics - Prometheus metrics
- POST /seed/sync - Sincronizar seed
- GET /seed/status - Status da seed

## Documentacao

- [CONFIG.md](CONFIG.md) - Configuracao de variaveis de ambiente
- [docs/INTEGRACAO_TEMPLATE.md](docs/INTEGRACAO_TEMPLATE.md) - Integracao SDK Template
- [docs/INTEGRACAO_ORQUESTRADOR.md](docs/INTEGRACAO_ORQUESTRADOR.md) - Integracao com orquestrador
- [tools/README.md](tools/README.md) - Scripts de automacao

## Desenvolvimento

``````bash
# Executar testes
go test ./...

# Executar com hot reload
go run ./cmd

# Scripts de desenvolvimento (PowerShell)
.\tools\seed-sync.ps1         # Sincronizar seed
.\tools\seed-run.ps1          # Executar SDK + Seed
.\tools\integracao-full.ps1   # Integracao completa
``````

## Estrutura

``````
sdk-ultra-wasm/
├── cmd/                    # Entry point
├── internal/
│   ├── handlers/          # HTTP handlers
│   ├── seeds/             # Seed management
│   └── ...
├── pkg/
│   └── orchestrator/      # Tipos do orquestrador (stub)
├── tools/                 # Scripts de automacao
├── docs/                  # Documentacao
└── seeds/                 # Seeds locais (nao versionadas)
``````

## Configuracao

Configure as seguintes variaveis de ambiente (ver [CONFIG.md](CONFIG.md)):

- SEED_PATH - Caminho da seed interna
- SDK_PATH - Caminho do SDK
- FIX_PATH - Caminho do mcp-ultra-wasm-fix
- TEMPLATE_PATH - Caminho do template mcp-ultra-wasm

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (git checkout -b feature/AmazingFeature)
3. Commit suas mudancas (git commit -m 'feat: Add some AmazingFeature')
4. Push para a branch (git push origin feature/AmazingFeature)
5. Abra um Pull Request

## Licenca

Este projeto esta sob a licenca MIT - veja o arquivo LICENSE para detalhes.

## Certificacao

Este SDK foi certificado pelo Enhanced Validator V4 com score de 100%.

Certificado: VTK-SDK-CUSTOM-V9-20251005-STABLE

Ver [docs/CERTIFICADO_VALIDACAO_V9.md](docs/CERTIFICADO_VALIDACAO_V9.md) para detalhes.

---

Desenvolvido por: Vertikon Team
Ultima Atualizacao: 2025-10-05
"@

    Set-Content -Path "README.md" -Value $readmeContent
    Write-Host "[OK] README.md criado" -ForegroundColor Green
} else {
    Write-Host "[OK] README.md ja existe" -ForegroundColor Green
}

Write-Host ""

# ----------------------------------------------------------------
# FASE 4: GIT ADD E COMMIT
# ----------------------------------------------------------------

Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "   FASE 4/5: Git Add e Commit" -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""

# Verificar branch
$currentBranch = git branch --show-current 2>$null
if ([string]::IsNullOrEmpty($currentBranch)) {
    Write-Host "Criando branch $Branch..." -ForegroundColor White
    git checkout -b $Branch
    Write-Host "[OK] Branch $Branch criada" -ForegroundColor Green
} elseif ($currentBranch -ne $Branch) {
    Write-Host "Branch atual: $currentBranch" -ForegroundColor Yellow
    Write-Host "Mudando para branch $Branch..." -ForegroundColor White
    git checkout -b $Branch 2>$null
    if ($LASTEXITCODE -ne 0) {
        git checkout $Branch
    }
    Write-Host "[OK] Branch $Branch" -ForegroundColor Green
} else {
    Write-Host "[OK] Ja na branch $Branch" -ForegroundColor Green
}

Write-Host ""

# Git add
Write-Host "Adicionando arquivos ao git..." -ForegroundColor White
git add .

$status = git status --short
if ([string]::IsNullOrEmpty($status)) {
    Write-Host "[AVISO] Nenhuma alteracao para commitar" -ForegroundColor Yellow
} else {
    Write-Host "[OK] Arquivos adicionados:" -ForegroundColor Green
    git status --short | ForEach-Object { Write-Host "   $_" -ForegroundColor Gray }

    Write-Host ""
    Write-Host "Criando commit..." -ForegroundColor White

    $commitMessage = "feat: sanitize for GitHub publication (v9.0.0)`n`n- Remove hardcoded internal paths`n- Add environment variable configuration`n- Create GitHub Actions CI workflow`n- Add .gitignore for local files`n- Create CONFIG.md documentation`n- Update go.mod to github.com/$GitHubOrg/$RepoName`n`nCo-Authored-By: Claude <noreply@anthropic.com>"

    git commit -m $commitMessage
    Write-Host "[OK] Commit criado" -ForegroundColor Green
}

Write-Host ""

# ----------------------------------------------------------------
# FASE 5: GIT PUSH
# ----------------------------------------------------------------

Write-Host "================================================================" -ForegroundColor Yellow
Write-Host "   FASE 5/5: Git Push" -ForegroundColor Yellow
Write-Host "================================================================" -ForegroundColor Yellow
Write-Host ""

# Verificar se remote existe
$remoteUrl = git remote get-url origin 2>$null

if ([string]::IsNullOrEmpty($remoteUrl)) {
    Write-Host "Adicionando remote origin..." -ForegroundColor White
    git remote add origin "https://github.com/$GitHubOrg/$RepoName.git"
    Write-Host "[OK] Remote adicionado: https://github.com/$GitHubOrg/$RepoName.git" -ForegroundColor Green
} else {
    Write-Host "[OK] Remote ja existe: $remoteUrl" -ForegroundColor Green

    # Verificar se a URL esta correta
    $expectedUrl = "https://github.com/$GitHubOrg/$RepoName.git"
    if ($remoteUrl -ne $expectedUrl) {
        Write-Host "[AVISO] URL do remote diferente da esperada" -ForegroundColor Yellow
        Write-Host "   Esperado: $expectedUrl" -ForegroundColor Gray
        Write-Host "   Atual: $remoteUrl" -ForegroundColor Gray

        if ($Force) {
            Write-Host "Atualizando remote..." -ForegroundColor White
            git remote set-url origin $expectedUrl
            Write-Host "[OK] Remote atualizado" -ForegroundColor Green
        } else {
            Write-Host "   Use -Force para atualizar automaticamente" -ForegroundColor Yellow
        }
    }
}

Write-Host ""

# Push
Write-Host "Fazendo push para GitHub..." -ForegroundColor White
Write-Host ""

$pushArgs = @("push", "-u", "origin", $Branch)
if ($Force) {
    $pushArgs += "--force"
}

& git $pushArgs

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "[OK] Push realizado com sucesso!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Repositorio publicado:" -ForegroundColor Cyan
    Write-Host "   https://github.com/$GitHubOrg/$RepoName" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "[ERRO] Erro ao fazer push" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possiveis solucoes:" -ForegroundColor Yellow
    Write-Host "   1. Verifique se o repositorio existe no GitHub" -ForegroundColor White
    Write-Host "   2. Verifique suas credenciais do Git" -ForegroundColor White
    Write-Host "   3. Use -Force se necessario (sobrescrever remote)" -ForegroundColor White
    Write-Host ""
    exit 1
}

# ----------------------------------------------------------------
# RELATORIO FINAL
# ----------------------------------------------------------------

Write-Host "================================================================" -ForegroundColor Green
Write-Host "              PUBLICACAO NO GITHUB COMPLETA                     " -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host ""

Write-Host "Proximos Passos:" -ForegroundColor Cyan
Write-Host ""
Write-Host "   1. Acesse o repositorio:" -ForegroundColor White
Write-Host "      https://github.com/$GitHubOrg/$RepoName" -ForegroundColor Gray
Write-Host ""
Write-Host "   2. Configure GitHub Settings:" -ForegroundColor White
Write-Host "      - Habilite GitHub Actions (CI)" -ForegroundColor Gray
Write-Host "      - Configure branch protection rules" -ForegroundColor Gray
Write-Host "      - Adicione description e topics" -ForegroundColor Gray
Write-Host ""
Write-Host "   3. Clone em outro ambiente para testar:" -ForegroundColor White
Write-Host "      git clone https://github.com/$GitHubOrg/$RepoName.git" -ForegroundColor Gray
Write-Host "      cd $RepoName" -ForegroundColor Gray
Write-Host "      go build ./cmd" -ForegroundColor Gray
Write-Host ""
Write-Host "   4. Configure variaveis de ambiente locais:" -ForegroundColor White
Write-Host "      cp .env.example .env" -ForegroundColor Gray
Write-Host "      # Edite .env com seus caminhos" -ForegroundColor Gray
Write-Host ""

Write-Host "Documentacao Disponivel:" -ForegroundColor Cyan
Write-Host "   - README.md - Getting started" -ForegroundColor White
Write-Host "   - CONFIG.md - Configuracao de ambiente" -ForegroundColor White
Write-Host "   - docs/ - Documentacao completa" -ForegroundColor White
Write-Host ""

Write-Host "Parabens! Seu SDK esta publicado e pronto para uso!" -ForegroundColor Green
Write-Host ""
