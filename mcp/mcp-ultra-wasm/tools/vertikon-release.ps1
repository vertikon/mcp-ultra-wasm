<#
.SYNOPSIS
    Pipeline de release público do MCP-Ultra

.DESCRIPTION
    Script de automação que:
    - Limpa código proprietário (scrub)
    - Redige segredos e URLs internas
    - Valida conformidade de licenças
    - Adiciona headers Apache 2.0
    - Gera changelog parcial
    - Publica no repositório público

.PARAMETER PublicRepoUrl
    URL do repositório público (ex: https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git)

.PARAMETER OutDir
    Diretório de saída para código limpo (padrão: ./public)

.PARAMETER Version
    Versão semântica para tag (ex: 1.0.0). Se omitido, usa data atual.

.PARAMETER DryRun
    Se presente, executa validação mas não faz push

.EXAMPLE
    .\tools\vertikon-release.ps1 -PublicRepoUrl "https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git" -Version "1.0.0"

.EXAMPLE
    .\tools\vertikon-release.ps1 -PublicRepoUrl "https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git" -DryRun
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$PublicRepoUrl,

    [string]$OutDir = "./public",
    [string]$Version = "",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"
$WarningPreference = "Continue"

# ============================================================================
# Funções Auxiliares
# ============================================================================

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Failure {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Read-YamlFile {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        throw "Arquivo de configuração não encontrado: $Path"
    }

    # Parser YAML simples (PowerShell 5.1 compatível)
    # Para produção, considere usar powershell-yaml module
    $content = Get-Content $Path -Raw

    # Converter YAML para objeto customizado (simplificado)
    $obj = @{}
    $currentKey = $null
    $lines = $content -split "`n"

    foreach ($line in $lines) {
        $line = $line.Trim()
        if ($line -match '^([a-zA-Z_][a-zA-Z0-9_]*):(.*)$') {
            $currentKey = $matches[1]
            $value = $matches[2].Trim()
            if ($value) {
                $obj[$currentKey] = $value
            } else {
                $obj[$currentKey] = @()
            }
        } elseif ($line -match '^\s*-\s*(.+)$' -and $currentKey) {
            $item = $matches[1].Trim()
            $obj[$currentKey] += $item
        }
    }

    return $obj
}

function Test-GitRepository {
    try {
        git rev-parse --is-inside-work-tree 2>&1 | Out-Null
        return $true
    } catch {
        return $false
    }
}

function Get-FilesByPattern {
    param(
        [string]$Path,
        [string[]]$Patterns
    )

    $files = @()
    foreach ($pattern in $Patterns) {
        $files += Get-ChildItem -Path $Path -Recurse -Include $pattern -File -ErrorAction SilentlyContinue
    }
    return $files
}

# ============================================================================
# Inicialização
# ============================================================================

Write-Step "Inicializando pipeline de release público MCP-Ultra"

$Root = (Resolve-Path ".").Path
$CfgDir = Join-Path $Root ".release"
$OutDirFull = Join-Path $Root $OutDir

# Validar git
if (-not (Test-GitRepository)) {
    Write-Failure "Diretório atual não é um repositório Git"
    exit 1
}

# Carregar configurações
Write-Step "Carregando configurações de release"

try {
    $scrubConfig = @{
        exclude = @(
            ".git/**", "internal/enterprise/**", "pkg/enterprise/**",
            "configs/prod/**", "vendor/**", "tests/integration/**",
            "**/.env", "**/.env.*", "**/secrets/**", "**/*.pem", "**/*.key",
            "**/id_rsa*", ".backup_*/**", "docs/gaps/**", "docs/melhorias/**",
            "**/analyze_gaps.ps1"
        )
        sanitize = @(
            @{pattern="vertikon.internal"; replace="localhost"},
            @{pattern="vertikon-private"; replace="example-org"},
            @{pattern="E:\\vertikon"; replace="/workspace"},
            @{pattern="E:\\rfesta"; replace="/config"}
        )
    }

    $redactPatterns = @(
        "(?i)api[_-]?key\s*[:=]\s*[`"'][A-Za-z0-9_\-]{16,}[`"']",
        "(?i)secret\s*[:=]\s*[`"'][A-Za-z0-9_\-]{12,}[`"']",
        "(?i)password\s*[:=]\s*[`"'][^`"']{6,}[`"']",
        "(?i)token\s*[:=]\s*[`"'][A-Za-z0-9_\-]{20,}[`"']",
        "(?i)wss?://[a-zA-Z0-9.-]*vertikon[a-zA-Z0-9.-]*\.[a-z]{2,}",
        "(?i)https?://[a-zA-Z0-9.-]*vertikon[a-zA-Z0-9.-]*\.[a-z]{2,}",
        "postgresql://[^\s@]+@[^\s/]+/[^\s?]+",
        "redis://[^\s@]+@[^\s/]+",
        "nats://[^\s@]+@[^\s/]+",
        "Bearer [A-Za-z0-9_\-\.]{20,}"
    )

    $blocklistDeps = @(
        "github.com/vertikon-private/*",
        "github.com/vertikon/internal-*",
        "gitlab.vertikon.com/*"
    )

    Write-Success "Configurações carregadas com sucesso"
} catch {
    Write-Failure "Erro ao carregar configurações: $_"
    exit 1
}

# ============================================================================
# 1. Preparar diretório limpo
# ============================================================================

Write-Step "Preparando diretório de saída: $OutDir"

if (Test-Path $OutDirFull) {
    Remove-Item $OutDirFull -Recurse -Force
    Write-Success "Diretório anterior removido"
}

New-Item -ItemType Directory -Force -Path $OutDirFull | Out-Null
Write-Success "Diretório criado: $OutDirFull"

# ============================================================================
# 2. Copiar arquivos (respeitando excludes)
# ============================================================================

Write-Step "Copiando arquivos (aplicando filtros de exclusão)"

$allFiles = git ls-files
$copiedCount = 0
$excludedCount = 0

foreach ($file in $allFiles) {
    $shouldExclude = $false

    foreach ($pattern in $scrubConfig.exclude) {
        $regexPattern = $pattern -replace '\*\*', '.*' -replace '\*', '[^/]*'
        if ($file -match $regexPattern) {
            $shouldExclude = $true
            $excludedCount++
            break
        }
    }

    if (-not $shouldExclude) {
        $srcPath = Join-Path $Root $file
        $dstPath = Join-Path $OutDirFull $file

        $dstDir = Split-Path $dstPath -Parent
        if (-not (Test-Path $dstDir)) {
            New-Item -ItemType Directory -Force -Path $dstDir | Out-Null
        }

        Copy-Item $srcPath $dstPath -Force
        $copiedCount++
    }
}

Write-Success "Arquivos copiados: $copiedCount | Excluídos: $excludedCount"

# ============================================================================
# 3. Sanitizar exemplos e redigir segredos
# ============================================================================

Write-Step "Sanitizando conteúdo (redação de segredos e substituição de padrões)"

$processedFiles = 0
$redactedCount = 0
$sanitizedCount = 0

Get-ChildItem -Path $OutDirFull -Recurse -File | ForEach-Object {
    $filePath = $_.FullName

    # Pular arquivos binários
    $ext = $_.Extension.ToLower()
    if ($ext -in @('.exe', '.dll', '.so', '.dylib', '.bin', '.pdf', '.png', '.jpg', '.jpeg', '.gif')) {
        return
    }

    try {
        $content = Get-Content $filePath -Raw -ErrorAction Stop
        $originalContent = $content

        # Aplicar sanitização
        foreach ($rule in $scrubConfig.sanitize) {
            $pattern = [regex]::Escape($rule.pattern)
            $content = $content -replace $pattern, $rule.replace
        }
        if ($content -ne $originalContent) { $sanitizedCount++ }

        # Aplicar redação
        $redacted = $false
        foreach ($pattern in $redactPatterns) {
            if ($content -match $pattern) {
                $content = $content -replace $pattern, "REDACTED"
                $redacted = $true
            }
        }
        if ($redacted) { $redactedCount++ }

        # Salvar se houve alterações
        if ($content -ne $originalContent) {
            Set-Content $filePath $content -NoNewline
        }

        $processedFiles++

    } catch {
        # Arquivo binário ou não legível - pular
    }
}

Write-Success "Arquivos processados: $processedFiles | Sanitizados: $sanitizedCount | Redigidos: $redactedCount"

# ============================================================================
# 4. Adicionar headers de licença
# ============================================================================

Write-Step "Adicionando headers de licença Apache 2.0"

$licenseHeader = @"
// Copyright (c) 2025 Vertikon
// Licensed under the Apache License, Version 2.0
// http://www.apache.org/licenses/LICENSE-2.0

"@

$headerCount = 0

Get-ChildItem -Path $OutDirFull -Recurse -Include "*.go" -File | ForEach-Object {
    $content = Get-Content $_.FullName -Raw

    if ($content -notmatch "Apache License, Version 2.0") {
        $newContent = $licenseHeader + $content
        Set-Content $_.FullName $newContent -NoNewline
        $headerCount++
    }
}

Write-Success "Headers adicionados a $headerCount arquivos Go"

# ============================================================================
# 5. Validações de compliance
# ============================================================================

Write-Step "Executando validações de compliance"

$violations = @()

# Validar dependências bloqueadas
foreach ($dep in $blocklistDeps) {
    $pattern = [regex]::Escape($dep) -replace '\\\*', '.*'
    $hits = Select-String -Path "$OutDirFull\**\*.go" -Pattern $pattern -ErrorAction SilentlyContinue

    if ($hits) {
        $violations += "Dependência bloqueada encontrada: $dep em $($hits.Count) arquivo(s)"
    }
}

# Validar padrões de arquivos bloqueados
$blockedPatterns = @("configs/prod/**", "internal/enterprise/**", "**/.env.prod")
foreach ($pattern in $blockedPatterns) {
    $regexPattern = $pattern -replace '\*\*', '*' -replace '\*', '*'
    $hits = Get-ChildItem -Path $OutDirFull -Recurse -Include $regexPattern -ErrorAction SilentlyContinue

    if ($hits) {
        $violations += "Arquivo bloqueado encontrado: $pattern"
    }
}

if ($violations.Count -gt 0) {
    Write-Failure "Falhas de compliance detectadas:"
    foreach ($v in $violations) {
        Write-Host "  - $v" -ForegroundColor Red
    }
    Write-Host "`nAbortando release por motivos de segurança." -ForegroundColor Red
    exit 1
} else {
    Write-Success "Todas as validações de compliance passaram"
}

# ============================================================================
# 6. Gerar changelog parcial
# ============================================================================

Write-Step "Gerando changelog parcial"

try {
    $latestTag = git describe --tags --abbrev=0 2>$null
    if ($latestTag) {
        $changelog = git log --oneline "$latestTag..HEAD"
        $changelogInfo = "Mudanças desde $latestTag"
    } else {
        $changelog = git log --oneline
        $changelogInfo = "Histórico completo (primeira release)"
    }

    $changelogPath = Join-Path $OutDirFull "CHANGELOG_PARTIAL.txt"
    $changelogHeader = @"
# Changelog Parcial - Release Público
$changelogInfo

"@

    ($changelogHeader + ($changelog -join "`n")) | Out-File $changelogPath -Encoding UTF8
    Write-Success "Changelog gerado: CHANGELOG_PARTIAL.txt"
} catch {
    Write-Warning "Não foi possível gerar changelog: $_"
}

# ============================================================================
# 7. Preparar repositório público
# ============================================================================

Write-Step "Preparando repositório Git público"

Push-Location $OutDirFull

try {
    git init | Out-Null
    git remote add origin $PublicRepoUrl
    git add .
    git commit -m "public: automated scrub and release preparation

- Removed proprietary code and enterprise modules
- Redacted secrets and internal URLs
- Added Apache 2.0 license headers
- Generated from internal commit: $(git -C $Root rev-parse --short HEAD)

🤖 Generated by Vertikon Release Pipeline"

    Write-Success "Repositório Git inicializado e commit criado"

    # ============================================================================
    # 8. Tag e release
    # ============================================================================

    Write-Step "Criando tag de versão"

    if ([string]::IsNullOrWhiteSpace($Version)) {
        $Version = (Get-Date -Format "yyyy.MM.dd") + ".0"
        Write-Warning "Versão não especificada, usando: $Version"
    }

    $tagName = "v$Version"
    $tagMessage = @"
Public release $tagName

Release automatizado do MCP-Ultra

Commit origem: $(git -C $Root rev-parse HEAD)
Data: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

🤖 Generated by Vertikon Release Pipeline
"@

    git tag -a $tagName -m $tagMessage
    Write-Success "Tag criada: $tagName"

    # ============================================================================
    # 9. Push (se não for dry-run)
    # ============================================================================

    if ($DryRun) {
        Write-Warning "DRY RUN MODE - Não fazendo push para repositório remoto"
        Write-Host "`nComandos que seriam executados:" -ForegroundColor Yellow
        Write-Host "  git branch -M public-release" -ForegroundColor Gray
        Write-Host "  git push -u origin public-release" -ForegroundColor Gray
        Write-Host "  git push origin $tagName" -ForegroundColor Gray
    } else {
        Write-Step "Publicando no repositório remoto"

        git branch -M public-release
        git push -u origin public-release
        git push origin $tagName

        Write-Success "Código publicado com sucesso!"
        Write-Host "`nRepositório público: $PublicRepoUrl" -ForegroundColor Cyan
        Write-Host "Tag: $tagName" -ForegroundColor Cyan
    }

} catch {
    Write-Failure "Erro durante preparação do repositório: $_"
    Pop-Location
    exit 1
} finally {
    Pop-Location
}

# ============================================================================
# Relatório Final
# ============================================================================

Write-Host "`n" + ("=" * 70) -ForegroundColor Cyan
Write-Host "RELEASE PÚBLICO - RELATÓRIO FINAL" -ForegroundColor Cyan
Write-Host ("=" * 70) -ForegroundColor Cyan

Write-Host "`nDiretório de saída: " -NoNewline
Write-Host $OutDirFull -ForegroundColor Yellow

Write-Host "Versão: " -NoNewline
Write-Host $tagName -ForegroundColor Yellow

Write-Host "Modo: " -NoNewline
if ($DryRun) {
    Write-Host "DRY RUN (simulação)" -ForegroundColor Yellow
} else {
    Write-Host "PRODUÇÃO (publicado)" -ForegroundColor Green
}

Write-Host "`nEstatísticas:"
Write-Host "  - Arquivos copiados: $copiedCount" -ForegroundColor Gray
Write-Host "  - Arquivos excluídos: $excludedCount" -ForegroundColor Gray
Write-Host "  - Arquivos sanitizados: $sanitizedCount" -ForegroundColor Gray
Write-Host "  - Arquivos com redação: $redactedCount" -ForegroundColor Gray
Write-Host "  - Headers de licença: $headerCount" -ForegroundColor Gray

Write-Host "`n✅ Pipeline concluída com sucesso!" -ForegroundColor Green
Write-Host ("=" * 70) -ForegroundColor Cyan
