param(
  [string]$Url = "http://localhost:8080/ping",
  [int]$Requests = 60,
  [int]$SpacingMs = 0,
  [int]$TimeoutSec = 3,
  [int]$SlaMs = 200,
  [int]$P95MaxMs = 300,
  [int]$StdMaxMs = 60,
  [double]$ErrorRateMax = 0.0,
  [switch]$WriteReport,
  [string]$OutDir = "docs\latency",
  [string]$Version = "local"
)

$ErrorActionPreference = "Stop"

function To-DoubleArray($v) {
  $list = New-Object System.Collections.Generic.List[double]
  foreach ($x in @($v)) { [void]$list.Add([double]$x) }
  ,$list.ToArray()
}

function Get-Percentile($v, [double]$p) {
  $arr = To-DoubleArray $v
  if ($arr.Length -eq 0) { return 0.0 }
  $s = $arr | Sort-Object
  $idx = [int][math]::Ceiling($p * $s.Count) - 1
  if ($idx -lt 0) { $idx = 0 }
  if ($idx -ge $s.Count) { $idx = $s.Count - 1 }
  return [double]$s[$idx]
}

function Get-StdDev($v, [double]$mean) {
  $arr = To-DoubleArray $v
  if ($arr.Length -le 1) { return 0.0 }
  $sum = 0.0
  foreach ($x in $arr) { $d = $x - $mean; $sum += $d * $d }
  [math]::Sqrt($sum / $arr.Length)
}

# Coleta
$times = New-Object System.Collections.Generic.List[double]
$errs = 0
for ($i = 0; $i -lt $Requests; $i++) {
  $sw = [System.Diagnostics.Stopwatch]::StartNew()
  try {
    Invoke-WebRequest -Method GET -Uri $Url -TimeoutSec $TimeoutSec | Out-Null
    $sw.Stop()
    [void]$times.Add([math]::Round($sw.Elapsed.TotalMilliseconds, 1))
  } catch {
    $sw.Stop()
    $errs++
  }
  if ($SpacingMs -gt 0) { Start-Sleep -Milliseconds $SpacingMs }
}

if ($times.Count -eq 0) { throw "No valid samples - is the server up? URL=$Url" }

# Estatísticas
$mean   = [math]::Round(($times | Measure-Object -Average).Average, 1)
$median = [math]::Round((Get-Percentile $times 0.5), 1)
$p95    = [math]::Round((Get-Percentile $times 0.95), 1)
$p99    = [math]::Round((Get-Percentile $times 0.99), 1)
$std    = [math]::Round((Get-StdDev $times $mean), 1)
$min    = [math]::Round(($times | Measure-Object -Minimum).Minimum, 1)
$max    = [math]::Round(($times | Measure-Object -Maximum).Maximum, 1)
$errRate = [math]::Round(($errs / [double]$Requests), 3)

Write-Host ("URL: {0}" -f $Url)
Write-Host ("N={0} | mean={1} ms | median={2} ms | p95={3} ms | p99={4} ms | std={5} ms | min={6} | max={7} | errRate={8}" -f $times.Count,$mean,$median,$p95,$p99,$std,$min,$max,$errRate)

# SLAs
$ok = ($mean -le $SlaMs) -and ($std -le $StdMaxMs) -and ($p95 -le $P95MaxMs) -and ($errRate -le $ErrorRateMax)
if ($ok) { Write-Host "Latency stability OK." } else { throw "Latency stability FAILED." }

# Relatórios
if ($WriteReport) {
  if (-not (Test-Path $OutDir)) { New-Item -ItemType Directory -Force -Path $OutDir | Out-Null }

  $csvLine = "{0},{1},{2},{3},{4},{5},{6},{7},{8},{9}" -f (Get-Date -Format "yyyy-MM-dd HH:mm:ss"),$Version,$times.Count,$mean,$median,$p95,$p99,$std,$errRate,$Url
  Add-Content -Path (Join-Path $OutDir 'latency_history.csv') -Value $csvLine -Encoding utf8

  $md = @'
# Latency Report

| N | mean (ms) | median | p95 | p99 | std | min | max | errRate |
| -:| --------: | -----: | --: | --: | --: | --: | --: | ------: |
'@
  $md += "| {0} | {1} | {2} | {3} | {4} | {5} | {6} | {7} | {8} |`n" -f $times.Count,$mean,$median,$p95,$p99,$std,$min,$max,$errRate
  $md += "`nURL: $Url`nVersion: $Version`nTimestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')`n"
  Set-Content -Path (Join-Path $OutDir 'latency_report.md') -Value $md -Encoding utf8
}
