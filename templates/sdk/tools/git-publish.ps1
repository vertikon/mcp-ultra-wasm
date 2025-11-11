# ════════════════════════════════════════════════════════════════════════
# GIT PUBLISH - Publicação no GitHub
# ════════════════════════════════════════════════════════════════════════
# Executa sanitização e faz push para GitHub
# Versão: 1.0.0
# ════════════════════════════════════════════════════════════════════════

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
Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "║        📦 GIT PUBLISH - Publicação no GitHub                        ║" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Set-Location $SDK

# ────────────────────────────────────────────────────────────────────────
# FASE 1: SANITIZAÇÃO (se não for pulada)
# ────────────────────────────────────────────────────────────────────────

if (-not $SkipSanitize) {
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host "   FASE 1/5: Sanitização" -ForegroundColor Yellow
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host ""

    Write-Host "Executando publish-github.ps1..." -ForegroundColor White
    & "$SDK\tools\publish-github.ps1" -GitHubOrg $GitHubOrg -RepoName $RepoName -SkipPush

    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Erro na sanitização" -ForegroundColor Red
        exit 1
    }

    Write-Host "✅ Sanitização completa" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "⏭️  Sanitização pulada (--SkipSanitize)" -ForegroundColor Yellow
    Write-Host ""
}

# ────────────────────────────────────────────────────────────────────────
# FASE 2: VERIFICAR GIT
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 2/5: Verificar Git" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Verificar se é repositório git
if (-not (Test-Path ".git")) {
    Write-Host "Inicializando repositório Git..." -ForegroundColor White
    git init
    Write-Host "✅ Git inicializado" -ForegroundColor Green
} else {
    Write-Host "✅ Repositório Git já existe" -ForegroundColor Green
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 3: VERIFICAR README.md
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 3/5: Verificar README.md" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

if (-not (Test-Path "README.md")) {
    Write-Host "README.md não encontrado. Criando..." -ForegroundColor Yellow

    $readmeContent = @"
# sdk-ultra-wasm

**Versão:** 9.0.0
**Status:** ✅ ULTRA VERIFIED CERTIFIED

MCP-Ultra SDK Custom é um SDK Go completo para integração com o ecossistema MCP-Ultra.

## 🎯 Características

- ✅ 100% validado (Enhanced Validator V4)
- ✅ Integração automatizada com template mcp-ultra-wasm
- ✅ Endpoints HTTP para gerenciamento de seeds
- ✅ Scripts PowerShell para automação
- ✅ Auditoria e logging estruturado
- ✅ Prometheus metrics integrado
- ✅ Health checks completos
- ✅ Preparado para integração com orquestrador

## 🚀 Quick Start

``````bash
# Clone o repositório
git clone https://github.com/$GitHubOrg/$RepoName.git
cd $RepoName

# Configure variáveis de ambiente (opcional)
cp .env.example .env

# Download de dependências
go mod download

# Build
go build ./cmd

# Executar
./cmd
``````

Servidor iniciará em http://localhost:8080

## 📊 Endpoints

- `GET /health` - Health check
- `GET /healthz` - Liveness probe
- `GET /readyz` - Readiness probe
- `GET /metrics` - Prometheus metrics
- `POST /seed/sync` - Sincronizar seed
- `GET /seed/status` - Status da seed

## 📚 Documentação

- [CONFIG.md](CONFIG.md) - Configuração de variáveis de ambiente
- [docs/INTEGRACAO_TEMPLATE.md](docs/INTEGRACAO_TEMPLATE.md) - Integração SDK ↔ Template
- [docs/INTEGRACAO_ORQUESTRADOR.md](docs/INTEGRACAO_ORQUESTRADOR.md) - Integração com orquestrador
- [tools/README.md](tools/README.md) - Scripts de automação

## 🛠️ Desenvolvimento

``````bash
# Executar testes
go test ./...

# Executar com hot reload
go run ./cmd

# Scripts de desenvolvimento (PowerShell)
.\tools\seed-sync.ps1         # Sincronizar seed
.\tools\seed-run.ps1          # Executar SDK + Seed
.\tools\integracao-full.ps1   # Integração completa
``````

## 📦 Estrutura

``````
sdk-ultra-wasm/
├── cmd/                    # Entry point
├── internal/
│   ├── handlers/          # HTTP handlers
│   ├── seeds/             # Seed management
│   └── ...
├── pkg/
│   └── orchestrator/      # Tipos do orquestrador (stub)
├── tools/                 # Scripts de automação
├── docs/                  # Documentação
└── seeds/                 # Seeds locais (não versionadas)
``````

## 🔧 Configuração

Configure as seguintes variáveis de ambiente (ver [CONFIG.md](CONFIG.md)):

- `SEED_PATH` - Caminho da seed interna
- `SDK_PATH` - Caminho do SDK
- `FIX_PATH` - Caminho do mcp-ultra-wasm-fix
- `TEMPLATE_PATH` - Caminho do template mcp-ultra-wasm

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (\`git checkout -b feature/AmazingFeature\`)
3. Commit suas mudanças (\`git commit -m 'feat: Add some AmazingFeature'\`)
4. Push para a branch (\`git push origin feature/AmazingFeature\`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT - veja o arquivo LICENSE para detalhes.

## 🏆 Certificação

Este SDK foi certificado pelo **Enhanced Validator V4** com score de 100%.

**Certificado:** VTK-SDK-CUSTOM-V9-20251005-STABLE

Ver [docs/CERTIFICADO_VALIDACAO_V9.md](docs/CERTIFICADO_VALIDACAO_V9.md) para detalhes.

---

**Desenvolvido por:** Vertikon Team
**Última Atualização:** 2025-10-05
"@

    Set-Content -Path "README.md" -Value $readmeContent
    Write-Host "✅ README.md criado" -ForegroundColor Green
} else {
    Write-Host "✅ README.md já existe" -ForegroundColor Green
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 4: GIT ADD E COMMIT
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 4/5: Git Add e Commit" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Verificar branch
$currentBranch = git branch --show-current 2>$null
if ([string]::IsNullOrEmpty($currentBranch)) {
    Write-Host "Criando branch $Branch..." -ForegroundColor White
    git checkout -b $Branch
    Write-Host "✅ Branch $Branch criada" -ForegroundColor Green
} elseif ($currentBranch -ne $Branch) {
    Write-Host "Branch atual: $currentBranch" -ForegroundColor Yellow
    Write-Host "Mudando para branch $Branch..." -ForegroundColor White
    git checkout -b $Branch 2>$null
    if ($LASTEXITCODE -ne 0) {
        git checkout $Branch
    }
    Write-Host "✅ Branch $Branch" -ForegroundColor Green
} else {
    Write-Host "✅ Já na branch $Branch" -ForegroundColor Green
}

Write-Host ""

# Git add
Write-Host "Adicionando arquivos ao git..." -ForegroundColor White
git add .

$status = git status --short
if ([string]::IsNullOrEmpty($status)) {
    Write-Host "⚠️  Nenhuma alteração para commitar" -ForegroundColor Yellow
} else {
    Write-Host "✅ Arquivos adicionados:" -ForegroundColor Green
    git status --short | ForEach-Object { Write-Host "   $_" -ForegroundColor Gray }

    Write-Host ""
    Write-Host "Criando commit..." -ForegroundColor White

    $commitMessage = @"
feat: sanitize for GitHub publication (v9.0.0)

- Remove hardcoded internal paths
- Add environment variable configuration
- Create GitHub Actions CI workflow
- Add .gitignore for local files
- Create CONFIG.md documentation
- Update go.mod to github.com/$GitHubOrg/$RepoName

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
"@

    git commit -m $commitMessage
    Write-Host "✅ Commit criado" -ForegroundColor Green
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 5: GIT PUSH
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 5/5: Git Push" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

# Verificar se remote existe
$remoteUrl = git remote get-url origin 2>$null

if ([string]::IsNullOrEmpty($remoteUrl)) {
    Write-Host "Adicionando remote origin..." -ForegroundColor White
    git remote add origin "https://github.com/$GitHubOrg/$RepoName.git"
    Write-Host "✅ Remote adicionado: https://github.com/$GitHubOrg/$RepoName.git" -ForegroundColor Green
} else {
    Write-Host "✅ Remote já existe: $remoteUrl" -ForegroundColor Green

    # Verificar se a URL está correta
    $expectedUrl = "https://github.com/$GitHubOrg/$RepoName.git"
    if ($remoteUrl -ne $expectedUrl) {
        Write-Host "⚠️  URL do remote diferente da esperada" -ForegroundColor Yellow
        Write-Host "   Esperado: $expectedUrl" -ForegroundColor Gray
        Write-Host "   Atual: $remoteUrl" -ForegroundColor Gray

        if ($Force) {
            Write-Host "Atualizando remote..." -ForegroundColor White
            git remote set-url origin $expectedUrl
            Write-Host "✅ Remote atualizado" -ForegroundColor Green
        } else {
            Write-Host "   Use --Force para atualizar automaticamente" -ForegroundColor Yellow
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
    Write-Host "✅ Push realizado com sucesso!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🎉 Repositório publicado:" -ForegroundColor Cyan
    Write-Host "   https://github.com/$GitHubOrg/$RepoName" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "❌ Erro ao fazer push" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possíveis soluções:" -ForegroundColor Yellow
    Write-Host "   1. Verifique se o repositório existe no GitHub" -ForegroundColor White
    Write-Host "   2. Verifique suas credenciais do Git" -ForegroundColor White
    Write-Host "   3. Use --Force se necessário (sobrescrever remote)" -ForegroundColor White
    Write-Host ""
    exit 1
}

# ────────────────────────────────────────────────────────────────────────
# RELATÓRIO FINAL
# ────────────────────────────────────────────────────────────────────────

Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "║              ✅ PUBLICAÇÃO NO GITHUB COMPLETA                         ║" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

Write-Host "📊 Próximos Passos:" -ForegroundColor Cyan
Write-Host ""
Write-Host "   1. Acesse o repositório:" -ForegroundColor White
Write-Host "      https://github.com/$GitHubOrg/$RepoName" -ForegroundColor Gray
Write-Host ""
Write-Host "   2. Configure GitHub Settings:" -ForegroundColor White
Write-Host "      • Habilite GitHub Actions (CI)" -ForegroundColor Gray
Write-Host "      • Configure branch protection rules" -ForegroundColor Gray
Write-Host "      • Adicione description e topics" -ForegroundColor Gray
Write-Host ""
Write-Host "   3. Clone em outro ambiente para testar:" -ForegroundColor White
Write-Host "      git clone https://github.com/$GitHubOrg/$RepoName.git" -ForegroundColor Gray
Write-Host "      cd $RepoName" -ForegroundColor Gray
Write-Host "      go build ./cmd" -ForegroundColor Gray
Write-Host ""
Write-Host "   4. Configure variáveis de ambiente locais:" -ForegroundColor White
Write-Host "      cp .env.example .env" -ForegroundColor Gray
Write-Host "      # Edite .env com seus caminhos" -ForegroundColor Gray
Write-Host ""

Write-Host "📚 Documentação Disponível:" -ForegroundColor Cyan
Write-Host "   • README.md - Getting started" -ForegroundColor White
Write-Host "   • CONFIG.md - Configuração de ambiente" -ForegroundColor White
Write-Host "   • docs/ - Documentação completa" -ForegroundColor White
Write-Host ""

Write-Host "🎉 Parabéns! Seu SDK está publicado e pronto para uso!" -ForegroundColor Green
Write-Host ""
