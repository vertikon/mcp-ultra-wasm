# Quick Push - Script simples para publicar no GitHub
# Sem sanitizacao, apenas git add/commit/push

param(
    [string]$GitHubOrg = "vertikon",
    [string]$RepoName = "sdk-ultra-wasm",
    [string]$Branch = "main"
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "   QUICK PUSH - Publicacao Rapida no GitHub" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

# Ir para o diretorio do SDK
$SDK = "E:\vertikon\business\SaaS\templates\sdk-ultra-wasm"
Set-Location $SDK

# Verificar se e repositorio git
if (-not (Test-Path ".git")) {
    Write-Host "Inicializando repositorio Git..." -ForegroundColor Yellow
    git init
}

# Verificar/criar branch
$currentBranch = git branch --show-current 2>$null
if ([string]::IsNullOrEmpty($currentBranch)) {
    Write-Host "Criando branch $Branch..." -ForegroundColor Yellow
    git checkout -b $Branch
} elseif ($currentBranch -ne $Branch) {
    git checkout -b $Branch 2>$null
    if ($LASTEXITCODE -ne 0) {
        git checkout $Branch
    }
}

# Criar README.md basico se nao existir
if (-not (Test-Path "README.md")) {
    Write-Host "Criando README.md basico..." -ForegroundColor Yellow
    "# sdk-ultra-wasm`n`nMCP-Ultra SDK Custom v9.0.0" | Out-File -FilePath "README.md" -Encoding UTF8
}

# Git add
Write-Host "Adicionando arquivos..." -ForegroundColor White
git add .

# Git status
Write-Host ""
Write-Host "Arquivos a serem commitados:" -ForegroundColor Cyan
git status --short

Write-Host ""
$continue = Read-Host "Continuar com o commit? (S/N)"
if ($continue -ne "S" -and $continue -ne "s") {
    Write-Host "Operacao cancelada" -ForegroundColor Yellow
    exit 0
}

# Git commit
Write-Host ""
Write-Host "Criando commit..." -ForegroundColor White
$commitMsg = "feat: initial commit sdk-ultra-wasm v9.0.0"
git commit -m $commitMsg

# Verificar/adicionar remote
$remoteUrl = git remote get-url origin 2>$null
if ([string]::IsNullOrEmpty($remoteUrl)) {
    Write-Host "Adicionando remote origin..." -ForegroundColor White
    git remote add origin "https://github.com/$GitHubOrg/$RepoName.git"
} else {
    Write-Host "Remote ja configurado: $remoteUrl" -ForegroundColor Green
}

# Git push
Write-Host ""
Write-Host "Fazendo push para GitHub..." -ForegroundColor White
Write-Host ""

git push -u origin $Branch

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "================================================================" -ForegroundColor Green
    Write-Host "   SUCESSO! Repositorio publicado" -ForegroundColor Green
    Write-Host "================================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "URL: https://github.com/$GitHubOrg/$RepoName" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "[ERRO] Falha no push" -ForegroundColor Red
    Write-Host ""
    Write-Host "Verifique:" -ForegroundColor Yellow
    Write-Host "  1. Repositorio existe no GitHub?" -ForegroundColor White
    Write-Host "  2. Credenciais do Git configuradas?" -ForegroundColor White
    Write-Host "  3. Permissao para push?" -ForegroundColor White
    Write-Host ""
    exit 1
}
