# ════════════════════════════════════════════════════════════════════════
# SETUP GO.WORK - Workspace Unificado Vertikon
# ════════════════════════════════════════════════════════════════════════
# Cria/atualiza go.work para resolver imports localmente
# ════════════════════════════════════════════════════════════════════════

param(
    [string]$Root = "E:\vertikon"
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
$TPL = "E:\vertikon\business\SaaS\templates\mcp-ultra-wasm"
$FIX = "E:\vertikon\.ecosistema-vertikon\shared\mcp-ultra-wasm-fix"

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "   🔧 SETUP GO.WORK - Workspace Unificado" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Verificar se os diretórios existem
Write-Host "📁 Verificando diretórios..." -ForegroundColor Yellow

if (-not (Test-Path $SDK)) {
    Write-Host "   ❌ SDK não encontrado: $SDK" -ForegroundColor Red
    exit 1
}
Write-Host "   ✅ SDK: $SDK" -ForegroundColor Green

if (-not (Test-Path $TPL)) {
    Write-Host "   ❌ Template não encontrado: $TPL" -ForegroundColor Red
    exit 1
}
Write-Host "   ✅ Template: $TPL" -ForegroundColor Green

if (-not (Test-Path $FIX)) {
    Write-Host "   ⚠️  mcp-ultra-wasm-fix não encontrado: $FIX" -ForegroundColor Yellow
    Write-Host "   ℹ️  Criando diretório..." -ForegroundColor Gray
    New-Item -ItemType Directory -Path $FIX -Force | Out-Null
}
Write-Host "   ✅ Fix: $FIX" -ForegroundColor Green

# Criar go.work
Write-Host ""
Write-Host "📝 Criando go.work..." -ForegroundColor Yellow

Set-Location $Root

$goWorkContent = @"
go 1.25

use (
  .\.ecosistema-vertikon\shared\mcp-ultra-wasm-fix
  .\business\SaaS\templates\mcp-ultra-wasm
  .\business\SaaS\templates\sdk-ultra-wasm
)
"@

$goWorkPath = Join-Path $Root "go.work"
Set-Content -Path $goWorkPath -Value $goWorkContent -Encoding UTF8

Write-Host "   ✅ go.work criado em: $goWorkPath" -ForegroundColor Green

# Executar go work sync
Write-Host ""
Write-Host "🔄 Sincronizando workspace..." -ForegroundColor Yellow

$syncOutput = & go work sync 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ Workspace sincronizado" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Aviso durante sincronização:" -ForegroundColor Yellow
    Write-Host $syncOutput -ForegroundColor Gray
}

Write-Host ""
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "   ✅ GO.WORK CONFIGURADO COM SUCESSO" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Módulos no workspace:" -ForegroundColor Cyan
Write-Host "   • mcp-ultra-wasm-fix" -ForegroundColor White
Write-Host "   • mcp-ultra-wasm (template)" -ForegroundColor White
Write-Host "   • sdk-ultra-wasm" -ForegroundColor White
Write-Host ""
