# Patch v11.3 — Lint Clean (Definitivo)

Data: 2025-10-18

Este patch entrega os **tipos mínimos** requeridos pelos testes e um **router exemplo com build tag** para você alinhar a assinatura sem quebrar produção.

## Conteúdo
- `internal/domain/dto.go` — DTOs: CreateTaskRequest, UpdateTaskRequest, TaskFilters, Task, TaskList
- `internal/services/health.go` — HealthStatus e HealthChecker
- `internal/handlers/http/router_example.go` — Assinatura canônica com build tag (`example|testpatch`)
- `scripts/open_errcheck_in_code.ps1` — Abre ocorrências do errcheck no VS Code
- `Makefile.additions.v113` — Alvos `errcheck-list` e `lint-fix`

## Como aplicar
1) Copie os arquivos para suas pastas correspondentes.
2) Se você usa `pkg/types` para UUID, ajuste `internal/domain/dto.go` (troque import de uuid).
3) Compare **a assinatura do seu router real** com a de `router_example.go` e alinhe o **teste** para a mesma assinatura.
4) Rode:
   ```bash
   make lint-fix
   pwsh ./scripts/open_errcheck_in_code.ps1
   # Trate cada ocorrência listada
   make test
   ```

## Observações
- O arquivo `router_example.go` tem **build tag** para **não** interferir no build de produção. Use apenas como referência ou para testes locais (`-tags example`).

Com isso, você zera os **undefined de tipos** e fica só com os pontos de `errcheck` (rápidos de tratar).
