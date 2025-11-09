# MIGRATION - Aplicando a IA no MCP-Ultra (sem tocar SDK)

## Passo a passo

### 1. Copiar estrutura
```bash
# Copiar ai/ para o repositorio do MCP-Ultra
cp -r ai/ E:/vertikon/business/SaaS/templates/mcp-ultra-wasm/templates/ai/
```

### 2. Mesclar .env
Adicione os blocos de `examples/.env.mcp.example` no `.env.example` do MCP:
- Feature flags (ENABLE_AI, AI_MODE, AI_CANARY_PERCENT, AI_ROUTER)
- Limites (AI_MAX_TOKENS, AI_TIMEOUT_MS, AI_MAX_RPM, AI_MAX_TPD)
- Budgets (AI_BUDGET_DAILY_BRLCAP)
- Providers (OPENAI_API_KEY, QWEN_API_KEY, LOCAL_ENDPOINT)
- Telemetry (OTEL_EXPORTER_OTLP_ENDPOINT, PROMETHEUS_METRICS_PORT)
- NATS (NATS_URL, NATS_STREAM, NATS_SUBJECT_PREFIX)

### 3. Atualizar Inventario
Adicione o campo `ai` no registro do Inventario (use `examples/inventory-registry.example.json` como referencia).

### 4. DRY-RUN (sem provider real)
```bash
ENABLE_AI=true
PROVIDER_PRIMARY=local
AI_CANARY_PERCENT=0

# Valide:
# - Policies (pii/profanity) -> policy.block event
# - Router decisions -> router.decision event
# - Budgets -> on_breach (degrade/block/alert_only)
# - Canary = 0 (sem trafego IA real)
```

### 5. Producao controlada
```bash
AI_CANARY_PERCENT=5
# Habilitar provider real via Secret Manager (NUNCA no .env commitado)
```

### 6. Observabilidade
- Confirme metricas Prometheus em `/metrics`
- Verifique spans OTEL (`policy.pre`, `inference.call`, `policy.post`)
- Monitore eventos NATS:
  - `ultra.ai.router.decision`
  - `ultra.ai.policy.block`
  - `ultra.ai.inference.summary`

## Rollback
```bash
ENABLE_AI=false
# ou
AI_CANARY_PERCENT=0
```

## Checklist de validacao

- [ ] Estrutura `ai/` copiada
- [ ] Blocos IA no `.env.example`
- [ ] Campo `ai` no Inventario
- [ ] DRY-RUN executado com sucesso
- [ ] Metricas Prometheus disponiveis
- [ ] Spans OTEL configurados
- [ ] Eventos NATS publicados
- [ ] Budgets configurados e testados
- [ ] Canary com 0% (ou 5% controlado)
