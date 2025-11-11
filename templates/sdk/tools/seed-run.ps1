# ════════════════════════════════════════════════════════════════════════
# SEED RUN - Execução SDK + Seed
# ════════════════════════════════════════════════════════════════════════
# Inicia o SDK (porta 8080) e a seed (porta 8081) simultaneamente
# ════════════════════════════════════════════════════════════════════════

param(
    [int]$SDKPort = 8080,
    [int]$SeedPort = 8081,
    [switch]$SeedOnly,
    [switch]$SDKOnly
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
$SEED_DST = Join-Path $SDK "seeds\mcp-ultra-wasm"

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "   🚀 SEED RUN - Executando SDK + Seed" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Verificar se seed existe
if (-not (Test-Path $SEED_DST)) {
    Write-Host "❌ Seed não encontrada: $SEED_DST" -ForegroundColor Red
    Write-Host "   Execute primeiro: .\tools\seed-sync.ps1" -ForegroundColor Yellow
    exit 1
}

# Função para verificar se porta está em uso
function Test-PortInUse {
    param([int]$Port)
    $connection = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue
    return $connection -ne $null
}

# Verificar portas
if (-not $SeedOnly) {
    if (Test-PortInUse $SDKPort) {
        Write-Host "⚠️  Porta $SDKPort já está em uso" -ForegroundColor Yellow
        $continue = Read-Host "   Deseja continuar mesmo assim? (S/N)"
        if ($continue -ne "S") {
            exit 1
        }
    }
}

if (-not $SDKOnly) {
    if (Test-PortInUse $SeedPort) {
        Write-Host "⚠️  Porta $SeedPort já está em uso" -ForegroundColor Yellow
        $continue = Read-Host "   Deseja continuar mesmo assim? (S/N)"
        if ($continue -ne "S") {
            exit 1
        }
    }
}

# Iniciar SDK
if (-not $SeedOnly) {
    Write-Host "🚀 Iniciando SDK na porta $SDKPort..." -ForegroundColor Yellow

    $sdkScript = @"
Set-Location '$SDK'
Write-Host '════════════════════════════════════════════════════════════' -ForegroundColor Cyan
Write-Host '   🔧 SDK-ULTRA-WASM' -ForegroundColor Cyan
Write-Host '════════════════════════════════════════════════════════════' -ForegroundColor Cyan
Write-Host ''
Write-Host 'Porta: $SDKPort' -ForegroundColor Green
Write-Host 'URL:   http://localhost:$SDKPort' -ForegroundColor Green
Write-Host ''
`$env:PORT = ':$SDKPort'
go run .\cmd
"@

    Start-Process pwsh -ArgumentList '-NoExit', '-Command', $sdkScript
    Write-Host "   ✅ SDK iniciado (janela separada)" -ForegroundColor Green
    Start-Sleep -Seconds 2
}

# Iniciar Seed
if (-not $SDKOnly) {
    Write-Host ""
    Write-Host "🌱 Iniciando Seed na porta $SeedPort..." -ForegroundColor Yellow

    # Verificar se seed tem cmd/main.go
    $seedMainPath = Join-Path $SEED_DST "cmd\main.go"
    if (-not (Test-Path $seedMainPath)) {
        Write-Host "   ⚠️  Seed não possui cmd/main.go" -ForegroundColor Yellow
        Write-Host "   A seed será pulada" -ForegroundColor Gray
    } else {
        $seedScript = @"
Set-Location '$SEED_DST'
Write-Host '════════════════════════════════════════════════════════════' -ForegroundColor Cyan
Write-Host '   🌱 MCP-ULTRA-WASM SEED' -ForegroundColor Cyan
Write-Host '════════════════════════════════════════════════════════════' -ForegroundColor Cyan
Write-Host ''
Write-Host 'Porta: $SeedPort' -ForegroundColor Green
Write-Host 'URL:   http://localhost:$SeedPort' -ForegroundColor Green
Write-Host ''
`$env:PORT = ':$SeedPort'
go run .\cmd
"@

        Start-Process pwsh -ArgumentList '-NoExit', '-Command', $seedScript
        Write-Host "   ✅ Seed iniciada (janela separada)" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "   ✅ SERVIÇOS INICIADOS" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""

if (-not $SeedOnly) {
    Write-Host "📊 SDK:" -ForegroundColor Cyan
    Write-Host "   URL: http://localhost:$SDKPort" -ForegroundColor White
    Write-Host "   Endpoints:" -ForegroundColor White
    Write-Host "     • GET  /health" -ForegroundColor Gray
    Write-Host "     • GET  /healthz" -ForegroundColor Gray
    Write-Host "     • GET  /readyz" -ForegroundColor Gray
    Write-Host "     • GET  /metrics" -ForegroundColor Gray
    Write-Host "     • POST /seed/sync" -ForegroundColor Gray
    Write-Host "     • GET  /seed/status" -ForegroundColor Gray
}

if (-not $SDKOnly) {
    Write-Host ""
    Write-Host "📊 Seed:" -ForegroundColor Cyan
    Write-Host "   URL: http://localhost:$SeedPort" -ForegroundColor White
}

Write-Host ""
Write-Host "💡 Para parar os serviços, feche as janelas do PowerShell" -ForegroundColor Yellow
Write-Host ""
