# Patch v11.1 — Test Sync + Lint Modernization

Data: 2025-10-17

## Conteúdo
- `.golangci.yml` atualizado (schema moderno v1.61+)
- Stubs mínimos de Compliance (`ProcessDataAccessRequest`, `AnonymizeData`, `LogAuditEvent`)
- Exemplo de interface `TaskRepository` com `List(...)([]*Task, int, error)` e `Exists(...)`
- Script `scripts/regenerate_mocks.sh` para gerar mocks com GoMock
- Workflow CI (`.github/workflows/lint-and-test.yml`)
- Alvos Makefile (`lint`, `coverage-html`, `mocks`)

## Como aplicar

1) **Backup** do seu `.golangci.yml` atual (se existir).
2) Copie os arquivos deste patch para o seu repositório:
   - `.golangci.yml`
   - `internal/compliance/framework_stubs.go`  (marca de build `testpatch` opcional)
   - `internal/services/task_repository_example.go` (apenas referência)
   - `scripts/regenerate_mocks.sh` (dar permissão de execução)
   - `.github/workflows/lint-and-test.yml`
   - `Makefile` (merge com o seu, se já existir)
3) Rode:
   ```bash
   chmod +x scripts/regenerate_mocks.sh
   make mocks
   make lint
   make test
   make coverage-html
   ```
4) Ajuste seus testes para usar os novos tipos/assinaturas. Onde necessário, atualize imports para as **facades**.
5) Faça commit:
   ```bash
   git add -A
   git commit -m "feat(patch v11.1): test sync + lint modernization"
   ```

## Notas
- Os stubs de compliance são **temporários**; substitua pela implementação real.
- A flag de build `//go:build testpatch` permite desligar os stubs em produção.