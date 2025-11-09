# 📊 Observabilidade - {{PROJECT_NAME}}

Stack completa de monitoramento, métricas e observabilidade do projeto **{{PROJECT_NAME}}**.

---

## 🎯 Stack de Observabilidade

### 📈 Métricas - Prometheus
- **Coleta**: métricas de aplicação e sistema
- **Storage**: TSDB (Time Series Database)
- **Alerting**: regras configuradas no Prometheus

### 📊 Visualização - Grafana
- **Dashboards**: métricas de negócio e técnicas
- **Alertas**: integração com Slack/Email
- **Drill-down**: análise detalhada

### 🔍 Tracing - Jaeger
- **Distributed tracing**: requisições end-to-end
- **Performance**: latência por componente
- **Debug**: troubleshooting de problemas

### 📝 Logs - ELK Stack
- **Elasticsearch**: armazenamento e busca
- **Logstash**: processamento de logs
- **Kibana**: visualização e análise

---

## 📊 Métricas Implementadas

### 🌐 HTTP Metrics
```prometheus
# Total de requests
http_requests_total{method="GET",endpoint="/api/v1/users",status="200"} 1250

# Latência de requests
http_request_duration_seconds_bucket{method="POST",endpoint="/api/v1/login",le="0.1"} 100

# Requests em andamento
http_requests_in_flight{endpoint="/api/v1/reports"} 5
```

### 💾 Database Metrics
```prometheus
# Conexões ativas
database_connections_active 45
database_connections_max 100

# Query performance
database_query_duration_seconds{query="select_users"} 0.025

# Connection pool
database_connection_pool_size 20
database_connection_pool_used 8
```

### 📈 Business Metrics
```prometheus
# Métricas de negócio específicas do projeto
{{business_metric_1}}_total{status="completed"} 1500
{{business_metric_2}}_duration_seconds{type="premium"} 45.2
{{business_metric_3}}_errors_total{reason="validation"} 12
```

### ⚡ Application Metrics
```prometheus
# Memory & CPU
process_resident_memory_bytes 104857600
process_cpu_seconds_total 3600

# Garbage Collection (Go)
go_gc_duration_seconds 0.001234
go_goroutines 150

# Custom metrics
{{PROJECT_NAME}}_active_users 245
{{PROJECT_NAME}}_cache_hits_total 89500
{{PROJECT_NAME}}_cache_misses_total 1200
```

---

## 🚨 Alertas Configurados

### 🔴 Críticos
```yaml
# Aplicação DOWN
- alert: ApplicationDown
  expr: up{job="{{PROJECT_NAME}}"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "{{PROJECT_NAME}} está DOWN"

# Alta latência
- alert: HighLatency
  expr: http_request_duration_seconds{quantile="0.95"} > 0.5
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Latência alta detectada (>500ms P95)"
```

### 🟡 Warnings
```yaml
# Uso alto de CPU
- alert: HighCPUUsage
  expr: process_cpu_seconds_total > 0.8
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "CPU usage alto (>80%)"

# Database connections
- alert: HighDatabaseConnections
  expr: database_connections_active / database_connections_max > 0.8
  for: 3m
  labels:
    severity: warning
  annotations:
    summary: "Conexões DB altas (>80%)"
```

---

## 📊 Dashboards Grafana

### 🎯 Dashboard Principal - Overview
```json
{
  "dashboard": {
    "title": "{{PROJECT_NAME}} - Overview",
    "panels": [
      {
        "title": "Requests/sec",
        "query": "rate(http_requests_total[5m])"
      },
      {
        "title": "P95 Latency",
        "query": "histogram_quantile(0.95, http_request_duration_seconds_bucket)"
      },
      {
        "title": "Error Rate",
        "query": "rate(http_requests_total{status=~\"4..|5..\"}[5m])"
      }
    ]
  }
}
```

### 💼 Dashboard de Negócio
- **{{Business_Metric_1}}** por período
- **{{Business_Metric_2}}** por categoria
- **Receita** e **conversões**
- **Usuários ativos** em tempo real

### 🔧 Dashboard Técnico
- **CPU, Memory, Disk** usage
- **Database** performance
- **Cache** hit ratio
- **Network** I/O

### 🚨 Dashboard de Alertas
- **Alertas ativos** por severidade
- **Histórico** de incidentes
- **MTTR** (Mean Time To Recovery)
- **SLA** status

---

## 🔍 Distributed Tracing

### Jaeger Implementation
```{{LANGUAGE_LOWER}}
// Inicialização do tracing
tracer := jaeger.NewTracer("{{PROJECT_NAME}}")

// Trace de request HTTP
span := tracer.StartSpan("http_request")
defer span.Finish()

// Child spans para operações internas
dbSpan := tracer.StartSpan("database_query", opentracing.ChildOf(span.Context()))
// ... database operation
dbSpan.Finish()
```

### Trace Examples
```
Request ID: abc123xyz789
├── HTTP Handler (45ms)
│   ├── Authentication (5ms)
│   ├── Authorization (3ms)
│   ├── Database Query (25ms)
│   │   ├── Connection Pool (2ms)
│   │   └── SQL Execution (23ms)
│   ├── Cache Store (8ms)
│   └── Response Serialization (4ms)
```

---

## 📝 Structured Logging

### Log Format (JSON)
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "service": "{{PROJECT_NAME}}",
  "version": "{{VERSION}}",
  "trace_id": "abc123xyz789",
  "span_id": "def456uvw012",
  "user_id": "user_123",
  "request_id": "req_789xyz",
  "method": "POST",
  "path": "/api/v1/{{entity}}",
  "status": 201,
  "duration_ms": 45,
  "message": "{{Entity}} created successfully"
}
```

### Log Levels
- **DEBUG**: Informações detalhadas para desenvolvimento
- **INFO**: Informações gerais de operação
- **WARN**: Situações que podem precisar atenção
- **ERROR**: Erros que não afetam a operação geral
- **FATAL**: Erros críticos que param a aplicação

### Structured Fields
```{{LANGUAGE_LOWER}}
log.WithFields(logrus.Fields{
    "user_id": userID,
    "action": "create_{{entity}}",
    "{{entity}}_id": {{entity}}ID,
    "duration_ms": duration,
}).Info("{{Entity}} created successfully")
```

---

## 🎛️ Health Checks

### Endpoints de Saúde
```http
# Liveness probe
GET /health/live
Response: {"status": "alive", "timestamp": "2024-01-15T10:30:00Z"}

# Readiness probe
GET /health/ready
Response: {
  "status": "ready",
  "dependencies": {
    "database": "connected",
    "redis": "connected",
    "external_api": "connected"
  },
  "checks": {
    "database": {"status": "pass", "response_time": "25ms"},
    "redis": {"status": "pass", "response_time": "2ms"}
  }
}
```

### Kubernetes Probes
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: {{PORT}}
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: {{PORT}}
  initialDelaySeconds: 5
  periodSeconds: 5
```

---

## 📈 SLI/SLO Configuration

### Service Level Indicators (SLI)
- **Availability**: uptime do serviço
- **Latency**: tempo de resposta P95 < 200ms
- **Throughput**: requests processadas por segundo
- **Error Rate**: % de erros < 1%

### Service Level Objectives (SLO)
- **99.9%** availability (8.77h downtime/ano)
- **P95 latency < 200ms** para 95% dos requests
- **P99 latency < 500ms** para 99% dos requests
- **Error rate < 0.1%** dos requests totais

### Error Budget
- **Monthly error budget**: 43.2 minutos
- **Burn rate alerts**: quando consumo > 10x normal
- **Policy**: stop releases se error budget < 5%

---

## 🔧 Configuração e Setup

### Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: '{{PROJECT_NAME}}'
    static_configs:
      - targets: ['app:9090']
    scrape_interval: 5s
    metrics_path: /metrics

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

### Grafana Datasources
```yaml
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    access: proxy

  - name: Jaeger
    type: jaeger
    url: http://jaeger:16686
    access: proxy

  - name: Elasticsearch
    type: elasticsearch
    url: http://elasticsearch:9200
    access: proxy
```

---

## 📊 Métricas de Resultado

### Performance Atual
- **Availability**: 99.95%
- **P95 Latency**: 125ms
- **P99 Latency**: 250ms
- **Error Rate**: 0.05%
- **MTTR**: 4.2 minutes

### Observability Coverage
- ✅ **100%** endpoints monitorados
- ✅ **100%** critical paths traced
- ✅ **95%** code coverage em logs
- ✅ **24/7** alerting ativo