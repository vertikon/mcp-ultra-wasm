# MCP Ultra - Setup Completo
# Executa todos os passos necessários para configurar o ambiente MCP Ultra

param(
    [Parameter(Mandatory=$true)]
    [string]$GithubToken,
    
    [Parameter(Mandatory=$false)]
    [string]$GitUserName = "MCP Ultra User",
    
    [Parameter(Mandatory=$false)]
    [string]$GitUserEmail = "user@vertikon.com",
    
    [Parameter(Mandatory=$false)]
    [string]$Organization = "vertikon",
    
    [Parameter(Mandatory=$false)]
    [string]$DefaultRepo = "ecosystem"
)

Write-Host "🚀 MCP Ultra - Setup Completo" -ForegroundColor Green
Write-Host "=============================" -ForegroundColor Yellow
Write-Host "🔧 Template: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm" -ForegroundColor Cyan
Write-Host "" -ForegroundColor White

# 1. Verificar se Node.js está instalado
Write-Host "🔍 Verificando Node.js..." -ForegroundColor Cyan
$nodeVersion = ""
try {
    $nodeVersion = & node --version 2>$null
    Write-Host "✅ Node.js encontrado: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Node.js não encontrado. Instalando via Chocolatey..." -ForegroundColor Red
    
    # Instalar Chocolatey se não estiver instalado
    if (!(Get-Command choco -ErrorAction SilentlyContinue)) {
        Write-Host "📦 Instalando Chocolatey..." -ForegroundColor Yellow
        Set-ExecutionPolicy Bypass -Scope Process -Force
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
        iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    }
    
    # Instalar Node.js
    Write-Host "📦 Instalando Node.js..." -ForegroundColor Yellow
    choco install nodejs -y
    
    # Recarregar PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    
    try {
        $nodeVersion = & node --version
        Write-Host "✅ Node.js instalado com sucesso: $nodeVersion" -ForegroundColor Green
    } catch {
        Write-Host "❌ Falha ao instalar Node.js. Instale manualmente: https://nodejs.org" -ForegroundColor Red
        exit 1
    }
}

# 2. Navegar para diretório MCP server e instalar dependências
Write-Host "📦 Configurando MCP Server..." -ForegroundColor Cyan
$mcpServerPath = Join-Path $PSScriptRoot "..\mcp-server"

if (Test-Path $mcpServerPath) {
    Set-Location $mcpServerPath
    
    # Instalar dependências NPM
    try {
        Write-Host "📦 Instalando dependências NPM..." -ForegroundColor Yellow
        & npm install
        Write-Host "✅ Dependências NPM instaladas" -ForegroundColor Green
    } catch {
        Write-Host "❌ Falha ao instalar dependências NPM" -ForegroundColor Red
        exit 1
    }
    
    # Configurar arquivo .env
    Write-Host "⚙️ Configurando arquivo .env..." -ForegroundColor Cyan
    $envContent = @"
# MCP Ultra Server Configuration

# GitHub Personal Access Token
# Create one at: https://github.com/settings/tokens
# Required scopes: repo, read:org, read:user, workflow
GITHUB_TOKEN=$GithubToken

# GitHub Enterprise (optional)
# GITHUB_ENTERPRISE_URL=https://github.your-company.com

# Repository Configuration
GITHUB_ORG=$Organization
GITHUB_DEFAULT_REPO=$DefaultRepo

# MCP Server Configuration
MCP_SERVER_PORT=3100
MCP_SERVER_NAME=vertikon-mcp-ultra-wasm

# Logging
LOG_LEVEL=info

# Cache Configuration
CACHE_TTL=300
ENABLE_CACHE=true

# Rate Limiting
RATE_LIMIT_PER_HOUR=5000

# Features
ENABLE_ISSUES=true
ENABLE_PULL_REQUESTS=true
ENABLE_ACTIONS=true
ENABLE_RELEASES=true
ENABLE_DISCUSSIONS=true
ENABLE_PROJECTS=true

# Security
GITHUB_READ_ONLY=false

# Toolsets
GITHUB_TOOLSETS=context,repos,issues,pull_requests,actions,code_security,dependabot,discussions
"@

    Set-Content -Path ".env" -Value $envContent
    Write-Host "✅ Arquivo .env configurado" -ForegroundColor Green
    
    # Build do projeto TypeScript
    Write-Host "🏗️ Compilando projeto TypeScript..." -ForegroundColor Cyan
    try {
        & npm run build
        Write-Host "✅ MCP Server compilado com sucesso" -ForegroundColor Green
    } catch {
        Write-Host "❌ Falha ao compilar MCP Server" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "⚠️ Diretório mcp-server não encontrado" -ForegroundColor Yellow
}

# 3. Configurar Git global
Write-Host "🔧 Configurando Git..." -ForegroundColor Cyan
try {
    & git config --global user.name "$GitUserName"
    & git config --global user.email "$GitUserEmail"
    & git config --global init.defaultBranch main
    Write-Host "✅ Git configurado globalmente" -ForegroundColor Green
} catch {
    Write-Host "❌ Falha ao configurar Git" -ForegroundColor Red
}

# 4. Verificar se Go está instalado (para autocommit tool)
Write-Host "🔍 Verificando Go..." -ForegroundColor Cyan
try {
    $goVersion = & go version 2>$null
    Write-Host "✅ Go encontrado: $goVersion" -ForegroundColor Green
    
    # Compilar ferramenta de autocommit
    $automationPath = Join-Path $PSScriptRoot "..\automation"
    if (Test-Path $automationPath) {
        Set-Location $automationPath
        Write-Host "🔧 Compilando ferramenta de autocommit..." -ForegroundColor Yellow
        & go build -o autocommit.exe autocommit.go
        Write-Host "✅ Ferramenta de autocommit compilada" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Go não encontrado. Ferramenta de autocommit não será compilada." -ForegroundColor Yellow
    Write-Host "💡 Instale Go em: https://golang.org/dl/" -ForegroundColor Cyan
}

# 5. Testar conexão com GitHub
Write-Host "🔍 Testando conexão com GitHub..." -ForegroundColor Cyan
$env:GITHUB_TOKEN = $GithubToken

try {
    $headers = @{
        "Authorization" = "token $GithubToken"
        "User-Agent" = "MCP-Ultra-Setup"
        "Accept" = "application/vnd.github.v3+json"
    }
    
    $response = Invoke-RestMethod -Uri "https://api.github.com/user" -Headers $headers -Method GET
    Write-Host "✅ Conexão com GitHub bem-sucedida! Usuário: $($response.login)" -ForegroundColor Green
    
    # Testar acesso à organização
    try {
        $orgResponse = Invoke-RestMethod -Uri "https://api.github.com/orgs/$Organization" -Headers $headers -Method GET
        Write-Host "✅ Acesso à organização $Organization confirmado!" -ForegroundColor Green
    } catch {
        Write-Host "⚠️ Não foi possível acessar a organização $Organization. Verifique as permissões." -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "❌ Falha na conexão com GitHub. Verifique o token." -ForegroundColor Red
    Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
}

# 6. Testar servidor MCP
if (Test-Path "$mcpServerPath\dist\index.js") {
    Write-Host "🔍 Testando servidor MCP..." -ForegroundColor Cyan
    Set-Location $mcpServerPath
    
    try {
        $testJson = '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
        $result = $testJson | & node dist/index.js
        Write-Host "✅ Servidor MCP funcionando!" -ForegroundColor Green
        
        # Verificar se create_repository está disponível
        if ($result -match "create_repository") {
            Write-Host "✅ Ferramenta create_repository disponível!" -ForegroundColor Green
        } else {
            Write-Host "⚠️ Ferramenta create_repository não encontrada" -ForegroundColor Yellow
        }
        
    } catch {
        Write-Host "❌ Falha ao testar servidor MCP" -ForegroundColor Red
        Write-Host "Erro: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# 7. Criar script de inicialização rápida
$quickStartPath = Join-Path $PSScriptRoot "quick-start.ps1"
$quickStartContent = @"
# MCP Ultra - Quick Start
# Inicia o servidor MCP rapidamente

Write-Host "🚀 Iniciando MCP Ultra Server..." -ForegroundColor Green

Set-Location "$mcpServerPath"
& node dist/index.js
"@

Set-Content -Path $quickStartPath -Value $quickStartContent
Write-Host "✅ Script quick-start.ps1 criado" -ForegroundColor Green

Write-Host "" -ForegroundColor White
Write-Host "=================================" -ForegroundColor Yellow
Write-Host "🎉 MCP Ultra Setup Completo!" -ForegroundColor Green
Write-Host "=================================" -ForegroundColor Yellow
Write-Host "" -ForegroundColor White

Write-Host "📋 Resumo da Configuração:" -ForegroundColor Cyan
Write-Host "✅ Node.js: $nodeVersion" -ForegroundColor Green
Write-Host "✅ MCP Server: Compilado e configurado" -ForegroundColor Green
Write-Host "✅ GitHub Token: Configurado" -ForegroundColor Green
Write-Host "✅ Git: Configurado globalmente" -ForegroundColor Green
if (Test-Path "$automationPath\autocommit.exe") {
    Write-Host "✅ AutoCommit Tool: Compilado" -ForegroundColor Green
} else {
    Write-Host "⚠️ AutoCommit Tool: Requer Go" -ForegroundColor Yellow
}

Write-Host "" -ForegroundColor White
Write-Host "🚀 Próximos passos:" -ForegroundColor Cyan
Write-Host "1. Execute: .\scripts\quick-start.ps1 (para iniciar MCP Server)" -ForegroundColor White
Write-Host "2. Execute: .\testing\test-complete-pipeline.ps1 (para testar)" -ForegroundColor White
Write-Host "3. Use as ferramentas MCP:" -ForegroundColor White
Write-Host "   - create_repository: Criar repositórios GitHub" -ForegroundColor White
Write-Host "   - create_issue: Criar issues" -ForegroundColor White
Write-Host "   - create_pull_request: Criar pull requests" -ForegroundColor White
Write-Host "   - search_code: Buscar código" -ForegroundColor White
Write-Host "   - list_workflow_runs: Listar GitHub Actions" -ForegroundColor White
Write-Host "   - get_repo_stats: Obter estatísticas" -ForegroundColor White

Write-Host "" -ForegroundColor White
Write-Host "🔗 Links úteis:" -ForegroundColor Cyan
Write-Host "- Template: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm" -ForegroundColor White
Write-Host "- Documentação MCP: https://modelcontextprotocol.io" -ForegroundColor White
Write-Host "- GitHub API: https://docs.github.com/en/rest" -ForegroundColor White

Write-Host "" -ForegroundColor White
Write-Host "🆘 Suporte: suporte@vertikon.com" -ForegroundColor Cyan

# === Normalize dependencies (idempotent) ===
Write-Host "🔧 Normalizando dependências..." -ForegroundColor Cyan
try {
    $repoRoot = Split-Path $PSScriptRoot -Parent
    $fixDepsScript = Join-Path $PSScriptRoot "fix-deps.ps1"

    if (Test-Path $fixDepsScript) {
        & $fixDepsScript -RepoRoot $repoRoot
        Write-Host "✅ Dependências normalizadas com sucesso" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Script fix-deps.ps1 não encontrado" -ForegroundColor Yellow
    }
} catch {
    Write-Warning "fix-deps falhou: $($_.Exception.Message)"
    Write-Host "⚠️ Continuando setup sem normalização de dependências" -ForegroundColor Yellow
}

# Retornar ao diretório original
Set-Location $PSScriptRoot