# ⚙️ Configuração - {{PROJECT_NAME}}

Guia completo de configuração do projeto **{{PROJECT_NAME}}**.

---

## 🔧 Variáveis de Ambiente

### 📋 Arquivo .env
```bash
# ===================================
# APLICAÇÃO
# ===================================
APP_NAME={{PROJECT_NAME}}
APP_VERSION={{VERSION}}
APP_ENV=development  # development | staging | production
APP_PORT={{PORT}}
APP_HOST=0.0.0.0

# ===================================
# BANCO DE DADOS
# ===================================
DB_HOST=localhost
DB_PORT={{DB_PORT}}
DB_NAME={{DB_NAME}}
DB_USER={{DB_USER}}
DB_PASSWORD={{DB_PASSWORD}}
DB_SSL_MODE=disable  # disable | require | verify-full
DB_MAX_CONNECTIONS=100
DB_MAX_IDLE_CONNECTIONS=10

# ===================================
# REDIS (CACHE)
# ===================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_MAX_CONNECTIONS=100

# ===================================
# JWT & SEGURANÇA
# ===================================
JWT_SECRET={{JWT_SECRET}}
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h  # 7 days
ENCRYPTION_KEY={{ENCRYPTION_KEY}}  # 32 bytes para AES-256

# ===================================
# OBSERVABILIDADE
# ===================================
METRICS_ENABLED=true
METRICS_PORT=9090
JAEGER_ENDPOINT=http://localhost:14268/api/traces
LOG_LEVEL=info  # debug | info | warn | error

# ===================================
# RATE LIMITING
# ===================================
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# ===================================
# CORS
# ===================================
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://{{DOMAIN}}
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization

# ===================================
# EXTERNAL APIs
# ===================================
{{EXTERNAL_API_1}}_URL={{API_URL_1}}
{{EXTERNAL_API_1}}_KEY={{API_KEY_1}}
{{EXTERNAL_API_2}}_URL={{API_URL_2}}
{{EXTERNAL_API_2}}_KEY={{API_KEY_2}}
```

---

## 🐳 Docker Compose

### docker-compose.yml
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "{{PORT}}:{{PORT}}"
      - "9090:9090"  # metrics
    environment:
      - APP_ENV=development
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - .:/app
    networks:
      - {{PROJECT_NAME}}-network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: {{DB_NAME}}
      POSTGRES_USER: {{DB_USER}}
      POSTGRES_PASSWORD: {{DB_PASSWORD}}
    ports:
      - "{{DB_PORT}}:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - {{PROJECT_NAME}}-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - {{PROJECT_NAME}}-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - {{PROJECT_NAME}}-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-changeme}
    volumes:
      - grafana_data:/var/lib/grafana
    networks:
      - {{PROJECT_NAME}}-network

volumes:
  postgres_data:
  redis_data:
  grafana_data:

networks:
  {{PROJECT_NAME}}-network:
    driver: bridge
```

---

## ☸️ Kubernetes ConfigMaps

### configmap.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{PROJECT_NAME}}-config
  namespace: {{NAMESPACE}}
data:
  APP_NAME: "{{PROJECT_NAME}}"
  APP_ENV: "production"
  APP_PORT: "{{PORT}}"
  DB_HOST: "{{DB_HOST}}"
  DB_PORT: "{{DB_PORT}}"
  DB_NAME: "{{DB_NAME}}"
  REDIS_HOST: "{{REDIS_HOST}}"
  REDIS_PORT: "6379"
  METRICS_ENABLED: "true"
  METRICS_PORT: "9090"
  LOG_LEVEL: "info"
```

---

## 🔐 Secrets (Kubernetes)

### secrets.yaml
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: {{PROJECT_NAME}}-secrets
  namespace: {{NAMESPACE}}
type: Opaque
data:
  DB_PASSWORD: {{DB_PASSWORD_BASE64}}
  JWT_SECRET: {{JWT_SECRET_BASE64}}
  ENCRYPTION_KEY: {{ENCRYPTION_KEY_BASE64}}
  {{EXTERNAL_API_1}}_KEY: {{API_KEY_1_BASE64}}
```

---

## 📊 Monitoramento - prometheus.yml

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: '{{PROJECT_NAME}}'
    static_configs:
      - targets: ['app:9090']
    metrics_path: /metrics
    scrape_interval: 5s

  - job_name: 'postgres-exporter'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis-exporter'
    static_configs:
      - targets: ['redis-exporter:9121']
```

---

## 🚀 Scripts de Setup

### setup.sh
```bash
#!/bin/bash

# Criar arquivo .env se não existir
if [ ! -f .env ]; then
    cp .env.example .env
    echo "✅ Arquivo .env criado"
fi

# Gerar JWT Secret
JWT_SECRET=$(openssl rand -hex 32)
sed -i "s/JWT_SECRET=.*/JWT_SECRET=$JWT_SECRET/" .env

# Gerar Encryption Key
ENCRYPTION_KEY=$(openssl rand -hex 32)
sed -i "s/ENCRYPTION_KEY=.*/ENCRYPTION_KEY=$ENCRYPTION_KEY/" .env

echo "🔐 Secrets gerados automaticamente"
echo "📝 Edite o arquivo .env com suas configurações específicas"
```

---

## 🔍 Validação de Configuração

### Checklist de Setup
- [ ] Arquivo .env configurado
- [ ] Banco de dados acessível
- [ ] Redis funcionando
- [ ] JWT secrets definidos
- [ ] Logs estruturados habilitados
- [ ] Métricas expostas
- [ ] CORS configurado corretamente
- [ ] Rate limiting ativo

### Comando de Teste
```bash
# Verificar configuração
{{RUN_COMMAND}} --config-check

# Health check completo
curl http://localhost:{{PORT}}/health/ready
```