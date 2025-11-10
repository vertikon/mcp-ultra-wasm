# ════════════════════════════════════════════════════════════════════════
# SEED SYNC - Sincronização Template → Seed Interna
# ════════════════════════════════════════════════════════════════════════
# Clona/atualiza o template mcp-ultra-wasm como seed interna do SDK
# ════════════════════════════════════════════════════════════════════════

param(
    [string]$TemplatePath = "E:\vertikon\business\SaaS\templates\mcp-ultra-wasm",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
$FIX = "E:\vertikon\.ecosistema-vertikon\shared\mcp-ultra-wasm-fix"
$SEED_DST = Join-Path $SDK "seeds\mcp-ultra-wasm"

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "   🌱 SEED SYNC - Sincronização de Template" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Verificar se template existe
if (-not (Test-Path $TemplatePath)) {
    Write-Host "❌ Template não encontrado: $TemplatePath" -ForegroundColor Red
    exit 1
}

Write-Host "📂 Origem:  $TemplatePath" -ForegroundColor Cyan
Write-Host "📂 Destino: $SEED_DST" -ForegroundColor Cyan
Write-Host ""

# 1. Criar diretório da seed
Write-Host "📁 Preparando diretório da seed..." -ForegroundColor Yellow

$seedDir = Split-Path $SEED_DST
if (-not (Test-Path $seedDir)) {
    New-Item -ItemType Directory -Path $seedDir -Force | Out-Null
    Write-Host "   ✅ Diretório criado: $seedDir" -ForegroundColor Green
}

# 2. Espelhar template → seed (robocopy)
Write-Host ""
Write-Host "🔄 Espelhando template para seed..." -ForegroundColor Yellow

$robocopyArgs = @(
    $TemplatePath,
    $SEED_DST,
    "/E",           # Copia subdiretórios incluindo vazios
    "/MIR",         # Mirror (espelha, remove o que não existe no source)
    "/XD", ".git",  # Exclui diretório .git
    "/NP",          # Sem progresso por arquivo
    "/NFL",         # Sem lista de arquivos
    "/NDL"          # Sem lista de diretórios
)

$robocopyOutput = & robocopy @robocopyArgs 2>&1

# Robocopy exit codes: 0-1 = success, 2+ = errors
if ($LASTEXITCODE -le 1) {
    Write-Host "   ✅ Template espelhado com sucesso" -ForegroundColor Green
} elseif ($LASTEXITCODE -eq 2) {
    Write-Host "   ⚠️  Alguns arquivos extras detectados" -ForegroundColor Yellow
} else {
    Write-Host "   ❌ Erro no espelhamento (código $LASTEXITCODE)" -ForegroundColor Red
    Write-Host $robocopyOutput -ForegroundColor Red
    exit 1
}

# 3. Ajustar go.mod da seed
Write-Host ""
Write-Host "📝 Ajustando go.mod da seed..." -ForegroundColor Yellow

$gomodPath = Join-Path $SEED_DST "go.mod"

if (-not (Test-Path $gomodPath)) {
    Write-Host "   ❌ go.mod não encontrado na seed" -ForegroundColor Red
    exit 1
}

# Ler go.mod
$gomodContent = Get-Content $gomodPath -Raw

# Substituir module name
$gomodContent = $gomodContent -replace '^module\s+.*', 'module seeds/mcp-ultra-wasm'

# Escrever go.mod ajustado
Set-Content -Path $gomodPath -Value $gomodContent -Encoding UTF8

Write-Host "   ✅ Module name ajustado: seeds/mcp-ultra-wasm" -ForegroundColor Green

# 4. Adicionar replaces
Write-Host ""
Write-Host "🔗 Adicionando replaces..." -ForegroundColor Yellow

$replaces = @"

replace github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm => $SDK
replace github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix => $FIX
"@

Add-Content -Path $gomodPath -Value $replaces -Encoding UTF8

Write-Host "   ✅ Replace SDK:  $SDK" -ForegroundColor Green
Write-Host "   ✅ Replace FIX:  $FIX" -ForegroundColor Green

# 5. go mod tidy
Write-Host ""
Write-Host "🧹 Executando go mod tidy..." -ForegroundColor Yellow

Push-Location $SEED_DST

$tidyOutput = & go mod tidy 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ go mod tidy executado com sucesso" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Avisos durante go mod tidy:" -ForegroundColor Yellow
    Write-Host $tidyOutput -ForegroundColor Gray
}

Pop-Location

# 6. Verificar integridade
Write-Host ""
Write-Host "🔍 Verificando integridade da seed..." -ForegroundColor Yellow

$checks = @{
    "go.mod" = (Test-Path (Join-Path $SEED_DST "go.mod"))
    "go.sum" = (Test-Path (Join-Path $SEED_DST "go.sum"))
    "cmd/main.go" = (Test-Path (Join-Path $SEED_DST "cmd\main.go"))
}

$allOk = $true
foreach ($check in $checks.GetEnumerator()) {
    if ($check.Value) {
        Write-Host "   ✅ $($check.Key)" -ForegroundColor Green
    } else {
        Write-Host "   ❌ $($check.Key)" -ForegroundColor Red
        $allOk = $false
    }
}

Write-Host ""
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "   ✅ SEED SINCRONIZADA COM SUCESSO" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Localização: $SEED_DST" -ForegroundColor Cyan
Write-Host "📦 Module: seeds/mcp-ultra-wasm" -ForegroundColor Cyan
Write-Host "🔗 Replaces: 2 (SDK + FIX)" -ForegroundColor Cyan
Write-Host ""

if (-not $allOk) {
    Write-Host "⚠️  Alguns arquivos esperados não foram encontrados" -ForegroundColor Yellow
    exit 1
}
