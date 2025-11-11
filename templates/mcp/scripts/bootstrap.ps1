# ============================================================================
# bootstrap.ps1
# ============================================================================
# Substitui {{MODULE_PATH}} pelo módulo real da semente
# RODAR EM CADA NOVA SEMENTE CRIADA A PARTIR DO TEMPLATE
# ============================================================================

Param(
  [Parameter(Mandatory = $true, HelpMessage = "Ex: github.com/vertikon/minha-semente")]
  [string]$ModulePath
)

Write-Host "🚀 Iniciando bootstrap da semente..." -ForegroundColor Cyan
Write-Host "📦 Módulo: $ModulePath`n" -ForegroundColor White

# Detectar diretório raiz do projeto
$Here = Split-Path -Parent $MyInvocation.MyCommand.Path
$Root = Resolve-Path (Join-Path $Here "..") | Select-Object -ExpandProperty Path
Set-Location $Root

Write-Host "📁 Diretório: $Root`n" -ForegroundColor Gray

# ============================================================================
# 1) Validar que estamos em um template não-bootstrapped
# ============================================================================
Write-Host "🔍 [1/6] Validando template..." -ForegroundColor Yellow

if (-not (Test-Path ".\go.mod")) {
    Write-Host "   ❌ go.mod não encontrado!" -ForegroundColor Red
    exit 1
}

$gomod = Get-Content .\go.mod -Raw

if ($gomod -notmatch '\{\{MODULE_PATH\}\}') {
    Write-Host "   ⚠️  Template já foi bootstrapped ou não contém {{MODULE_PATH}}" -ForegroundColor Yellow
    Write-Host "   ℹ️  Módulo atual: " -NoNewline -ForegroundColor Gray
    if ($gomod -match 'module\s+([^\s]+)') {
        Write-Host $matches[1] -ForegroundColor White
    }
    $continue = Read-Host "`n   Deseja continuar mesmo assim? (S/N)"
    if ($continue -ne 'S' -and $continue -ne 's') {
        Write-Host "   🛑 Bootstrap cancelado" -ForegroundColor Red
        exit 0
    }
}

Write-Host "   ✅ Template válido`n" -ForegroundColor Green

# ============================================================================
# 2) Substituir module no go.mod
# ============================================================================
Write-Host "📝 [2/6] Atualizando go.mod..." -ForegroundColor Yellow

$gomod = Get-Content .\go.mod -Raw
$gomod = $gomod -replace 'module\s+\{\{MODULE_PATH\}\}', "module $ModulePath"

# Garantir dependência do mcp-ultra-wasm-fix
if ($gomod -notmatch 'github\.com/vertikon/mcp-ultra-wasm-fix') {
    Write-Host "   📦 Adicionando mcp-ultra-wasm-fix..." -ForegroundColor Gray
    $gomod += "`n`nrequire github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix v0.1.0"
}

Set-Content .\go.mod $gomod -NoNewline

Write-Host "   ✅ go.mod atualizado: module $ModulePath`n" -ForegroundColor Green

# ============================================================================
# 3) Substituir imports nos arquivos .go
# ============================================================================
Write-Host "🔄 [3/6] Substituindo imports nos arquivos .go..." -ForegroundColor Yellow

$files = Get-ChildItem -Recurse -Include *.go -File | Where-Object {
  $_.FullName -notmatch '\\vendor\\' -and
  $_.FullName -notmatch '\\.git\\' -and
  $_.FullName -notmatch '\\node_modules\\'
}

$totalFiles = 0

foreach ($f in $files) {
  $c = Get-Content $f.FullName -Raw
  $original = $c

  # Substituir placeholder pelo módulo real
  $c = $c -replace '\{\{MODULE_PATH\}\}', $ModulePath

  if ($c -ne $original) {
    Set-Content $f.FullName $c -NoNewline
    $totalFiles++
    Write-Host "   📝 $($f.Name)" -ForegroundColor Cyan
  }
}

Write-Host "   ✅ $totalFiles arquivos atualizados`n" -ForegroundColor Green

# ============================================================================
# 4) Substituir em arquivos de configuração (YAML, MD, etc)
# ============================================================================
Write-Host "⚙️  [4/6] Atualizando arquivos de configuração..." -ForegroundColor Yellow

$configFiles = Get-ChildItem -Recurse -Include *.md,*.yaml,*.yml,*.json,*.toml -File | Where-Object {
  $_.FullName -notmatch '\\vendor\\' -and
  $_.FullName -notmatch '\\.git\\' -and
  $_.FullName -notmatch '\\node_modules\\'
}

$totalConfigFiles = 0

foreach ($f in $configFiles) {
  try {
    $c = Get-Content $f.FullName -Raw -ErrorAction Stop
    $original = $c

    $c = $c -replace '\{\{MODULE_PATH\}\}', $ModulePath

    if ($c -ne $original) {
      Set-Content $f.FullName $c -NoNewline
      $totalConfigFiles++
      Write-Host "   📝 $($f.Name)" -ForegroundColor Cyan
    }
  } catch {
    # Ignorar arquivos binários
  }
}

Write-Host "   ✅ $totalConfigFiles arquivos de config atualizados`n" -ForegroundColor Green

# ============================================================================
# 5) Garantir estrutura de test/mocks
# ============================================================================
Write-Host "🧪 [5/6] Configurando test/mocks..." -ForegroundColor Yellow

New-Item -ItemType Directory -Force -Path ".\test\mocks" | Out-Null

if (-not (Test-Path ".\test\mocks\.gitkeep")) {
    New-Item -ItemType File -Path ".\test\mocks\.gitkeep" -Force | Out-Null
    Write-Host "   ✅ .gitkeep criado" -ForegroundColor Green
}

Write-Host "   ℹ️  Mocks locais prontos em test/mocks/`n" -ForegroundColor Gray

# ============================================================================
# 6) Resolver dependências
# ============================================================================
Write-Host "📦 [6/6] Resolvendo dependências..." -ForegroundColor Yellow

try {
    & go mod tidy 2>&1 | Out-Host
    Write-Host "   ✅ go mod tidy concluído`n" -ForegroundColor Green
} catch {
    Write-Host "   ⚠️  Erro ao executar go mod tidy" -ForegroundColor Yellow
    Write-Host "   ℹ️  Execute manualmente: go mod tidy`n" -ForegroundColor Gray
}

# ============================================================================
# Resumo Final
# ============================================================================
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✅ BOOTSTRAP CONCLUÍDO COM SUCESSO!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "📦 Módulo: $ModulePath" -ForegroundColor White
Write-Host "📊 Estatísticas:" -ForegroundColor Yellow
Write-Host "   • Arquivos .go atualizados: $totalFiles" -ForegroundColor White
Write-Host "   • Arquivos de config atualizados: $totalConfigFiles" -ForegroundColor White
Write-Host ""
Write-Host "🧪 Próximos passos de validação:" -ForegroundColor Yellow
Write-Host "   1. Compilar: " -NoNewline -ForegroundColor White
Write-Host "go build ./..." -ForegroundColor Cyan
Write-Host "   2. Testar: " -NoNewline -ForegroundColor White
Write-Host "go test ./..." -ForegroundColor Cyan
Write-Host "   3. Validar: " -NoNewline -ForegroundColor White
Write-Host "E:\go1.25.0\go\bin\go.exe run E:\vertikon\.ecosistema-vertikon\mcp-tester-system\enhanced_validator_v7.go ." -ForegroundColor Cyan
Write-Host ""
Write-Host "📚 Documentação:" -ForegroundColor Yellow
Write-Host "   • Template: TEMPLATE_GUIDE.md" -ForegroundColor White
Write-Host "   • Mocks: test/mocks/README.md" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  Lembre-se:" -ForegroundColor Yellow
Write-Host "   • Use APENAS github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-fix/pkg/..." -ForegroundColor White
Write-Host "   • NUNCA importe github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/..." -ForegroundColor White
Write-Host "   • Mocks locais em test/mocks/" -ForegroundColor White
Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
