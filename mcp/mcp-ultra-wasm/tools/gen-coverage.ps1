# tools/gen-coverage.ps1  (PowerShell 5.1 friendly)
$ErrorActionPreference = "Stop"

# Pacotes "seguros" (evita build-tags ultra_advanced)
$pkgs = @()
if (Test-Path ".\internal\handlers")     { $pkgs += "./internal/handlers" }
if (Test-Path ".\tests\integration")     { $pkgs += "./tests/integration" }
if (Test-Path ".\tests\smoke")           { $pkgs += "./tests/smoke" }
if ($pkgs.Count -eq 0)                   { $pkgs = @("./...") }

# 1) Gera coverage.out em modo "count" e inclui todas as pkgs no cálculo
& go test $pkgs -count=1 -covermode=count -coverpkg=./... -coverprofile coverage.out

# 2) Gera HTML (PS 5.1: redireção ao invés de -o)
& go tool cover -html coverage.out > coverage.html

# 3) Extrai o total e mostra na tela
$func = & go tool cover -func coverage.out | Out-String
$match = [regex]::Match($func, "total:\s*\(statements\)\s*([\d\.]+)%")
if ($match.Success) {
  $pct = $match.Groups[1].Value
  Write-Host ("✅ Coverage total: {0}%" -f $pct) -ForegroundColor Green
} else {
  Write-Warning "Não foi possível extrair o total do coverage."
}

# 4) (Opcional) Atualiza histórico/badge se os scripts existirem
if ((Test-Path ".\tools\update-coverage-history.ps1") -and (Test-Path ".\docs")) {
  try {
    & pwsh .\tools\update-coverage-history.ps1
  } catch {
    Write-Warning $_
  }
}
