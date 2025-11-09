# ⚙️ Runbook Operacional - {{PROJECT_NAME}}

Manual operacional completo para gerenciamento do projeto **{{PROJECT_NAME}}** em produção.

---

## 🚨 Contatos de Emergência

### Escalation Matrix
| Severidade | Contato Primário | Backup | SLA Response |
|------------|------------------|--------|--------------|
| **SEV1** | On-call Engineer | Tech Lead | 15 min |
| **SEV2** | DevOps Team | Product Owner | 1 hour |
| **SEV3** | Support Team | Dev Team | 4 hours |
| **SEV4** | Support Team | - | 24 hours |

### Contatos
- **On-call**: +55 (11) 99999-9999
- **DevOps**: devops@{{DOMAIN}}
- **Support**: support@{{DOMAIN}}
- **Security**: security@{{DOMAIN}}

---

## 📊 Monitoramento e Alertas

### Dashboards Principais
- **Overview**: https://grafana.{{DOMAIN}}/d/overview
- **Infrastructure**: https://grafana.{{DOMAIN}}/d/infra
- **Application**: https://grafana.{{DOMAIN}}/d/app
- **Business Metrics**: https://grafana.{{DOMAIN}}/d/business

### Alertas Críticos (SEV1)

#### 🔴 Application Down
```bash
# Verificar status dos pods
kubectl get pods -n {{NAMESPACE}}

# Logs da aplicação
kubectl logs -f deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# Restart se necessário
kubectl rollout restart deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}
```

#### 🔴 Database Connection Lost
```bash
# Verificar conexões DB
kubectl exec -it deployment/postgres -n {{NAMESPACE}} -- psql -U {{DB_USER}} -d {{DB_NAME}} -c "SELECT count(*) FROM pg_stat_activity;"

# Restart database pod se necessário
kubectl delete pod postgres-xxx -n {{NAMESPACE}}
```

#### 🔴 High Error Rate (>5%)
```bash
# Verificar logs de erro
kubectl logs deployment/{{PROJECT_NAME}} -n {{NAMESPACE}} --tail=100 | grep ERROR

# Verificar métricas de erro
curl https://{{DOMAIN}}/metrics | grep http_requests_total | grep "5.."
```

---

## 🔄 Procedimentos Operacionais

### Deploy de Emergência
```bash
# 1. Fazer backup do deployment atual
kubectl get deployment {{PROJECT_NAME}} -n {{NAMESPACE}} -o yaml > backup-deployment.yaml

# 2. Deploy da versão de emergência
kubectl set image deployment/{{PROJECT_NAME}} {{PROJECT_NAME}}={{EMERGENCY_IMAGE}} -n {{NAMESPACE}}

# 3. Verificar rollout
kubectl rollout status deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# 4. Verificar saúde
curl https://{{DOMAIN}}/health
```

### Rollback de Produção
```bash
# Ver histórico de deploys
kubectl rollout history deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# Rollback para versão anterior
kubectl rollout undo deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# Rollback para revisão específica
kubectl rollout undo deployment/{{PROJECT_NAME}} --to-revision=2 -n {{NAMESPACE}}

# Verificar status
kubectl rollout status deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}
```

### Scaling Manual
```bash
# Scale up para handle de carga
kubectl scale deployment {{PROJECT_NAME}} --replicas=10 -n {{NAMESPACE}}

# Scale down após pico
kubectl scale deployment {{PROJECT_NAME}} --replicas=3 -n {{NAMESPACE}}

# Verificar HPA status
kubectl get hpa {{PROJECT_NAME}}-hpa -n {{NAMESPACE}}
```

---

## 💾 Backup e Recovery

### Database Backup
```bash
# Backup manual imediato
kubectl exec deployment/postgres -n {{NAMESPACE}} -- pg_dump -U {{DB_USER}} {{DB_NAME}} > backup-$(date +%Y%m%d-%H%M%S).sql

# Verificar backups automáticos
kubectl get cronjobs -n {{NAMESPACE}}

# Restore de backup
kubectl exec -i deployment/postgres -n {{NAMESPACE}} -- psql -U {{DB_USER}} -d {{DB_NAME}} < backup-20240115-103000.sql
```

### Application State Backup
```bash
# Export de configurações
kubectl get configmap {{PROJECT_NAME}}-config -n {{NAMESPACE}} -o yaml > config-backup.yaml
kubectl get secret {{PROJECT_NAME}}-secrets -n {{NAMESPACE}} -o yaml > secrets-backup.yaml

# Restore de configurações
kubectl apply -f config-backup.yaml
kubectl apply -f secrets-backup.yaml
```

---

## 🔍 Troubleshooting Guide

### Alta Latência (P95 > 500ms)
```bash
# 1. Verificar CPU/Memory dos pods
kubectl top pods -n {{NAMESPACE}}

# 2. Verificar conexões de database
kubectl exec deployment/postgres -n {{NAMESPACE}} -- psql -U {{DB_USER}} -d {{DB_NAME}} -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# 3. Verificar queries lentas
kubectl exec deployment/postgres -n {{NAMESPACE}} -- psql -U {{DB_USER}} -d {{DB_NAME}} -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# 4. Verificar cache hit ratio
kubectl exec deployment/redis -n {{NAMESPACE}} -- redis-cli info stats | grep keyspace_hits
```

### Memory Leaks
```bash
# 1. Verificar usage por pod
kubectl top pods -n {{NAMESPACE}} --sort-by=memory

# 2. Analisar memory profile da aplicação
kubectl port-forward deployment/{{PROJECT_NAME}} 6060:6060 -n {{NAMESPACE}}
curl http://localhost:6060/debug/pprof/heap > heap.profile

# 3. Restart pods com memory usage alta
kubectl delete pod {{POD_NAME}} -n {{NAMESPACE}}
```

### Disk Space Issues
```bash
# 1. Verificar disk usage dos nodes
kubectl describe nodes | grep -A 5 "Allocated resources"

# 2. Cleanup de logs antigos
kubectl logs deployment/{{PROJECT_NAME}} -n {{NAMESPACE}} --tail=1000 > recent-logs.txt

# 3. Verificar persistent volumes
kubectl get pv
kubectl describe pv {{PV_NAME}}
```

---

## 📈 Performance Tuning

### Database Optimization
```sql
-- Verificar queries mais lentas
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY total_time DESC LIMIT 10;

-- Verificar índices não utilizados
SELECT schemaname, tablename, indexname
FROM pg_stat_user_indexes
WHERE idx_scan = 0;

-- Análise de vacuum
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, n_dead_tup
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000;
```

### Application Tuning
```bash
# Ajustar connection pool
kubectl patch configmap {{PROJECT_NAME}}-config -n {{NAMESPACE}} --patch '{"data":{"DB_MAX_CONNECTIONS":"200"}}'

# Ajustar memory limits
kubectl patch deployment {{PROJECT_NAME}} -n {{NAMESPACE}} --patch '{"spec":{"template":{"spec":{"containers":[{"name":"{{PROJECT_NAME}}","resources":{"limits":{"memory":"1Gi"}}}]}}}}'

# Restart para aplicar mudanças
kubectl rollout restart deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}
```

---

## 🔐 Incidentes de Segurança

### Suspeita de Intrusão
```bash
# 1. Isolar o ambiente suspeito
kubectl scale deployment {{PROJECT_NAME}} --replicas=0 -n {{NAMESPACE}}

# 2. Capturar logs para análise
kubectl logs deployment/{{PROJECT_NAME}} -n {{NAMESPACE}} --previous > incident-logs.txt

# 3. Verificar acessos suspeitos
grep "401\|403\|429" incident-logs.txt

# 4. Notificar security team
curl -X POST https://security.{{DOMAIN}}/incident \
  -H "Content-Type: application/json" \
  -d '{"type": "security_incident", "severity": "high", "description": "Suspicious activity detected"}'
```

### Vazamento de Dados Suspeito
```bash
# 1. Verificar logs de acesso a dados sensíveis
kubectl logs deployment/{{PROJECT_NAME}} -n {{NAMESPACE}} | grep "SENSITIVE_DATA_ACCESS"

# 2. Verificar queries de mass export
kubectl exec deployment/postgres -n {{NAMESPACE}} -- psql -U {{DB_USER}} -d {{DB_NAME}} -c "SELECT query, calls FROM pg_stat_statements WHERE query LIKE '%SELECT%' AND calls > 1000;"

# 3. Implementar rate limiting temporário
kubectl patch configmap {{PROJECT_NAME}}-config -n {{NAMESPACE}} --patch '{"data":{"RATE_LIMIT_REQUESTS":"10"}}'
```

---

## 📊 Maintenance Windows

### Planejamento de Manutenção
1. **Notificar usuários** 48h antes
2. **Backup completo** da aplicação e dados
3. **Deploy em staging** primeiro
4. **Executar testes** pós-deploy
5. **Monitorar métricas** por 2h após

### Checklist de Manutenção
```bash
# Pré-manutenção
- [ ] Backup database realizado
- [ ] Backup configurações realizado
- [ ] Usuários notificados
- [ ] Equipe de plantão disponível

# Durante manutenção
- [ ] Deploy executado com sucesso
- [ ] Health checks passando
- [ ] Métricas dentro do normal
- [ ] Testes smoke executados

# Pós-manutenção
- [ ] Monitoramento ativo por 2h
- [ ] Performance dentro do SLA
- [ ] Usuários notificados da conclusão
- [ ] Documentação atualizada
```

---

## 📞 Runbook de Incidentes

### Severidade de Incidentes

#### SEV1 - Crítico
- **Definição**: Serviço completamente indisponível
- **Exemplos**: Site down, database corrupted
- **Response Time**: 15 minutos
- **Resolution Time**: 4 horas

#### SEV2 - Alto
- **Definição**: Funcionalidade principal afetada
- **Exemplos**: Login não funciona, API lenta
- **Response Time**: 1 hora
- **Resolution Time**: 24 horas

#### SEV3 - Médio
- **Definição**: Funcionalidade secundária afetada
- **Exemplos**: Relatório não gera, cache miss alto
- **Response Time**: 4 horas
- **Resolution Time**: 5 dias

#### SEV4 - Baixo
- **Definição**: Problema cosmético ou melhoria
- **Exemplos**: UI bug, performance optimization
- **Response Time**: 24 horas
- **Resolution Time**: 2 semanas

### Processo de Incident Response
1. **Detecção** - Alerta automático ou report manual
2. **Assessment** - Determinar severidade e impacto
3. **Notification** - Alertar equipes e stakeholders
4. **Investigation** - Root cause analysis
5. **Mitigation** - Implementar workaround temporário
6. **Resolution** - Fix permanente
7. **Post-mortem** - Lições aprendidas e melhorias

---

## ✅ Health Checks Regulares

### Checklist Diário
- [ ] Verificar dashboards de monitoramento
- [ ] Analisar logs de erro
- [ ] Verificar métricas de performance
- [ ] Conferir status dos backups
- [ ] Validar certificados SSL

### Checklist Semanal
- [ ] Review de alertas da semana
- [ ] Análise de trends de performance
- [ ] Verificar disk space usage
- [ ] Review de security logs
- [ ] Update de dependências críticas

### Checklist Mensal
- [ ] Review completo de performance
- [ ] Análise de capacity planning
- [ ] Security vulnerability scan
- [ ] Disaster recovery test
- [ ] Documentation update