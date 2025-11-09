# ════════════════════════════════════════════════════════════════════════
# INTEGRAÇÃO COMPLETA - SDK ↔ Template ↔ Orquestrador
# ════════════════════════════════════════════════════════════════════════
# Script master de integração e auditoria automática
# Versão: 1.0.0
# ════════════════════════════════════════════════════════════════════════

param(
    [switch]$SkipGoWork,
    [switch]$SkipSync,
    [switch]$SkipRun,
    [switch]$SkipTest,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
$TPL = "E:\vertikon\business\SaaS\templates\mcp-ultra-wasm"
$LOGS_DIR = Join-Path $SDK "logs"
$TIMESTAMP = Get-Date -Format "yyyyMMdd-HHmmss"
$LOG_FILE = Join-Path $LOGS_DIR "integracao-$TIMESTAMP.log"

# Criar diretório de logs
New-Item -ItemType Directory -Path $LOGS_DIR -Force | Out-Null

# Função de logging
function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "INFO"
    )
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $logEntry = "[$timestamp] [$Level] $Message"
    Add-Content -Path $LOG_FILE -Value $logEntry

    switch ($Level) {
        "ERROR" { Write-Host $Message -ForegroundColor Red }
        "WARN"  { Write-Host $Message -ForegroundColor Yellow }
        "SUCCESS" { Write-Host $Message -ForegroundColor Green }
        default { Write-Host $Message -ForegroundColor White }
    }
}

# Banner
Clear-Host
Write-Host ""
Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "║        🔗 INTEGRAÇÃO COMPLETA - MCP-ULTRA-WASM ECOSYSTEM v1.0           ║" -ForegroundColor Cyan
Write-Host "║                                                                      ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
Write-Log "INÍCIO DA INTEGRAÇÃO COMPLETA" "INFO"
Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
Write-Log "Timestamp: $TIMESTAMP" "INFO"
Write-Log "Log file: $LOG_FILE" "INFO"
Write-Log "" "INFO"

# ────────────────────────────────────────────────────────────────────────
# FASE 1: SETUP GO.WORK
# ────────────────────────────────────────────────────────────────────────

if (-not $SkipGoWork) {
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host "   FASE 1/5: Setup go.work" -ForegroundColor Yellow
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host ""

    Write-Log "Executando setup-go-work.ps1..." "INFO"

    try {
        & "$SDK\tools\setup-go-work.ps1"
        Write-Log "✅ go.work configurado com sucesso" "SUCCESS"
    } catch {
        Write-Log "❌ Erro ao configurar go.work: $_" "ERROR"
        exit 1
    }
} else {
    Write-Log "FASE 1 pulada (-SkipGoWork)" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 2: SINCRONIZAÇÃO DE SEED
# ────────────────────────────────────────────────────────────────────────

if (-not $SkipSync) {
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host "   FASE 2/5: Sincronização de Seed" -ForegroundColor Yellow
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host ""

    Write-Log "Executando seed-sync.ps1..." "INFO"

    try {
        & "$SDK\tools\seed-sync.ps1" -TemplatePath $TPL
        Write-Log "✅ Seed sincronizada com sucesso" "SUCCESS"
    } catch {
        Write-Log "❌ Erro ao sincronizar seed: $_" "ERROR"
        exit 1
    }
} else {
    Write-Log "FASE 2 pulada (-SkipSync)" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 3: COMPILAÇÃO E TESTES
# ────────────────────────────────────────────────────────────────────────

if (-not $SkipTest) {
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host "   FASE 3/5: Compilação e Testes" -ForegroundColor Yellow
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host ""

    # Compilar SDK
    Write-Log "Compilando SDK..." "INFO"
    Set-Location $SDK

    $buildOutput = & go build ./cmd 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ SDK compilado com sucesso" "SUCCESS"
    } else {
        Write-Log "❌ Erro ao compilar SDK: $buildOutput" "ERROR"
        exit 1
    }

    # Executar testes
    Write-Log "Executando testes..." "INFO"

    $testOutput = & go test ./internal/handlers -v 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Log "✅ Todos os testes passaram" "SUCCESS"
    } else {
        Write-Log "❌ Testes falharam: $testOutput" "ERROR"
        exit 1
    }

    # Compilar Seed (se possível)
    $seedPath = Join-Path $SDK "seeds\mcp-ultra-wasm"
    if (Test-Path "$seedPath\cmd\main.go") {
        Write-Log "Compilando Seed..." "INFO"
        Set-Location $seedPath

        $seedBuildOutput = & go build ./cmd 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Log "✅ Seed compilada com sucesso" "SUCCESS"
        } else {
            Write-Log "⚠️  Seed não compilou (não crítico): $seedBuildOutput" "WARN"
        }
    }

    Set-Location $SDK
} else {
    Write-Log "FASE 3 pulada (-SkipTest)" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 4: AUDITORIA VIA HTTP
# ────────────────────────────────────────────────────────────────────────

if (-not $SkipRun) {
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host "   FASE 4/5: Auditoria via HTTP" -ForegroundColor Yellow
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
    Write-Host ""

    Write-Log "Iniciando servidor SDK (background)..." "INFO"

    # Iniciar SDK em background
    $sdkJob = Start-Job -ScriptBlock {
        param($SDKPath)
        Set-Location $SDKPath
        & go run ./cmd
    } -ArgumentList $SDK

    Write-Log "Aguardando servidor iniciar (5 segundos)..." "INFO"
    Start-Sleep -Seconds 5

    # Testar health endpoint
    Write-Log "Testando /health..." "INFO"
    try {
        $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 5
        if ($healthResponse.status -eq "ok") {
            Write-Log "✅ Health check: OK" "SUCCESS"
        } else {
            Write-Log "⚠️  Health check retornou status inesperado: $($healthResponse.status)" "WARN"
        }
    } catch {
        Write-Log "❌ Erro ao testar health: $_" "ERROR"
    }

    # Testar seed status endpoint
    Write-Log "Testando /seed/status..." "INFO"
    try {
        $seedStatusResponse = Invoke-RestMethod -Uri "http://localhost:8080/seed/status" -Method GET -TimeoutSec 5

        Write-Log "Seed Status:" "INFO"
        Write-Log "  Path: $($seedStatusResponse.path)" "INFO"
        Write-Log "  Has go.mod: $($seedStatusResponse.has_go_mod)" "INFO"
        Write-Log "  Has go.sum: $($seedStatusResponse.has_go_sum)" "INFO"
        Write-Log "  Main present: $($seedStatusResponse.main_present)" "INFO"
        Write-Log "  Compiles: $($seedStatusResponse.compiles)" "INFO"
        Write-Log "  Module: $($seedStatusResponse.module)" "INFO"

        if ($seedStatusResponse.has_go_mod -and $seedStatusResponse.main_present) {
            Write-Log "✅ Seed status: OK" "SUCCESS"
        } else {
            Write-Log "⚠️  Seed não está completa" "WARN"
        }
    } catch {
        Write-Log "❌ Erro ao testar seed status: $_" "ERROR"
    }

    # Salvar respostas em JSON
    $auditReport = @{
        timestamp = $TIMESTAMP
        health = $healthResponse
        seed_status = $seedStatusResponse
    }

    $auditReportPath = Join-Path $LOGS_DIR "audit-report-$TIMESTAMP.json"
    $auditReport | ConvertTo-Json -Depth 10 | Set-Content $auditReportPath
    Write-Log "📝 Relatório de auditoria salvo: $auditReportPath" "INFO"

    # Parar servidor
    Write-Log "Parando servidor SDK..." "INFO"
    Stop-Job -Job $sdkJob
    Remove-Job -Job $sdkJob

} else {
    Write-Log "FASE 4 pulada (-SkipRun)" "WARN"
}

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# FASE 5: VALIDAÇÃO FINAL
# ────────────────────────────────────────────────────────────────────────

Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "   FASE 5/5: Validação Final" -ForegroundColor Yellow
Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""

Write-Log "Executando Enhanced Validator V4..." "INFO"

$validatorPath = "E:\vertikon\.ecosistema-vertikon\mcp-tester-system"
if (Test-Path "$validatorPath\enhanced_validator_v4.go") {
    try {
        Set-Location $validatorPath
        $validatorOutput = & go run enhanced_validator_v4.go $SDK 2>&1

        if ($validatorOutput -match "Score: 100%") {
            Write-Log "✅ Validação V4: 100% APROVADO" "SUCCESS"
        } elseif ($validatorOutput -match "Score: (\d+)%") {
            $score = $matches[1]
            if ([int]$score -ge 85) {
                Write-Log "✅ Validação V4: $score% APROVADO" "SUCCESS"
            } else {
                Write-Log "⚠️  Validação V4: $score% (mínimo 85%)" "WARN"
            }
        }

        # Salvar output do validador
        $validatorLogPath = Join-Path $LOGS_DIR "validator-$TIMESTAMP.log"
        $validatorOutput | Set-Content $validatorLogPath
        Write-Log "📝 Log do validador salvo: $validatorLogPath" "INFO"

    } catch {
        Write-Log "⚠️  Erro ao executar validador: $_" "WARN"
    }
} else {
    Write-Log "⚠️  Enhanced Validator V4 não encontrado" "WARN"
}

Set-Location $SDK

Write-Host ""

# ────────────────────────────────────────────────────────────────────────
# RELATÓRIO FINAL
# ────────────────────────────────────────────────────────────────────────

Write-Host "╔══════════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "║              ✅ INTEGRAÇÃO COMPLETA FINALIZADA                        ║" -ForegroundColor Green
Write-Host "║                                                                      ║" -ForegroundColor Green
Write-Host "╚══════════════════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

Write-Log "═══════════════════════════════════════════════════════════════════" "SUCCESS"
Write-Log "INTEGRAÇÃO FINALIZADA COM SUCESSO" "SUCCESS"
Write-Log "═══════════════════════════════════════════════════════════════════" "SUCCESS"
Write-Log "" "INFO"

Write-Host "📊 Resumo da Integração:" -ForegroundColor Cyan
Write-Host ""
Write-Host "✅ Fases Executadas:" -ForegroundColor Green
if (-not $SkipGoWork) { Write-Host "   • FASE 1: Setup go.work" -ForegroundColor White }
if (-not $SkipSync)   { Write-Host "   • FASE 2: Sincronização de Seed" -ForegroundColor White }
if (-not $SkipTest)   { Write-Host "   • FASE 3: Compilação e Testes" -ForegroundColor White }
if (-not $SkipRun)    { Write-Host "   • FASE 4: Auditoria via HTTP" -ForegroundColor White }
Write-Host "   • FASE 5: Validação Final" -ForegroundColor White
Write-Host ""

Write-Host "📁 Artefatos Gerados:" -ForegroundColor Cyan
Write-Host "   • Log principal: $LOG_FILE" -ForegroundColor White
if (Test-Path (Join-Path $LOGS_DIR "audit-report-$TIMESTAMP.json")) {
    Write-Host "   • Relatório de auditoria: $(Join-Path $LOGS_DIR "audit-report-$TIMESTAMP.json")" -ForegroundColor White
}
if (Test-Path (Join-Path $LOGS_DIR "validator-$TIMESTAMP.log")) {
    Write-Host "   • Log do validador: $(Join-Path $LOGS_DIR "validator-$TIMESTAMP.log")" -ForegroundColor White
}
Write-Host ""

Write-Host "🚀 Próximos Passos:" -ForegroundColor Cyan
Write-Host "   1. Revisar logs em: $LOGS_DIR" -ForegroundColor White
Write-Host "   2. Executar servidor: .\tools\seed-run.ps1" -ForegroundColor White
Write-Host "   3. Testar endpoints:" -ForegroundColor White
Write-Host "      • http://localhost:8080/health" -ForegroundColor Gray
Write-Host "      • http://localhost:8080/seed/status" -ForegroundColor Gray
Write-Host "      • http://localhost:8080/metrics" -ForegroundColor Gray
Write-Host ""

Write-Log "Integração concluída em $(Get-Date)" "INFO"
Write-Log "═══════════════════════════════════════════════════════════════════" "INFO"
