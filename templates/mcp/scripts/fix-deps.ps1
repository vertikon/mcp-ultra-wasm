# scripts\fix-deps.ps1
# Normaliza dependências Node.js e Go no template MCP Ultra
# Resolve problemas comuns: semconv malformado, GOWORK interference, npm build

param([string]$RepoRoot = (Get-Location).Path)

$ErrorActionPreference = 'Stop'

function Step($text) {
    Write-Host $text -ForegroundColor Cyan
}

function Success($text) {
    Write-Host "  ✓ $text" -ForegroundColor Green
}

function Warning($text) {
    Write-Warning "  ⚠ $text"
}

Step "=== MCP Ultra Dependencies Fix ==="
Write-Host "Repo: $RepoRoot" -ForegroundColor DarkGray

# 1) Node.js/NPM Dependencies
Step "[1/3] Node.js → reinstalando dependências"
$serverDir = Join-Path $RepoRoot "mcp-server"
if (!(Test-Path $serverDir)) {
    $serverDir = $RepoRoot  # fallback caso estrutura mude
}

Push-Location $serverDir
try {
    if (Test-Path "package-lock.json") {
        Step "  usando npm ci (package-lock.json encontrado)"
        npm ci
    } else {
        Step "  usando npm install (sem package-lock.json)"
        npm install
    }
    Step "  executando npm run build"
    npm run build
    Success "Node.js build concluído"
} catch {
    Warning "Falha no Node.js build: $($_.Exception.Message)"
} finally {
    Pop-Location
}

# 2) Go Dependencies - usar GOWORK=off para evitar interferência
Step "[2/3] Go → corrigindo dependências (GOWORK=off)"
$originalGOWORK = $env:GOWORK
$env:GOWORK = "off"  # ignora go.work durante correções

try {
    # Encontrar todos os go.mod (exceto vendor/node_modules)
    $gomods = Get-ChildItem -Recurse -Filter "go.mod" | Where-Object {
        $_.FullName -notmatch "\\vendor\\|\\node_modules\\"
    }

    foreach ($gomod in $gomods) {
        Step "  processando: $($gomod.FullName)"

        # Remover requires inválidos de semconv (padrão problemático)
        $content = Get-Content $gomod.FullName -Raw
        $pattern = "(?ms)^\s*require\s+go\.opentelemetry\.io/otel/semconv/v\d+\.\d+\.\d+\s+v\d+\.\d+\.\d+\s*$"
        $patched = $content -replace $pattern, ""

        if ($patched -ne $content) {
            Set-Content -Path $gomod.FullName -Value $patched -NoNewline
            Warning "removido require inválido de semconv em: $($gomod.FullName)"
        }

        # Executar go mod tidy no diretório do módulo
        Push-Location $gomod.DirectoryName
        try {
            Step "    go get otel@v1.38.0"
            go get go.opentelemetry.io/otel@v1.38.0

            Step "    go mod tidy"
            go mod tidy

            Success "go.mod processado: $($gomod.DirectoryName)"
        } catch {
            Warning "Falha ao processar go.mod em $($gomod.DirectoryName): $($_.Exception.Message)"
        } finally {
            Pop-Location
        }
    }
} finally {
    # Restaurar GOWORK original
    if ($null -ne $originalGOWORK) {
        $env:GOWORK = $originalGOWORK
    } else {
        Remove-Item Env:\GOWORK -ErrorAction SilentlyContinue
    }
}

# 3) Go Autocommit Build (opcional, não deve travar setup)
Step "[3/3] Go → build opcional do autocommit"
$autoDir = Join-Path $RepoRoot "scripts\automation"
$autoSrc = Join-Path $autoDir "autocommit.go"

if (Test-Path $autoSrc) {
    Push-Location $autoDir
    try {
        $env:GOWORK = "off"  # força disable de workspace
        go build -o autocommit.exe autocommit.go
        Success "autocommit.exe compilado"
    } catch {
        Warning "Falha ao compilar autocommit (GOWORK=off). Seguindo sem travar o setup."
        Warning "Erro: $($_.Exception.Message)"
    } finally {
        Pop-Location
        if ($null -ne $originalGOWORK) {
            $env:GOWORK = $originalGOWORK
        } else {
            Remove-Item Env:\GOWORK -ErrorAction SilentlyContinue
        }
    }
} else {
    Write-Host "  (autocommit.go não encontrado em $autoDir; seguindo)" -ForegroundColor DarkGray
}

Step "=== Fix Dependencies Concluído ==="
Write-Host "✅ Dependências normalizadas. Próximos passos:" -ForegroundColor Green
Write-Host "   - npm run build (já executado)" -ForegroundColor DarkGray
Write-Host "   - go build (executado com GOWORK=off)" -ForegroundColor DarkGray
Write-Host "   - Pronto para usar o template!" -ForegroundColor Green