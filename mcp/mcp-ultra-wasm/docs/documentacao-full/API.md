# 🌐 API REST - {{PROJECT_NAME}}

Documentação completa da API REST do projeto **{{PROJECT_NAME}}**.

---

## 📌 Base URL
```
Production: https://{{DOMAIN}}/api/v1
Staging: https://staging.{{DOMAIN}}/api/v1
Development: http://localhost:{{PORT}}/api/v1
```

---

## 🔐 Autenticação
Todas as rotas protegidas requerem token JWT no header:
```bash
Authorization: Bearer <token>
```

---

## 🏥 Health Endpoints

### GET /health
Status geral da aplicação
```json
{
  "status": "healthy",
  "service": "{{PROJECT_NAME}}",
  "version": "{{VERSION}}",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /health/ready
Verifica dependências (DB, Redis, etc.)
```json
{
  "status": "ready",
  "dependencies": {
    "database": "connected",
    "redis": "connected"
  }
}
```

---

## 📋 Endpoints Principais

### {{ENTITY_1}} Endpoints
- `GET /{{entity1}}` - Listar {{entity1}}s
- `POST /{{entity1}}` - Criar {{entity1}}
- `GET /{{entity1}}/{id}` - Obter {{entity1}} por ID
- `PUT /{{entity1}}/{id}` - Atualizar {{entity1}}
- `DELETE /{{entity1}}/{id}` - Excluir {{entity1}}

### {{ENTITY_2}} Endpoints
- `GET /{{entity2}}` - Listar {{entity2}}s
- `POST /{{entity2}}` - Criar {{entity2}}
- `GET /{{entity2}}/{id}` - Obter {{entity2}} por ID
- `PUT /{{entity2}}/{id}` - Atualizar {{entity2}}
- `DELETE /{{entity2}}/{id}` - Excluir {{entity2}}

---

## 📊 Métricas & Relatórios
- `GET /metrics` - Métricas Prometheus
- `GET /reports` - Relatórios de negócio

---

## 🚫 Códigos de Erro
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `422` - Validation Error
- `500` - Internal Server Error

---

## 📝 Exemplos de Uso

### Criar {{ENTITY_1}}
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Exemplo",
    "description": "Descrição do exemplo"
  }' \
  https://{{DOMAIN}}/api/v1/{{entity1}}
```

### Listar {{ENTITY_1}}s
```bash
curl -X GET \
  -H "Authorization: Bearer <token>" \
  https://{{DOMAIN}}/api/v1/{{entity1}}?page=1&limit=10
```