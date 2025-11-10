#!/usr/bin/env bash
#
# Pipeline de release público do MCP-Ultra (versão Bash)
#
# Uso:
#   ./tools/vertikon-release.sh <public_repo_url> [out_dir] [version] [--dry-run]
#
# Exemplo:
#   ./tools/vertikon-release.sh https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git ./public 1.0.0
#   ./tools/vertikon-release.sh https://github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm.git ./public 1.0.0 --dry-run

set -euo pipefail

# ============================================================================
# Configuração
# ============================================================================

PUBLIC_REPO_URL="${1:-}"
OUT_DIR="${2:-./public}"
VERSION="${3:-}"
DRY_RUN=false

# Verificar argumentos
if [[ -z "$PUBLIC_REPO_URL" ]]; then
  echo "❌ Uso: $0 <public_repo_url> [out_dir] [version] [--dry-run]"
  exit 1
fi

# Verificar flag dry-run
for arg in "$@"; do
  if [[ "$arg" == "--dry-run" ]]; then
    DRY_RUN=true
  fi
done

# ============================================================================
# Funções Auxiliares
# ============================================================================

print_step() {
  echo ""
  echo "==> $1"
}

print_success() {
  echo "✓ $1"
}

print_warning() {
  echo "⚠ $1"
}

print_failure() {
  echo "✗ $1"
  exit 1
}

# ============================================================================
# Inicialização
# ============================================================================

print_step "Inicializando pipeline de release público MCP-Ultra"

ROOT="$(pwd)"
OUT_DIR_FULL="$ROOT/$OUT_DIR"

# Validar git
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  print_failure "Diretório atual não é um repositório Git"
fi

print_success "Repositório Git válido"

# ============================================================================
# Configuração de exclusão e sanitização
# ============================================================================

EXCLUDE_PATTERNS=(
  ".git/**"
  "internal/enterprise/**"
  "pkg/enterprise/**"
  "configs/prod/**"
  "vendor/**"
  "tests/integration/**"
  "**/.env"
  "**/.env.*"
  "**/secrets/**"
  "**/*.pem"
  "**/*.key"
  "**/id_rsa*"
  ".backup_*/**"
  "docs/gaps/**"
  "docs/melhorias/**"
  "**/analyze_gaps.ps1"
)

# Padrões de sanitização (pattern|replace)
SANITIZE_PATTERNS=(
  "vertikon.internal|localhost"
  "vertikon-private|example-org"
  "E:\\\\vertikon|/workspace"
  "E:\\\\rfesta|/config"
)

# Padrões de redação (regex)
REDACT_PATTERNS=(
  "(?i)api[_-]?key\s*[:=]\s*['\"][A-Za-z0-9_\-]{16,}['\"]"
  "(?i)secret\s*[:=]\s*['\"][A-Za-z0-9_\-]{12,}['\"]"
  "(?i)password\s*[:=]\s*['\"][^'\"]{6,}['\"]"
  "(?i)token\s*[:=]\s*['\"][A-Za-z0-9_\-]{20,}['\"]"
  "(?i)wss?://[a-zA-Z0-9.-]*vertikon[a-zA-Z0-9.-]*\.[a-z]{2,}"
  "(?i)https?://[a-zA-Z0-9.-]*vertikon[a-zA-Z0-9.-]*\.[a-z]{2,}"
  "postgresql://[^\s@]+@[^\s/]+/[^\s?]+"
  "redis://[^\s@]+@[^\s/]+"
  "nats://[^\s@]+@[^\s/]+"
  "Bearer [A-Za-z0-9_\-\.]{20,}"
)

BLOCKLIST_DEPS=(
  "github.com/vertikon-private/"
  "github.com/vertikon/internal-"
  "gitlab.vertikon.com/"
)

# ============================================================================
# 1. Preparar diretório limpo
# ============================================================================

print_step "Preparando diretório de saída: $OUT_DIR"

if [[ -d "$OUT_DIR_FULL" ]]; then
  rm -rf "$OUT_DIR_FULL"
  print_success "Diretório anterior removido"
fi

mkdir -p "$OUT_DIR_FULL"
print_success "Diretório criado: $OUT_DIR_FULL"

# ============================================================================
# 2. Copiar arquivos (respeitando excludes)
# ============================================================================

print_step "Copiando arquivos (aplicando filtros de exclusão)"

copied_count=0
excluded_count=0

while IFS= read -r file; do
  skip=false

  # Verificar se arquivo deve ser excluído
  for pattern in "${EXCLUDE_PATTERNS[@]}"; do
    # Converter glob para regex básico
    regex_pattern="${pattern//\*\*/.*}"
    regex_pattern="${regex_pattern//\*/[^/]*}"

    if [[ "$file" =~ $regex_pattern ]]; then
      skip=true
      ((excluded_count++))
      break
    fi
  done

  if [[ "$skip" == false ]]; then
    mkdir -p "$OUT_DIR_FULL/$(dirname "$file")"
    cp "$file" "$OUT_DIR_FULL/$file"
    ((copied_count++))
  fi
done < <(git ls-files)

print_success "Arquivos copiados: $copied_count | Excluídos: $excluded_count"

# ============================================================================
# 3. Sanitizar exemplos e redigir segredos
# ============================================================================

print_step "Sanitizando conteúdo (redação de segredos e substituição de padrões)"

processed_files=0
sanitized_count=0
redacted_count=0

while IFS= read -r -d '' file; do
  # Pular binários
  if file "$file" | grep -qi "text"; then
    original_hash=$(md5sum "$file" 2>/dev/null | cut -d' ' -f1 || echo "")

    # Aplicar sanitização
    for pattern in "${SANITIZE_PATTERNS[@]}"; do
      IFS='|' read -r search replace <<< "$pattern"
      if command -v perl >/dev/null 2>&1; then
        perl -i -pe "s/\Q$search\E/$replace/g" "$file"
      else
        sed -i "s|$search|$replace|g" "$file"
      fi
    done

    # Aplicar redação
    redacted=false
    for regex in "${REDACT_PATTERNS[@]}"; do
      if command -v perl >/dev/null 2>&1; then
        if perl -0777 -pe "s/$regex/REDACTED/g" "$file" > "$file.tmp"; then
          if ! cmp -s "$file" "$file.tmp"; then
            mv "$file.tmp" "$file"
            redacted=true
          else
            rm -f "$file.tmp"
          fi
        fi
      fi
    done

    # Verificar se houve alterações
    new_hash=$(md5sum "$file" 2>/dev/null | cut -d' ' -f1 || echo "")
    if [[ "$original_hash" != "$new_hash" ]]; then
      ((sanitized_count++))
    fi
    if [[ "$redacted" == true ]]; then
      ((redacted_count++))
    fi

    ((processed_files++))
  fi
done < <(find "$OUT_DIR_FULL" -type f -print0)

print_success "Arquivos processados: $processed_files | Sanitizados: $sanitized_count | Redigidos: $redacted_count"

# ============================================================================
# 4. Adicionar headers de licença
# ============================================================================

print_step "Adicionando headers de licença Apache 2.0"

LICENSE_HEADER="// Copyright (c) 2025 Vertikon
// Licensed under the Apache License, Version 2.0
// http://www.apache.org/licenses/LICENSE-2.0

"

header_count=0

while IFS= read -r -d '' gofile; do
  if ! grep -q "Apache License, Version 2.0" "$gofile"; then
    echo "$LICENSE_HEADER$(cat "$gofile")" > "$gofile"
    ((header_count++))
  fi
done < <(find "$OUT_DIR_FULL" -type f -name '*.go' -print0)

print_success "Headers adicionados a $header_count arquivos Go"

# ============================================================================
# 5. Validações de compliance
# ============================================================================

print_step "Executando validações de compliance"

violations=()

# Validar dependências bloqueadas
for dep in "${BLOCKLIST_DEPS[@]}"; do
  if grep -r "$dep" "$OUT_DIR_FULL" --include="*.go" >/dev/null 2>&1; then
    violations+=("Dependência bloqueada encontrada: $dep")
  fi
done

# Validar padrões de arquivos bloqueados
if [[ -d "$OUT_DIR_FULL/configs/prod" ]]; then
  violations+=("Diretório configs/prod/ não deveria existir")
fi

if [[ -d "$OUT_DIR_FULL/internal/enterprise" ]]; then
  violations+=("Diretório internal/enterprise/ não deveria existir")
fi

if [[ ${#violations[@]} -gt 0 ]]; then
  print_failure "Falhas de compliance detectadas:"
  for v in "${violations[@]}"; do
    echo "  - $v"
  done
  exit 1
else
  print_success "Todas as validações de compliance passaram"
fi

# ============================================================================
# 6. Gerar changelog parcial
# ============================================================================

print_step "Gerando changelog parcial"

if latest_tag=$(git describe --tags --abbrev=0 2>/dev/null); then
  changelog=$(git log --oneline "$latest_tag..HEAD")
  changelog_info="Mudanças desde $latest_tag"
else
  changelog=$(git log --oneline)
  changelog_info="Histórico completo (primeira release)"
fi

cat > "$OUT_DIR_FULL/CHANGELOG_PARTIAL.txt" <<EOF
# Changelog Parcial - Release Público
$changelog_info

$changelog
EOF

print_success "Changelog gerado: CHANGELOG_PARTIAL.txt"

# ============================================================================
# 7. Preparar repositório público
# ============================================================================

print_step "Preparando repositório Git público"

pushd "$OUT_DIR_FULL" >/dev/null

git init
git remote add origin "$PUBLIC_REPO_URL"
git add .

COMMIT_MSG="public: automated scrub and release preparation

- Removed proprietary code and enterprise modules
- Redacted secrets and internal URLs
- Added Apache 2.0 license headers
- Generated from internal commit: $(git -C "$ROOT" rev-parse --short HEAD)

🤖 Generated by Vertikon Release Pipeline"

git commit -m "$COMMIT_MSG"

print_success "Repositório Git inicializado e commit criado"

# ============================================================================
# 8. Tag e release
# ============================================================================

print_step "Criando tag de versão"

if [[ -z "$VERSION" ]]; then
  VERSION="$(date +%Y.%m.%d).0"
  print_warning "Versão não especificada, usando: $VERSION"
fi

TAG_NAME="v$VERSION"
TAG_MSG="Public release $TAG_NAME

Release automatizado do MCP-Ultra

Commit origem: $(git -C "$ROOT" rev-parse HEAD)
Data: $(date +'%Y-%m-%d %H:%M:%S')

🤖 Generated by Vertikon Release Pipeline"

git tag -a "$TAG_NAME" -m "$TAG_MSG"
print_success "Tag criada: $TAG_NAME"

# ============================================================================
# 9. Push (se não for dry-run)
# ============================================================================

if [[ "$DRY_RUN" == true ]]; then
  print_warning "DRY RUN MODE - Não fazendo push para repositório remoto"
  echo ""
  echo "Comandos que seriam executados:"
  echo "  git branch -M public-release"
  echo "  git push -u origin public-release"
  echo "  git push origin $TAG_NAME"
else
  print_step "Publicando no repositório remoto"

  git branch -M public-release
  git push -u origin public-release
  git push origin "$TAG_NAME"

  print_success "Código publicado com sucesso!"
  echo ""
  echo "Repositório público: $PUBLIC_REPO_URL"
  echo "Tag: $TAG_NAME"
fi

popd >/dev/null

# ============================================================================
# Relatório Final
# ============================================================================

echo ""
echo "======================================================================"
echo "RELEASE PÚBLICO - RELATÓRIO FINAL"
echo "======================================================================"
echo ""
echo "Diretório de saída: $OUT_DIR_FULL"
echo "Versão: $TAG_NAME"

if [[ "$DRY_RUN" == true ]]; then
  echo "Modo: DRY RUN (simulação)"
else
  echo "Modo: PRODUÇÃO (publicado)"
fi

echo ""
echo "Estatísticas:"
echo "  - Arquivos copiados: $copied_count"
echo "  - Arquivos excluídos: $excluded_count"
echo "  - Arquivos sanitizados: $sanitized_count"
echo "  - Arquivos com redação: $redacted_count"
echo "  - Headers de licença: $header_count"
echo ""
echo "✅ Pipeline concluída com sucesso!"
echo "======================================================================"
