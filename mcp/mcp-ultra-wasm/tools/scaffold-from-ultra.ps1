param(
  [Parameter(Mandatory=$true)][string]$Name,
  [Parameter(Mandatory=$true)][string]$NewModule,
  [Parameter(Mandatory=$true)][string]$NewPath,
  [switch]$KeepHistory
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "🌱 Scaffold from mcp-ultra-wasm" -ForegroundColor Cyan
Write-Host "──────────────────────────" -ForegroundColor DarkCyan
Write-Host ""

# 1) Copiar (sem .git)
Write-Host "📦 Copiando estrutura..." -ForegroundColor Yellow
$excludeDirs = ".git", ".github\workflows\_cache", "bin", "obj"
if (-not $KeepHistory) {
  $excludeDirs += @("docs\coverage_history", "docs\latency")
}
$xd = ($excludeDirs | ForEach-Object { "/XD `"$_`"" }) -join " "
$xf = "/XF coverage.out coverage.html coverage_func.txt"
$cmd = "robocopy `"$PSScriptRoot\..`" `"$NewPath`" /E $xd $xf"
Invoke-Expression $cmd | Out-Null
if ($LASTEXITCODE -gt 7) { throw "Falha ao copiar (rc=$LASTEXITCODE)" }

# 2) Limpar VCS/artefatos adicionais
Write-Host "🧹 Limpando artefatos..." -ForegroundColor Yellow
Get-ChildItem -Path $NewPath -Recurse -Force -Include '.git','coverage.out','coverage.html','coverage_func.txt' |
  Remove-Item -Recurse -Force -ErrorAction SilentlyContinue

# 3) Ajustar module
Write-Host "📝 Ajustando go.mod..." -ForegroundColor Yellow
$gomod = Join-Path $NewPath 'go.mod'
if (Test-Path $gomod) {
  (Get-Content $gomod) -replace '^module\s+.*', "module $NewModule" | Set-Content $gomod -Encoding UTF8
}

# 4) Reescrever imports e placeholders
Write-Host "🔄 Reescrevendo imports e placeholders..." -ForegroundColor Yellow
$old = 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm'
Get-ChildItem -Path $NewPath -Recurse -Include *.go,*.md,README-Template.md |
  ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $content = $content -replace [regex]::Escape($old), $NewModule
    $content = $content -replace 'mcp-ultra-wasm', $Name
    $content = $content -replace '%PROJECT_NAME%', $Name
    Set-Content -Path $_.FullName -Value $content -Encoding UTF8 -NoNewline
  }

# 5) Renomear README-Template.md para README.md se necessário
$readmeTemplate = Join-Path $NewPath 'README-Template.md'
$readme = Join-Path $NewPath 'README.md'
if (Test-Path $readmeTemplate) {
  if (Test-Path $readme) { Remove-Item $readme -Force }
  Move-Item $readmeTemplate $readme -Force
  Write-Host "📄 README-Template.md → README.md" -ForegroundColor Green
}

# 5) go mod tidy + smoke test
Write-Host "⚙️  Rodando go mod tidy..." -ForegroundColor Yellow
Push-Location $NewPath
try {
  go mod tidy

  Write-Host "🧪 Smoke test..." -ForegroundColor Yellow
  go test ./... -count=1

  Write-Host ""
  Write-Host "✅ Scaffold criado com sucesso!" -ForegroundColor Green
  Write-Host ""
  Write-Host "📂 Local: $NewPath" -ForegroundColor White
  Write-Host "📦 Module: $NewModule" -ForegroundColor White
  Write-Host ""

} catch {
  Write-Host ""
  Write-Host "⚠️  Erro durante setup: $_" -ForegroundColor Red
  Write-Host ""
  throw
} finally {
  Pop-Location
}
