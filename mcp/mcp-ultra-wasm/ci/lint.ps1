$ErrorActionPreference = "Stop"

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  🚀 CI Pipeline - mcp-ultra-wasm (Windows)" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# 1) Garantir saúde do módulo
Write-Host "📦 [1/4] Executando go mod tidy..." -ForegroundColor Yellow
go mod tidy

Write-Host "✓ [2/4] Verificando go.sum..." -ForegroundColor Yellow
go mod verify

# 2) Lint padrão com gomodguard
Write-Host ""
Write-Host "🔍 [3/4] Executando golangci-lint (gomodguard)..." -ForegroundColor Yellow
golangci-lint run --config=.golangci-new.yml --timeout=5m

# 3) Vettool nativo (depguard-lite)
Write-Host ""
Write-Host "🔨 [4/4] Compilando e executando depguard-lite..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path vettools | Out-Null
go build -o vettools/depguard-lite.exe ./cmd/depguard-lite
$vettool = "$(Get-Location)\vettools\depguard-lite.exe"
go vet -vettool="$vettool" ./...

Write-Host ""
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  ✅ CI PASSED - Todas as verificações OK" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Green
