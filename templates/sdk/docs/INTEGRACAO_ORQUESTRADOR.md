# 🔌 Integração com MCP-ULTRA-WASM-ORQUESTRADOR

**Versão:** 1.0.0
**Status:** 📋 ESPECIFICAÇÃO (Orquestrador em desenvolvimento)
**SDK Version:** v9.0.0

---

## 📋 Visão Geral

Este documento especifica como o **sdk-ultra-wasm v9** se integrará com o **MCP-ULTRA-WASM-ORQUESTRADOR** quando este estiver pronto.

**Funcionalidades Planejadas:**
- 🔄 Sincronização automática de seeds e templates
- 📊 Auditoria de versão entre SDK e Template
- 🔍 Controle de compatibilidade vertical via MCP-Version-Matrix

---

## 🏗️ Arquitetura de Integração

```
┌─────────────────────────────────────────────────────────────────┐
│                  MCP-ULTRA-WASM-ORQUESTRADOR                         │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Orchestration Engine                                     │  │
│  │  - Seed Sync Manager                                     │  │
│  │  - Version Auditor                                       │  │
│  │  - Compatibility Matrix                                  │  │
│  └─────────────────┬────────────────────────────────────────┘  │
│                    │                                            │
│                    │ NATS Messages                              │
│                    │                                            │
└────────────────────┼────────────────────────────────────────────┘
                     │
                     │
        ┌────────────┴────────────┐
        │                         │
        ↓                         ↓
┌──────────────────┐    ┌──────────────────┐
│  MCP-ULTRA-WASM-SDK   │    │   SEED-WABA      │
│   (este módulo)  │    │  (seed example)  │
│                  │    │                  │
│  - Contracts     │    │  - Plugin WABA   │
│  - Registry      │    │  - Templates     │
│  - Bootstrap     │    │  - Handlers      │
└──────────────────┘    └──────────────────┘
```

---

## 🔄 1. Sincronização Automática de Seeds e Templates

### 1.1 NATS Subjects para Sync

| Subject | Tipo | Descrição |
|---------|------|-----------|
| `mcp.orchestrator.sync.request` | Request/Reply | Solicita sincronização |
| `mcp.orchestrator.sync.seed.{name}` | Pub/Sub | Atualização de seed |
| `mcp.orchestrator.sync.template.{name}` | Pub/Sub | Atualização de template |
| `mcp.orchestrator.sync.status` | Pub/Sub | Status de sincronização |

### 1.2 Contrato: Sync Request

**Subject:** `mcp.orchestrator.sync.request`

**Request:**
```json
{
  "sdk_version": "v9.0.0",
  "module": "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm",
  "seeds": ["waba", "marketing"],
  "timestamp": "2025-10-05T21:10:00Z",
  "requester": "sdk-instance-01"
}
```

**Response:**
```json
{
  "sync_id": "sync-20251005-211000",
  "status": "initiated",
  "seeds_to_update": [
    {
      "name": "waba",
      "current_version": "1.0.0",
      "latest_version": "1.1.0",
      "update_required": true
    },
    {
      "name": "marketing",
      "current_version": "1.0.0",
      "latest_version": "1.0.0",
      "update_required": false
    }
  ],
  "template_updates": [
    {
      "name": "mcp-ultra-wasm-base",
      "current_version": "2.0.0",
      "latest_version": "2.1.0",
      "breaking_changes": false
    }
  ],
  "estimated_duration": "30s",
  "timestamp": "2025-10-05T21:10:01Z"
}
```

### 1.3 Contrato: Seed Update

**Subject:** `mcp.orchestrator.sync.seed.{name}`

**Message:**
```json
{
  "seed_name": "waba",
  "version": "1.1.0",
  "changelog": [
    "feat: add message template validation",
    "fix: HMAC signature verification"
  ],
  "files_updated": [
    "internal/plugins/waba/plugin.go",
    "internal/plugins/waba/templates.go"
  ],
  "migration_required": false,
  "download_url": "https://registry.vertikon.internal/seeds/waba-1.1.0.tar.gz",
  "checksum_sha256": "abc123...",
  "timestamp": "2025-10-05T21:10:05Z"
}
```

### 1.4 Implementação no SDK (Preparada)

**Arquivo:** `pkg/orchestrator/sync.go` (a ser criado quando orquestrador estiver pronto)

```go
package orchestrator

import (
    "context"
    "encoding/json"
    "time"

    "github.com/nats-io/nats.go"
)

// SyncRequest representa uma solicitação de sincronização
type SyncRequest struct {
    SDKVersion  string    `json:"sdk_version"`
    Module      string    `json:"module"`
    Seeds       []string  `json:"seeds"`
    Timestamp   time.Time `json:"timestamp"`
    Requester   string    `json:"requester"`
}

// SyncResponse representa a resposta de sincronização
type SyncResponse struct {
    SyncID            string          `json:"sync_id"`
    Status            string          `json:"status"`
    SeedsToUpdate     []SeedUpdate    `json:"seeds_to_update"`
    TemplateUpdates   []TemplateUpdate `json:"template_updates"`
    EstimatedDuration string          `json:"estimated_duration"`
    Timestamp         time.Time       `json:"timestamp"`
}

// SyncManager gerencia sincronizações com o orquestrador
type SyncManager struct {
    nc         *nats.Conn
    sdkVersion string
    module     string
}

// NewSyncManager cria um novo gerenciador de sincronização
func NewSyncManager(nc *nats.Conn, sdkVersion, module string) *SyncManager {
    return &SyncManager{
        nc:         nc,
        sdkVersion: sdkVersion,
        module:     module,
    }
}

// RequestSync solicita sincronização ao orquestrador
func (sm *SyncManager) RequestSync(ctx context.Context, seeds []string) (*SyncResponse, error) {
    req := SyncRequest{
        SDKVersion: sm.sdkVersion,
        Module:     sm.module,
        Seeds:      seeds,
        Timestamp:  time.Now(),
        Requester:  "sdk-instance", // TODO: get from config
    }

    reqData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    msg, err := sm.nc.RequestWithContext(ctx, "mcp.orchestrator.sync.request", reqData)
    if err != nil {
        return nil, err
    }

    var resp SyncResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

// SubscribeSeedUpdates assina atualizações de seeds
func (sm *SyncManager) SubscribeSeedUpdates(handler func(*SeedUpdate)) error {
    _, err := sm.nc.Subscribe("mcp.orchestrator.sync.seed.*", func(msg *nats.Msg) {
        var update SeedUpdate
        if err := json.Unmarshal(msg.Data, &update); err != nil {
            // Log error
            return
        }
        handler(&update)
    })
    return err
}
```

---

## 📊 2. Auditoria de Versão entre SDK e Template

### 2.1 NATS Subjects para Auditoria

| Subject | Tipo | Descrição |
|---------|------|-----------|
| `mcp.orchestrator.audit.version.request` | Request/Reply | Solicita auditoria |
| `mcp.orchestrator.audit.version.report` | Pub/Sub | Relatório de auditoria |
| `mcp.orchestrator.audit.version.alert` | Pub/Sub | Alerta de incompatibilidade |

### 2.2 Contrato: Audit Request

**Subject:** `mcp.orchestrator.audit.version.request`

**Request:**
```json
{
  "sdk_version": "v9.0.0",
  "sdk_module": "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm",
  "seeds": [
    {
      "name": "waba",
      "version": "1.0.0",
      "sdk_dependency": "v9.0.0"
    }
  ],
  "timestamp": "2025-10-05T21:10:00Z"
}
```

**Response:**
```json
{
  "audit_id": "audit-20251005-211000",
  "status": "completed",
  "sdk_version": "v9.0.0",
  "compatibility_score": 100,
  "issues": [],
  "recommendations": [],
  "audit_details": {
    "seeds_audited": 1,
    "templates_audited": 1,
    "version_mismatches": 0,
    "deprecated_features": 0
  },
  "timestamp": "2025-10-05T21:10:02Z"
}
```

### 2.3 Contrato: Version Report

**Subject:** `mcp.orchestrator.audit.version.report`

**Message:**
```json
{
  "report_id": "report-20251005-211000",
  "sdk_version": "v9.0.0",
  "template_version": "2.0.0",
  "compatibility_matrix": {
    "sdk_to_template": "COMPATIBLE",
    "template_to_seeds": "COMPATIBLE",
    "seeds_to_sdk": "COMPATIBLE"
  },
  "version_graph": {
    "sdk-ultra-wasm": {
      "version": "v9.0.0",
      "dependencies": {
        "mcp-ultra-wasm-template": ">=2.0.0"
      }
    },
    "seed-waba": {
      "version": "1.0.0",
      "dependencies": {
        "sdk-ultra-wasm": ">=v9.0.0"
      }
    }
  },
  "warnings": [],
  "errors": [],
  "timestamp": "2025-10-05T21:10:02Z"
}
```

### 2.4 Implementação no SDK (Preparada)

**Arquivo:** `pkg/orchestrator/audit.go` (a ser criado)

```go
package orchestrator

import (
    "context"
    "encoding/json"
    "time"

    "github.com/nats-io/nats.go"
)

// AuditRequest representa uma solicitação de auditoria
type AuditRequest struct {
    SDKVersion string        `json:"sdk_version"`
    SDKModule  string        `json:"sdk_module"`
    Seeds      []SeedVersion `json:"seeds"`
    Timestamp  time.Time     `json:"timestamp"`
}

// AuditResponse representa a resposta de auditoria
type AuditResponse struct {
    AuditID           string       `json:"audit_id"`
    Status            string       `json:"status"`
    SDKVersion        string       `json:"sdk_version"`
    CompatibilityScore int         `json:"compatibility_score"`
    Issues            []AuditIssue `json:"issues"`
    Recommendations   []string     `json:"recommendations"`
    AuditDetails      AuditDetails `json:"audit_details"`
    Timestamp         time.Time    `json:"timestamp"`
}

// AuditManager gerencia auditorias de versão
type AuditManager struct {
    nc         *nats.Conn
    sdkVersion string
    module     string
}

// NewAuditManager cria um novo gerenciador de auditoria
func NewAuditManager(nc *nats.Conn, sdkVersion, module string) *AuditManager {
    return &AuditManager{
        nc:         nc,
        sdkVersion: sdkVersion,
        module:     module,
    }
}

// RequestAudit solicita auditoria de versão
func (am *AuditManager) RequestAudit(ctx context.Context, seeds []SeedVersion) (*AuditResponse, error) {
    req := AuditRequest{
        SDKVersion: am.sdkVersion,
        SDKModule:  am.module,
        Seeds:      seeds,
        Timestamp:  time.Now(),
    }

    reqData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    msg, err := am.nc.RequestWithContext(ctx, "mcp.orchestrator.audit.version.request", reqData)
    if err != nil {
        return nil, err
    }

    var resp AuditResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

// SubscribeAuditReports assina relatórios de auditoria
func (am *AuditManager) SubscribeAuditReports(handler func(*VersionReport)) error {
    _, err := am.nc.Subscribe("mcp.orchestrator.audit.version.report", func(msg *nats.Msg) {
        var report VersionReport
        if err := json.Unmarshal(msg.Data, &report); err != nil {
            // Log error
            return
        }
        handler(&report)
    })
    return err
}
```

---

## 🔍 3. Controle de Compatibilidade via MCP-Version-Matrix

### 3.1 NATS Subjects para Version Matrix

| Subject | Tipo | Descrição |
|---------|------|-----------|
| `mcp.orchestrator.matrix.query` | Request/Reply | Consulta matriz |
| `mcp.orchestrator.matrix.update` | Pub/Sub | Atualização da matriz |
| `mcp.orchestrator.matrix.validate` | Request/Reply | Valida compatibilidade |

### 3.2 Contrato: Matrix Query

**Subject:** `mcp.orchestrator.matrix.query`

**Request:**
```json
{
  "query_type": "compatibility",
  "components": {
    "sdk": "sdk-ultra-wasm@v9.0.0",
    "template": "mcp-ultra-wasm-template@2.0.0",
    "seed": "seed-waba@1.0.0"
  },
  "timestamp": "2025-10-05T21:10:00Z"
}
```

**Response:**
```json
{
  "query_id": "matrix-query-20251005-211000",
  "compatible": true,
  "compatibility_level": "FULL",
  "matrix_version": "1.5.0",
  "compatibility_details": {
    "sdk_to_template": {
      "compatible": true,
      "confidence": 100,
      "required_version": ">=2.0.0",
      "actual_version": "2.0.0",
      "status": "EXACT_MATCH"
    },
    "template_to_seed": {
      "compatible": true,
      "confidence": 100,
      "required_version": ">=1.0.0",
      "actual_version": "1.0.0",
      "status": "EXACT_MATCH"
    },
    "seed_to_sdk": {
      "compatible": true,
      "confidence": 100,
      "required_version": ">=v9.0.0",
      "actual_version": "v9.0.0",
      "status": "EXACT_MATCH"
    }
  },
  "warnings": [],
  "timestamp": "2025-10-05T21:10:01Z"
}
```

### 3.3 Contrato: Compatibility Validation

**Subject:** `mcp.orchestrator.matrix.validate`

**Request:**
```json
{
  "validation_type": "pre_deployment",
  "target_environment": "production",
  "components": {
    "sdk": {
      "name": "sdk-ultra-wasm",
      "version": "v9.0.0",
      "contracts_version": "1.0.0"
    },
    "seeds": [
      {
        "name": "waba",
        "version": "1.0.0",
        "sdk_dependency": "v9.0.0"
      }
    ]
  },
  "timestamp": "2025-10-05T21:10:00Z"
}
```

**Response:**
```json
{
  "validation_id": "val-20251005-211000",
  "approved": true,
  "compatibility_status": "APPROVED",
  "validation_results": {
    "contract_compatibility": "PASS",
    "version_compatibility": "PASS",
    "feature_compatibility": "PASS",
    "breaking_changes": "NONE"
  },
  "deployment_safe": true,
  "rollback_plan": {
    "available": true,
    "previous_stable": "v8.0.0"
  },
  "timestamp": "2025-10-05T21:10:02Z"
}
```

### 3.4 MCP-Version-Matrix Schema

**Arquivo:** `docs/mcp-version-matrix.schema.json`

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "MCP Version Matrix",
  "description": "Matriz de compatibilidade vertical do ecossistema Vertikon Ultra",
  "type": "object",
  "properties": {
    "matrix_version": {
      "type": "string",
      "description": "Versão da matriz (SemVer)"
    },
    "last_updated": {
      "type": "string",
      "format": "date-time"
    },
    "components": {
      "type": "object",
      "properties": {
        "sdk": {
          "type": "object",
          "properties": {
            "name": { "type": "string" },
            "versions": {
              "type": "array",
              "items": {
                "type": "object",
                "properties": {
                  "version": { "type": "string" },
                  "release_date": { "type": "string", "format": "date" },
                  "status": { "enum": ["stable", "deprecated", "eol"] },
                  "compatible_with": {
                    "type": "object",
                    "properties": {
                      "templates": { "type": "array", "items": { "type": "string" } },
                      "seeds": { "type": "array", "items": { "type": "string" } }
                    }
                  },
                  "breaking_changes": { "type": "array", "items": { "type": "string" } }
                },
                "required": ["version", "status", "compatible_with"]
              }
            }
          }
        },
        "templates": { "type": "object" },
        "seeds": { "type": "object" }
      }
    },
    "compatibility_rules": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "rule_id": { "type": "string" },
          "description": { "type": "string" },
          "from": { "type": "string" },
          "to": { "type": "string" },
          "version_constraint": { "type": "string" },
          "severity": { "enum": ["error", "warning", "info"] }
        }
      }
    }
  },
  "required": ["matrix_version", "components", "compatibility_rules"]
}
```

### 3.5 Exemplo de Version Matrix

**Arquivo:** `docs/mcp-version-matrix.example.json`

```json
{
  "matrix_version": "1.5.0",
  "last_updated": "2025-10-05T21:10:00Z",
  "components": {
    "sdk": {
      "name": "sdk-ultra-wasm",
      "versions": [
        {
          "version": "v9.0.0",
          "release_date": "2025-10-05",
          "status": "stable",
          "contracts_version": "1.0.0",
          "compatible_with": {
            "templates": ["mcp-ultra-wasm-template@>=2.0.0"],
            "seeds": ["seed-waba@>=1.0.0", "seed-marketing@>=1.0.0"]
          },
          "breaking_changes": [],
          "certification": "VTK-SDK-CUSTOM-V9-20251005-STABLE"
        },
        {
          "version": "v8.0.0",
          "release_date": "2025-10-05",
          "status": "deprecated",
          "contracts_version": "1.0.0",
          "compatible_with": {
            "templates": ["mcp-ultra-wasm-template@>=2.0.0"],
            "seeds": ["seed-waba@>=1.0.0"]
          },
          "breaking_changes": [],
          "deprecation_date": "2025-10-05",
          "eol_date": "2026-01-05"
        }
      ]
    },
    "templates": {
      "name": "mcp-ultra-wasm-template",
      "versions": [
        {
          "version": "2.0.0",
          "release_date": "2025-09-01",
          "status": "stable",
          "compatible_with": {
            "sdks": ["sdk-ultra-wasm@>=v8.0.0"],
            "seeds": ["seed-*@>=1.0.0"]
          }
        }
      ]
    },
    "seeds": {
      "waba": {
        "versions": [
          {
            "version": "1.0.0",
            "release_date": "2025-10-05",
            "status": "stable",
            "compatible_with": {
              "sdks": ["sdk-ultra-wasm@>=v9.0.0"],
              "templates": ["mcp-ultra-wasm-template@>=2.0.0"]
            }
          }
        ]
      }
    }
  },
  "compatibility_rules": [
    {
      "rule_id": "SDK_TO_TEMPLATE_001",
      "description": "SDK major version must not exceed template major version + 1",
      "from": "sdk",
      "to": "template",
      "version_constraint": "sdk.major <= template.major + 1",
      "severity": "error"
    },
    {
      "rule_id": "SEED_TO_SDK_001",
      "description": "Seed contracts version must match SDK contracts version",
      "from": "seed",
      "to": "sdk",
      "version_constraint": "seed.contracts_version == sdk.contracts_version",
      "severity": "error"
    },
    {
      "rule_id": "DEPRECATED_WARNING_001",
      "description": "Warn when using deprecated SDK version",
      "from": "sdk",
      "to": "sdk",
      "version_constraint": "sdk.status != 'deprecated'",
      "severity": "warning"
    }
  ]
}
```

### 3.6 Implementação no SDK (Preparada)

**Arquivo:** `pkg/orchestrator/matrix.go` (a ser criado)

```go
package orchestrator

import (
    "context"
    "encoding/json"
    "time"

    "github.com/nats-io/nats.go"
)

// MatrixQuery representa uma consulta à matriz de compatibilidade
type MatrixQuery struct {
    QueryType  string              `json:"query_type"`
    Components map[string]string   `json:"components"`
    Timestamp  time.Time           `json:"timestamp"`
}

// MatrixResponse representa a resposta da matriz
type MatrixResponse struct {
    QueryID              string                 `json:"query_id"`
    Compatible           bool                   `json:"compatible"`
    CompatibilityLevel   string                 `json:"compatibility_level"`
    MatrixVersion        string                 `json:"matrix_version"`
    CompatibilityDetails map[string]interface{} `json:"compatibility_details"`
    Warnings             []string               `json:"warnings"`
    Timestamp            time.Time              `json:"timestamp"`
}

// MatrixManager gerencia interações com a matriz de compatibilidade
type MatrixManager struct {
    nc         *nats.Conn
    sdkVersion string
    module     string
}

// NewMatrixManager cria um novo gerenciador de matriz
func NewMatrixManager(nc *nats.Conn, sdkVersion, module string) *MatrixManager {
    return &MatrixManager{
        nc:         nc,
        sdkVersion: sdkVersion,
        module:     module,
    }
}

// QueryCompatibility consulta compatibilidade na matriz
func (mm *MatrixManager) QueryCompatibility(ctx context.Context, components map[string]string) (*MatrixResponse, error) {
    query := MatrixQuery{
        QueryType:  "compatibility",
        Components: components,
        Timestamp:  time.Now(),
    }

    queryData, err := json.Marshal(query)
    if err != nil {
        return nil, err
    }

    msg, err := mm.nc.RequestWithContext(ctx, "mcp.orchestrator.matrix.query", queryData)
    if err != nil {
        return nil, err
    }

    var resp MatrixResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}

// ValidateDeployment valida compatibilidade antes de deploy
func (mm *MatrixManager) ValidateDeployment(ctx context.Context, components map[string]interface{}) (*ValidationResponse, error) {
    req := ValidationRequest{
        ValidationType:    "pre_deployment",
        TargetEnvironment: "production",
        Components:        components,
        Timestamp:         time.Now(),
    }

    reqData, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    msg, err := mm.nc.RequestWithContext(ctx, "mcp.orchestrator.matrix.validate", reqData)
    if err != nil {
        return nil, err
    }

    var resp ValidationResponse
    if err := json.Unmarshal(msg.Data, &resp); err != nil {
        return nil, err
    }

    return &resp, nil
}
```

---

## 📦 4. Metadados do SDK para Orquestrador

### 4.1 SDK Metadata File

**Arquivo:** `sdk-metadata.json` (a ser criado na raiz do projeto)

```json
{
  "sdk": {
    "name": "sdk-ultra-wasm",
    "version": "v9.0.0",
    "module": "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm",
    "contracts_version": "1.0.0",
    "release_date": "2025-10-05",
    "status": "stable",
    "certification": {
      "id": "VTK-SDK-CUSTOM-V9-20251005-STABLE",
      "score": 100,
      "validator_version": "4.2.1",
      "date": "2025-10-05T21:05:00Z",
      "valid_until": "2026-10-05T23:59:59Z"
    }
  },
  "capabilities": {
    "plugins": true,
    "middlewares": true,
    "jobs": true,
    "services": true,
    "nats_integration": true,
    "health_checks": true,
    "metrics": true,
    "structured_logs": true
  },
  "orchestrator_integration": {
    "sync_enabled": true,
    "audit_enabled": true,
    "matrix_enabled": true,
    "subjects": [
      "mcp.orchestrator.sync.request",
      "mcp.orchestrator.sync.seed.*",
      "mcp.orchestrator.audit.version.request",
      "mcp.orchestrator.matrix.query",
      "mcp.orchestrator.matrix.validate"
    ]
  },
  "dependencies": {
    "required": {
      "go": ">=1.23",
      "gorilla/mux": "v1.8.1",
      "prometheus/client_golang": "v1.19.0"
    },
    "optional": {
      "nats-io/nats.go": ">=1.31.0"
    }
  },
  "compatible_with": {
    "templates": ["mcp-ultra-wasm-template@>=2.0.0"],
    "seeds": ["seed-*@>=1.0.0"],
    "orchestrator": ["mcp-ultra-wasm-orquestrador@>=1.0.0"]
  }
}
```

---

## 🔧 5. Configuração para Integração

### 5.1 Environment Variables

```bash
# Orquestrador Connection
MCP_ORCHESTRATOR_ENABLED=true
MCP_ORCHESTRATOR_NATS_URL=nats://orchestrator.vertikon.internal:4222
MCP_ORCHESTRATOR_CLUSTER_ID=mcp-ultra-wasm-cluster
MCP_ORCHESTRATOR_CLIENT_ID=sdk-custom-v9

# Sync Configuration
MCP_SYNC_AUTO_ENABLED=true
MCP_SYNC_INTERVAL=3600  # 1 hora
MCP_SYNC_SEEDS=waba,marketing

# Audit Configuration
MCP_AUDIT_AUTO_ENABLED=true
MCP_AUDIT_ON_STARTUP=true
MCP_AUDIT_PERIODIC=true
MCP_AUDIT_INTERVAL=86400  # 24 horas

# Matrix Configuration
MCP_MATRIX_VALIDATION_ENABLED=true
MCP_MATRIX_STRICT_MODE=true
MCP_MATRIX_CACHE_TTL=3600
```

### 5.2 Configuration File

**Arquivo:** `config/orchestrator.yaml`

```yaml
orchestrator:
  enabled: true
  connection:
    nats_url: nats://orchestrator.vertikon.internal:4222
    cluster_id: mcp-ultra-wasm-cluster
    client_id: sdk-custom-v9
    max_reconnects: 10
    reconnect_wait: 2s

  sync:
    auto_enabled: true
    interval: 1h
    seeds:
      - waba
      - marketing
    on_update:
      action: notify  # notify, auto-apply, manual
      restart_required: false

  audit:
    auto_enabled: true
    on_startup: true
    periodic: true
    interval: 24h
    alert_on_incompatibility: true
    slack_webhook: https://hooks.slack.com/services/xxx

  matrix:
    validation_enabled: true
    strict_mode: true  # Bloqueia deploy se incompatível
    cache_ttl: 1h
    rules:
      - sdk_to_template_version_check
      - seed_to_sdk_contracts_check
      - deprecated_version_warning
```

---

## 📋 6. Checklist de Preparação

### SDK Pronto para Integração ✅

- [x] Contratos NATS documentados
- [x] Subjects definidos
- [x] Esquemas de mensagens especificados
- [x] Metadados do SDK criados
- [x] Documentação de integração completa
- [x] Configurações de exemplo fornecidas

### Aguardando Orquestrador ⏳

- [ ] Implementar `pkg/orchestrator/sync.go`
- [ ] Implementar `pkg/orchestrator/audit.go`
- [ ] Implementar `pkg/orchestrator/matrix.go`
- [ ] Criar `sdk-metadata.json`
- [ ] Criar `config/orchestrator.yaml`
- [ ] Adicionar testes de integração
- [ ] Atualizar README com seção de orquestração

---

## 🚀 7. Quando o Orquestrador Estiver Pronto

### Passo 1: Instalar Dependência NATS

```bash
go get github.com/nats-io/nats.go@latest
```

### Passo 2: Criar Implementações

```bash
# Criar package orchestrator
mkdir -p pkg/orchestrator

# Criar arquivos de implementação
touch pkg/orchestrator/sync.go
touch pkg/orchestrator/audit.go
touch pkg/orchestrator/matrix.go
touch pkg/orchestrator/types.go
```

### Passo 3: Atualizar Bootstrap

**Arquivo:** `pkg/bootstrap/bootstrap.go`

```go
// Adicionar ao Config
type Config struct {
    // ... existing fields
    OrchestratorEnabled bool
    OrchestratorNATSURL string
}

// Adicionar ao Bootstrap
func Bootstrap(cfg Config) *mux.Router {
    // ... existing code

    if cfg.OrchestratorEnabled {
        nc, err := nats.Connect(cfg.OrchestratorNATSURL)
        if err != nil {
            log.Fatal("Failed to connect to orchestrator:", err)
        }

        // Initialize managers
        syncMgr := orchestrator.NewSyncManager(nc, "v9.0.0", "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm")
        auditMgr := orchestrator.NewAuditManager(nc, "v9.0.0", "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm")
        matrixMgr := orchestrator.NewMatrixManager(nc, "v9.0.0", "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm")

        // Start background sync/audit
        go runPeriodicSync(syncMgr)
        go runPeriodicAudit(auditMgr)
    }

    return mux
}
```

### Passo 4: Testes de Integração

```bash
# Criar testes
touch pkg/orchestrator/sync_test.go
touch pkg/orchestrator/audit_test.go
touch pkg/orchestrator/matrix_test.go
```

### Passo 5: Validação Final

```bash
# Compilar
go build ./cmd

# Testar
go test ./pkg/orchestrator/...

# Validar
cd E:\vertikon\.ecosistema-vertikon\mcp-tester-system
go run enhanced_validator_v4.go E:\vertikon\business\SaaS\templates\sdk-ultra-wasm
```

---

## 📊 8. Diagrama de Sequência (Sync Flow)

```
┌────────┐         ┌──────────────┐         ┌─────────────┐
│  SDK   │         │ Orquestrador │         │ NATS Server │
└───┬────┘         └──────┬───────┘         └──────┬──────┘
    │                     │                        │
    │ 1. RequestSync()    │                        │
    ├────────────────────>│                        │
    │                     │                        │
    │                     │ 2. Check Version Matrix│
    │                     ├───────────────────────>│
    │                     │                        │
    │                     │<───────────────────────┤
    │                     │ 3. Matrix Response     │
    │                     │                        │
    │<────────────────────┤                        │
    │ 4. Sync Response    │                        │
    │                     │                        │
    │ 5. Subscribe Seed Updates                    │
    ├─────────────────────────────────────────────>│
    │                     │                        │
    │                     │ 6. Seed Update Available
    │                     ├───────────────────────>│
    │                     │                        │
    │<──────────────────────────────────────────────┤
    │ 7. Seed Update Notification                  │
    │                     │                        │
    │ 8. Download & Apply │                        │
    ├────────────────────>│                        │
    │                     │                        │
    │<────────────────────┤                        │
    │ 9. Update Confirmed │                        │
    │                     │                        │
```

---

## 📞 Próximos Passos

1. **Aguardar release do MCP-ULTRA-WASM-ORQUESTRADOR**
2. **Quando pronto:**
   - Implementar os 3 arquivos Go (`sync.go`, `audit.go`, `matrix.go`)
   - Criar `sdk-metadata.json`
   - Criar `config/orchestrator.yaml`
   - Adicionar testes de integração
   - Atualizar documentação
   - Executar validação final

3. **Testar integração:**
   - Sync automático de seeds
   - Auditoria de versões
   - Validação via matriz

---

## ✅ Resumo

**Status Atual:**
- ✅ Especificação completa
- ✅ Contratos NATS definidos
- ✅ Esquemas de mensagens documentados
- ✅ Código de exemplo preparado
- ✅ Configurações exemplificadas
- ✅ Documentação completa

**Próxima Ação:**
- ⏳ Aguardar MCP-ULTRA-WASM-ORQUESTRADOR v1.0.0
- ⏳ Implementar integrações quando disponível

---

**Versão do Documento:** 1.0.0
**Última Atualização:** 2025-10-05
**Autor:** Claude Sonnet 4.5 (Autonomous Mode)
**Status:** 📋 ESPECIFICAÇÃO COMPLETA
