# Metricas Prometheus - IA

## Inferencias
- ai_inference_requests_total{mcp_id,sdk_name,tenant_id,provider,model}
- ai_inference_latency_ms{mcp_id,sdk_name,tenant_id,provider,model}
- ai_tokens_in_total{mcp_id,sdk_name,tenant_id}
- ai_tokens_out_total{mcp_id,sdk_name,tenant_id}
- ai_cost_brl_total{mcp_id,sdk_name,tenant_id}

## Policies
- ai_policy_blocks_total{mcp_id,sdk_name,rule,severity}

## Router
- ai_router_decisions_total{mcp_id,sdk_name,provider,model,reason}

## Budgets
- ai_budget_breaches_total{scope}  # global|tenant|mcp
- ai_budget_remaining_brl{scope,tenant_id,mcp_id}
