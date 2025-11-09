# 🚀 Deployment Guide - {{PROJECT_NAME}}

Guia completo de deployment do projeto **{{PROJECT_NAME}}** em diferentes ambientes.

---

## 🎯 Estratégia de Deployment

### Ambientes
```
Development -> Staging -> Production
     ↓            ↓          ↓
   Local      Integration  Live Users
```

| Ambiente | Branch | Auto-Deploy | Approval | URL |
|----------|--------|-------------|----------|-----|
| **Development** | `develop` | ✅ | Não | https://dev.{{DOMAIN}} |
| **Staging** | `develop` | ✅ | Não | https://staging.{{DOMAIN}} |
| **Production** | `main` | ✅ | Manual | https://{{DOMAIN}} |

---

## 🏁 Quick Start Deployment

### Pré-requisitos
- [x] **Docker** 20.10+
- [x] **Kubernetes** cluster 1.28+
- [x] **kubectl** configurado
- [x] **Helm** 3.0+ (opcional)
- [x] **{{CLOUD_PROVIDER}}** account e CLI

### Deploy Rápido (5 minutos)
```bash
# 1. Clone do repositório
git clone https://github.com/{{ORG}}/{{PROJECT_NAME}}.git
cd {{PROJECT_NAME}}

# 2. Configurar ambiente
cp .env.example .env
# Editar variáveis necessárias

# 3. Build e deploy
make docker-build
make k8s-deploy-staging

# 4. Verificar saúde
kubectl get pods -n {{PROJECT_NAME}}-staging
curl https://staging.{{DOMAIN}}/health
```

---

## 🐳 Containerização

### Dockerfile
```dockerfile
# Multi-stage build para otimização
FROM {{BASE_IMAGE}}:{{VERSION}} AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# Production image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# Change ownership to non-root user
RUN chown -R appuser:appuser /root/
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:{{PORT}}/health/live || exit 1

EXPOSE {{PORT}} 9090

CMD ["./main"]
```

### Docker Build
```bash
# Build da imagem
docker build -t {{PROJECT_NAME}}:latest .

# Build com cache optimization
docker build \
  --cache-from {{REGISTRY}}/{{PROJECT_NAME}}:latest \
  -t {{PROJECT_NAME}}:$(git rev-parse --short HEAD) \
  -t {{PROJECT_NAME}}:latest .

# Push para registry
docker tag {{PROJECT_NAME}}:latest {{REGISTRY}}/{{PROJECT_NAME}}:latest
docker push {{REGISTRY}}/{{PROJECT_NAME}}:latest
```

---

## ☸️ Kubernetes Deployment

### Namespace e Resources
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: {{PROJECT_NAME}}-{{ENV}}
  labels:
    environment: {{ENV}}
    project: {{PROJECT_NAME}}
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: {{PROJECT_NAME}}-quota
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
    persistentvolumeclaims: "4"
```

### ConfigMap e Secrets
```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{PROJECT_NAME}}-config
  namespace: {{PROJECT_NAME}}-{{ENV}}
data:
  APP_ENV: "{{ENV}}"
  APP_PORT: "{{PORT}}"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "{{DB_NAME}}"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  METRICS_ENABLED: "true"
  METRICS_PORT: "9090"
  LOG_LEVEL: "info"

---
# secrets.yaml (aplicar via CI/CD com valores reais)
apiVersion: v1
kind: Secret
metadata:
  name: {{PROJECT_NAME}}-secrets
  namespace: {{PROJECT_NAME}}-{{ENV}}
type: Opaque
data:
  DB_PASSWORD: # base64 encoded
  JWT_SECRET: # base64 encoded
  ENCRYPTION_KEY: # base64 encoded
```

### Application Deployment
```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{PROJECT_NAME}}
  namespace: {{PROJECT_NAME}}-{{ENV}}
  labels:
    app: {{PROJECT_NAME}}
    version: "{{VERSION}}"
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  selector:
    matchLabels:
      app: {{PROJECT_NAME}}
  template:
    metadata:
      labels:
        app: {{PROJECT_NAME}}
        version: "{{VERSION}}"
    spec:
      serviceAccountName: {{PROJECT_NAME}}-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: {{PROJECT_NAME}}
        image: {{REGISTRY}}/{{PROJECT_NAME}}:{{IMAGE_TAG}}
        ports:
        - containerPort: {{PORT}}
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: APP_NAME
          value: "{{PROJECT_NAME}}"
        envFrom:
        - configMapRef:
            name: {{PROJECT_NAME}}-config
        - secretRef:
            name: {{PROJECT_NAME}}-secrets
        livenessProbe:
          httpGet:
            path: /health/live
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 5
          failureThreshold: 3
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
```

### Services e Networking
```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: {{PROJECT_NAME}}-service
  namespace: {{PROJECT_NAME}}-{{ENV}}
  labels:
    app: {{PROJECT_NAME}}
spec:
  selector:
    app: {{PROJECT_NAME}}
  ports:
  - name: http
    port: 80
    targetPort: {{PORT}}
    protocol: TCP
  - name: metrics
    port: 9090
    targetPort: 9090
    protocol: TCP
  type: ClusterIP

---
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{PROJECT_NAME}}-ingress
  namespace: {{PROJECT_NAME}}-{{ENV}}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
    nginx.ingress.kubernetes.io/rate-limit-window: "1m"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - {{DOMAIN_FOR_ENV}}
    secretName: {{PROJECT_NAME}}-tls
  rules:
  - host: {{DOMAIN_FOR_ENV}}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{PROJECT_NAME}}-service
            port:
              number: 80
```

---

## 🗄️ Database Deployment

### PostgreSQL StatefulSet
```yaml
# postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          value: {{DB_NAME}}
        - name: POSTGRES_USER
          value: {{DB_USER}}
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: {{PROJECT_NAME}}-secrets
              key: DB_PASSWORD
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### Redis Cache
```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"

---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
```

---

## 📈 Auto Scaling

### Horizontal Pod Autoscaler
```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{PROJECT_NAME}}-hpa
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{PROJECT_NAME}}
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 30
```

### Vertical Pod Autoscaler (Opcional)
```yaml
# vpa.yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: {{PROJECT_NAME}}-vpa
  namespace: {{PROJECT_NAME}}-{{ENV}}
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{PROJECT_NAME}}
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: {{PROJECT_NAME}}
      minAllowed:
        cpu: 100m
        memory: 128Mi
      maxAllowed:
        cpu: 1000m
        memory: 1Gi
```

---

## 🔄 CI/CD Pipeline

### GitHub Actions
```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: {{ORG}}/{{PROJECT_NAME}}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '{{GO_VERSION}}'

      - name: Run Tests
        run: |
          go mod tidy
          go test ./... -race -coverprofile=coverage.out
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload Coverage
        uses: codecov/codecov-action@v3

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run Gosec Security Scanner
        uses: securecodewarrior/github-action-gosec@master

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Login to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}

  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    needs: [build]
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4

      - name: Setup kubectl
        uses: azure/setup-kubectl@v3

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: {{AWS_REGION}}

      - name: Deploy to Staging
        run: |
          aws eks update-kubeconfig --name {{EKS_CLUSTER_NAME}}
          kubectl apply -f k8s/staging/
          kubectl set image deployment/{{PROJECT_NAME}} {{PROJECT_NAME}}=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} -n {{PROJECT_NAME}}-staging
          kubectl rollout status deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-staging

  deploy-production:
    if: github.ref == 'refs/heads/main'
    needs: [build]
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Production
        run: |
          aws eks update-kubeconfig --name {{EKS_CLUSTER_PROD}}
          kubectl apply -f k8s/production/
          kubectl set image deployment/{{PROJECT_NAME}} {{PROJECT_NAME}}=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} -n {{PROJECT_NAME}}-production
          kubectl rollout status deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-production
```

---

## 📊 Monitoring Setup

### ServiceMonitor (Prometheus)
```yaml
# servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{PROJECT_NAME}}-monitor
  namespace: {{PROJECT_NAME}}-{{ENV}}
  labels:
    app: {{PROJECT_NAME}}
spec:
  selector:
    matchLabels:
      app: {{PROJECT_NAME}}
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

### Grafana Dashboard ConfigMap
```yaml
# dashboard-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{PROJECT_NAME}}-dashboard
  namespace: monitoring
  labels:
    grafana_dashboard: "1"
data:
  {{PROJECT_NAME}}.json: |
    {
      "dashboard": {
        "title": "{{PROJECT_NAME}} - Overview",
        "panels": [
          {
            "title": "Requests per Second",
            "targets": [
              {
                "expr": "rate(http_requests_total{job=\"{{PROJECT_NAME}}\"}[5m])"
              }
            ]
          }
        ]
      }
    }
```

---

## 🔄 Rollback Procedures

### Automatic Rollback
```bash
# Rollback para versão anterior
kubectl rollout undo deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}}

# Rollback para revisão específica
kubectl rollout undo deployment/{{PROJECT_NAME}} --to-revision=2 -n {{PROJECT_NAME}}-{{ENV}}

# Verificar status do rollback
kubectl rollout status deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}}

# Ver histórico de deploys
kubectl rollout history deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}}
```

### Database Rollback
```bash
# Restore de backup específico
kubectl exec -i deployment/postgres -n {{PROJECT_NAME}}-{{ENV}} -- \
  psql -U {{DB_USER}} -d {{DB_NAME}} < backup-20240115-103000.sql

# Aplicar migration reversa se disponível
kubectl exec deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}} -- \
  ./migrate -database "postgres://..." -source "file://migrations" down 1
```

---

## ✅ Deployment Checklist

### Pré-Deploy
- [ ] **Testes** passando 100%
- [ ] **Security scan** limpo
- [ ] **Code review** aprovado
- [ ] **Database migrations** testadas
- [ ] **Backup** realizado

### Deploy
- [ ] **Health checks** configurados
- [ ] **Resource limits** definidos
- [ ] **Auto-scaling** configurado
- [ ] **Monitoring** ativo
- [ ] **Logs** sendo coletados

### Pós-Deploy
- [ ] **Smoke tests** executados
- [ ] **Métricas** normais
- [ ] **Performance** dentro do SLA
- [ ] **Alertas** funcionando
- [ ] **Team** notificado

---

## 🆘 Troubleshooting

### Problemas Comuns

#### Pods não inicializam
```bash
# Verificar status dos pods
kubectl get pods -n {{PROJECT_NAME}}-{{ENV}}

# Ver logs detalhados
kubectl describe pod {{POD_NAME}} -n {{PROJECT_NAME}}-{{ENV}}
kubectl logs {{POD_NAME}} -n {{PROJECT_NAME}}-{{ENV}} --previous

# Verificar resources
kubectl top pods -n {{PROJECT_NAME}}-{{ENV}}
```

#### Database connection issues
```bash
# Verificar service DNS
kubectl exec -it deployment/{{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}} -- nslookup postgres-service

# Testar conexão
kubectl exec -it deployment/postgres -n {{PROJECT_NAME}}-{{ENV}} -- psql -U {{DB_USER}} -d {{DB_NAME}} -c "SELECT 1;"

# Verificar secrets
kubectl get secrets -n {{PROJECT_NAME}}-{{ENV}}
kubectl describe secret {{PROJECT_NAME}}-secrets -n {{PROJECT_NAME}}-{{ENV}}
```

#### Performance issues
```bash
# Verificar métricas de resource usage
kubectl top nodes
kubectl top pods -n {{PROJECT_NAME}}-{{ENV}}

# HPA status
kubectl get hpa -n {{PROJECT_NAME}}-{{ENV}}

# Verificar limites e requests
kubectl describe deployment {{PROJECT_NAME}} -n {{PROJECT_NAME}}-{{ENV}}
```