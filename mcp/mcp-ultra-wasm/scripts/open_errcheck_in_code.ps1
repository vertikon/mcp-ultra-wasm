Param(
  [string]$RepoRoot = "."
)

$errFile = Join-Path $RepoRoot "errcheck.txt"
if (!(Test-Path $errFile)) {
  Write-Host "errcheck.txt não encontrado. Rode: make errcheck-list" -ForegroundColor Yellow
  exit 1
}

Get-Content $errFile | ForEach-Object {
  if ($_ -match "^(.*?):(\d+):\d+:") {
    $file = $Matches[1]
    $line = $Matches[2]
    code "$file:$line"
  }
}
