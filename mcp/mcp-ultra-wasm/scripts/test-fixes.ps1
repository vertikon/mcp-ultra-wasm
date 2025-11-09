param(
  [string]$RepoRoot = (Get-Location).Path
)

$ErrorActionPreference = 'Stop'
function PASS($m){ Write-Host "✓ $m" -ForegroundColor Green }
function FAIL($m){ Write-Host "✗ $m" -ForegroundColor Red; exit 1 }
function STEP($m){ Write-Host "[TEST] $m" -ForegroundColor Cyan }

# 1) Node/NPM
STEP "Node/NPM build"
$server = Join-Path $RepoRoot 'mcp-server'
if (!(Test-Path $server)) { $server = $RepoRoot }
Push-Location $server
try {
  $node = (node -v) 2>$null
  if (-not $node) { FAIL "Node não encontrado no PATH" }
  npm --version | Out-Null
  if (Test-Path "package-lock.json") { npm ci } else { npm install }
  npm run -s build
  PASS "npm build OK ($node)"
} catch { FAIL "npm build falhou: $($_.Exception.Message)" }
finally { Pop-Location }

# 2) Go — checar semconv e tidy
STEP "Go: semconv e tidy"
$gomods = Get-ChildItem -Recurse -Filter "go.mod" | Where-Object { $_.FullName -notmatch "\\vendor\\|\\node_modules\\" }
foreach ($gm in $gomods) {
  $txt = Get-Content $gm.FullName -Raw
  if ($txt -match "(?m)^\s*require\s+go\.opentelemetry\.io/otel/semconv/") {
    FAIL "Encontrado 'require .../semconv/...' em $($gm.FullName)"
  }
}
$old = $env:GOWORK; $env:GOWORK = "off"
try {
  foreach ($gm in $gomods) {
    Push-Location $gm.DirectoryName
    go mod tidy
    go list ./... | Out-Null
    Pop-Location
  }
  PASS "Go tidy/list OK com GOWORK=off"
} catch { FAIL "Go tidy/list falhou: $($_.Exception.Message)" }
finally {
  if ($null -ne $old) { $env:GOWORK = $old } else { Remove-Item Env:\GOWORK -ErrorAction SilentlyContinue }
}

# 3) (Opcional) build do autocommit
STEP "Go: build autocommit (opcional)"
$autoDir = Join-Path $RepoRoot 'scripts\automation'
$autoSrc = Join-Path $autoDir 'autocommit.go'
if (Test-Path $autoSrc) {
  Push-Location $autoDir
  try {
    $env:GOWORK = "off"
    go build -o autocommit.exe autocommit.go
    PASS "autocommit.exe compilado"
  } catch {
    Write-Warning "autocommit: falha tolerada: $($_.Exception.Message)"
  } finally {
    Pop-Location
    Remove-Item Env:\GOWORK -ErrorAction SilentlyContinue
  }
} else {
  Write-Host "(autocommit não presente — pulando)" -ForegroundColor DarkGray
}

PASS "Testes concluídos com sucesso"
exit 0