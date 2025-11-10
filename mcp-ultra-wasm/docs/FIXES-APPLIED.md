# 🔧 Correções Aplicadas — MCP Ultra
**Data:** 2025-09-23
**Objetivo:** Resolver erros de dependências em Node.js e Go de forma **repetível e idempotente**.

## 🎯 Problemas Resolvidos

### 1) Go: `malformed module path` (semconv)
- **Sintoma:** `require go.opentelemetry.io/otel/semconv/vX.Y.Z vX.Y.Z` no `go.mod` quebra o `go mod tidy/build`.
- **Causa:** `semconv` usa **sufixo de versão no *import path***, mas **não** deve aparecer como módulo separado no `require`.
- **Correção:** Removido `require .../semconv/...` dos `go.mod` e fixado `go.opentelemetry.io/otel@v1.38.0` + `go mod tidy`.

### 2) Go: interferência de `go.work`
- **Sintoma:** build tentando resolver caminhos relativos de outros MCPs (ex.: `..\..\mcp-crm-vendas`).
- **Correção:** compilação do utilitário de autocommit com `GOWORK=off` (build isolado do workspace).

### 3) Node/NPM: builds inconsistentes
- **Sintoma:** divergência de dependências em ambientes diferentes.
- **Correção:** `npm ci` (quando há `package-lock.json`) ou `npm install` (fallback) + `npm run build`.

### 4) Build do autocommit
- **Sintoma:** falha do binário interrompia o setup.
- **Correção:** compilação **resiliente** (não bloqueia pipeline se falhar).

## 📁 Arquivos Criados/Alterados
- `scripts/fix-deps.ps1` — normaliza Node/Go, usa `GOWORK=off`, limpa `semconv`, compila autocommit.
- `scripts/setup-complete.ps1` — **hook idempotente** que chama `fix-deps.ps1` ao final do setup.
- `scripts/test-fixes.ps1` — verificação automatizada (Node build + checagens Go).
- `.github/workflows/fix-deps.yml` — CI para validar em cada PR/push.
- `FIXES-APPLIED.md` — este documento.

## 🚀 Como usar
```powershell
# Correção de dependências (pode rodar quantas vezes quiser)
.\scripts\fix-deps.ps1

# Validação automatizada
.\scripts\test-fixes.ps1
```

## ✅ Critérios de Pronto

- `npm ci && npm run build` finaliza sem erro (no diretório do servidor MCP).
- Nenhum `go.mod` contém `require .../semconv/...`.
- `go mod tidy/go list` executam sem erro com `GOWORK=off`.
- (Opcional) `autocommit.exe` compilado; falha não bloqueia pipeline.

## 🔁 Aplicar em MCPs existentes (recomendado)

Abra PR adicionando:

- `scripts/fix-deps.ps1`
- `scripts/test-fixes.ps1`
- `.github/workflows/fix-deps.yml`

Hook no final de `scripts/setup-complete.ps1`:

```powershell
try {
  $repoRoot = Split-Path $PSScriptRoot -Parent
  & (Join-Path $PSScriptRoot 'fix-deps.ps1') -RepoRoot $repoRoot
} catch {
  Write-Warning ('fix-deps falhou: ' + $_.Exception.Message)
}
```

## 🧪 CI (opcional, mas recomendado)

Veja `.github/workflows/fix-deps.yml` para validar em cada PR/push.