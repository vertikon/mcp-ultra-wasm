# Message Schemas

JSON schemas para validação de mensagens NATS.

## Schemas Disponíveis

- ultra.base.event.v1.json - Evento base
- ultra.health.ping.v1.json - Health ping
- ultra.ai.router.decision.v1.json - Decisão de roteamento IA
- ultra.ai.policy.block.v1.json - Bloqueio de política IA

## Uso

```go
import (
    "encoding/json"
    "os"
    "github.com/xeipuuv/gojsonschema"
)

// Carregar schema
schemaLoader := gojsonschema.NewReferenceLoader("file://./internal/schemas/ultra.base.event.v1.json")

// Validar documento
documentLoader := gojsonschema.NewStringLoader(jsonString)
result, err := gojsonschema.Validate(schemaLoader, documentLoader)
```
