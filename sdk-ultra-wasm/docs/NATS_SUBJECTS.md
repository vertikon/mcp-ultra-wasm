# NATS Subjects — sdk-ultra-wasm

**Namespace:** `mcp.ultra.sdk.custom`

## Subjects Documentados

| Subject | Descrição | Tipo |
|---------|-----------|------|
| `mcp.ultra.sdk.custom.health.ping` | Ping/Pong para validação de conectividade | Request/Reply |
| `mcp.ultra.sdk.custom.seed.validate` | Validação de sementes customizadas | Request/Reply |
| `mcp.ultra.sdk.custom.template.sync` | Sincronização de templates | Pub/Sub |
| `mcp.ultra.sdk.custom.sdk.check` | Verificação de compatibilidade do SDK | Request/Reply |

## Padrão de Nomenclatura

Todos os subjects seguem o padrão:

```
mcp.ultra.sdk.custom.<domain>.<action>
```

### Exemplos:

- `mcp.ultra.sdk.custom.plugin.register` - Registro de plugins
- `mcp.ultra.sdk.custom.plugin.unregister` - Remoção de plugins
- `mcp.ultra.sdk.custom.middleware.apply` - Aplicação de middleware
- `mcp.ultra.sdk.custom.route.add` - Adição de rotas

## Contratos

### Health Ping

**Subject:** `mcp.ultra.sdk.custom.health.ping`

**Request:**
```json
{
  "timestamp": "2025-10-05T20:30:00Z"
}
```

**Response:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2025-10-05T20:30:00Z"
}
```

### Seed Validate

**Subject:** `mcp.ultra.sdk.custom.seed.validate`

**Request:**
```json
{
  "seed_name": "waba",
  "version": "1.0.0"
}
```

**Response:**
```json
{
  "valid": true,
  "errors": [],
  "warnings": []
}
```

## Configuração

Para usar NATS no SDK, configure as seguintes variáveis de ambiente:

```bash
NATS_URL=nats://localhost:4222
NATS_CLUSTER_ID=mcp-cluster
NATS_CLIENT_ID=sdk-ultra-wasm
```
