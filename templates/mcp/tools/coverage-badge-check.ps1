param(
  [string]$CoverageFuncPath = "coverage_func.txt",
  [string]$BadgePath = "docs\badges\coverage.svg",
  [double]$Tolerance = 0.1
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $CoverageFuncPath)) { throw "Não achei $CoverageFuncPath" }
if (-not (Test-Path $BadgePath))      { throw "Não achei $BadgePath" }

$func = Get-Content $CoverageFuncPath -Raw
$badge = Get-Content $BadgePath -Raw

$covMatch = [regex]::Match($func, 'total:\s+\(statements\)\s+([0-9]+\.[0-9]+)%')
if (-not $covMatch.Success) { throw "Não consegui extrair total do coverage_func.txt" }
$cov = [double]$covMatch.Groups[1].Value

# o badge escreve o valor no segundo <text>, ex: <text x="110" ...>19.7%</text>
$badgeMatch = [regex]::Match($badge, '>\s*([0-9]+\.[0-9]+)%\s*<')
if (-not $badgeMatch.Success) { throw "Não consegui extrair % do badge SVG" }
$badgeCov = [double]$badgeMatch.Groups[1].Value

$diff = [math]::Abs($cov - $badgeCov)
Write-Host ("Badge: {0:n1}% | Func: {1:n1}% | Δ={2:n2}" -f $badgeCov, $cov, $diff)

if ($diff -gt $Tolerance) {
  throw ("❌ Badge inconsistente: SVG={0:n1}% vs Func={1:n1}% (tolerance={2})" -f $badgeCov, $cov, $Tolerance)
}

Write-Host "✅ Badge consistente."
