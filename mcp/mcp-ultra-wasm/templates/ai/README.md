# MCP-Ultra - AI Turbo Update

Este pacote adiciona **semente de IA** e metadados no **MCP-Ultra** para que qualquer MCP ja nasca **IA-ready** com custo zero quando desativado.

## Como usar

1. Copie a pasta `ai/` inteira para o repositorio do **MCP-Ultra** (ex.: `mcp-ultra-wasm/templates/ai/`)
2. Mescle o bloco IA do `.env.example` do MCP (veja `ai/examples/.env.mcp.example`)
3. Atualize o **Inventario** com o campo `ai` (veja `ai/examples/inventory-registry.example.json`)
4. Rode DRY-RUN: `ENABLE_AI=true` e `PROVIDER_PRIMARY=local`
5. Se OK, ajuste budgets e canary (ex.: 5%)

## Politica de ativacao

- **Opt-in**: IA so liga com `ENABLE_AI=true`
- **Sem custo**: providers reais exigem chaves em Secret Manager; aqui ha apenas placeholders
- **Observabilidade**: metricas e spans so aparecem quando IA esta ativa

## Estrutura

- `feature_flags.json` - flags padrao de IA
- `config/*` - router, policies, guardrails, budgets
- `nats-schemas/*` - eventos de decisao, bloqueio e erros de inferencia
- `telemetry/*` - metricas Prometheus e OTEL example
- `examples/*` - `.env` do MCP e registro de inventario
- `MIGRATION.md` - roteiro para aplicar em MCPs existentes

## Proximos passos

1. Implementar Go handlers para router, policies e budgets
2. Integrar com NATS para eventos
3. Adicionar spans OTEL para observabilidade
4. Criar testes DRY-RUN
