# 🚀 Deploy - {{PROJECT_NAME}}

Guia completo de deploy e CI/CD do projeto **{{PROJECT_NAME}}**.

---

## 🎯 Pipeline CI/CD Overview

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Commit    │ -> │   Build     │ -> │    Test     │ -> │   Deploy    │
│   to Git    │    │  & Package  │    │  & Quality  │    │ to K8s      │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### 🔄 Stages Pipeline
1. **Checkout** - Clone repository
2. **Build** - Compile application
3. **Test** - Unit, integration, security tests
4. **Quality** - Code coverage, lint, SAST
5. **Package** - Build Docker image
6. **Deploy** - Deploy to Kubernetes

---

## 📋 GitHub Actions Workflow

### .github/workflows/ci-cd.yml
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: {{GITHUB_ORG}}/{{PROJECT_NAME}}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup {{LANGUAGE}}
        uses: {{SETUP_ACTION}}
        with:
          {{language}}-version: '{{VERSION}}'

      - name: Install Dependencies
        run: {{INSTALL_COMMAND}}

      - name: Run Linter
        run: {{LINT_COMMAND}}

      - name: Run Tests
        run: {{TEST_COMMAND}}

      - name: Generate Coverage
        run: {{COVERAGE_COMMAND}}

      - name: Upload Coverage
        uses: codecov/codecov-action@v3

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Security Scan
        uses: securecodewarrior/github-action-add-sarif@v1
        with:
          sarif-file: security-scan.sarif

      - name: Dependency Check
        run: {{DEPENDENCY_SCAN_COMMAND}}

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
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=sha,prefix={{date 'YYYYMMDD'}}-

      - name: Build and Push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    needs: [build]
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - name: Deploy to Staging
        uses: {{DEPLOY_ACTION}}
        with:
          kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}
          namespace: {{PROJECT_NAME}}-staging
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:develop

  deploy-production:
    if: github.ref == 'refs/heads/main'
    needs: [build]
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Deploy to Production
        uses: {{DEPLOY_ACTION}}
        with:
          kubeconfig: ${{ secrets.KUBE_CONFIG_PROD }}
          namespace: {{PROJECT_NAME}}-production
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main
```

---

## 🐳 Containerização

### Dockerfile
```dockerfile
# Multi-stage build
FROM {{BASE_IMAGE}}:{{BASE_VERSION}} AS builder

WORKDIR /app
COPY . .

# Install dependencies and build
RUN {{BUILD_COMMANDS}}

# Production image
FROM {{RUNTIME_IMAGE}}:{{RUNTIME_VERSION}}

# Security: non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/{{BINARY_PATH}} /app/{{BINARY_NAME}}
COPY --from=builder /app/configs /app/configs

# Ownership
RUN chown -R appuser:appgroup /app
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:{{PORT}}/health/live || exit 1

EXPOSE {{PORT}} 9090

CMD ["/app/{{BINARY_NAME}}"]
```

### .dockerignore
```
.git
.github
.env*
*.md
tests/
docs/
coverage.html
*.log
node_modules/
.DS_Store
Thumbs.db
```

---

## ☸️ Kubernetes Deployment

### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{PROJECT_NAME}}
  namespace: {{NAMESPACE}}
  labels:
    app: {{PROJECT_NAME}}
    version: {{VERSION}}
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
        version: {{VERSION}}
    spec:
      containers:
      - name: {{PROJECT_NAME}}
        image: {{REGISTRY}}/{{PROJECT_NAME}}:{{IMAGE_TAG}}
        ports:
        - containerPort: {{PORT}}
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: APP_ENV
          value: "production"
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
        readinessProbe:
          httpGet:
            path: /health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
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
          runAsNonRoot: true
          runAsUser: 1001
```

### service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{PROJECT_NAME}}-service
  namespace: {{NAMESPACE}}
  labels:
    app: {{PROJECT_NAME}}
spec:
  selector:
    app: {{PROJECT_NAME}}
  ports:
  - name: http
    port: 80
    targetPort: {{PORT}}
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

### ingress.yaml
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{PROJECT_NAME}}-ingress
  namespace: {{NAMESPACE}}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - {{DOMAIN}}
    secretName: {{PROJECT_NAME}}-tls
  rules:
  - host: {{DOMAIN}}
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

## 🔧 Auto Scaling

### hpa.yaml (Horizontal Pod Autoscaler)
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{PROJECT_NAME}}-hpa
  namespace: {{NAMESPACE}}
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

---

## 🏗️ Infrastructure as Code (Terraform)

### main.tf
```hcl
terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
  }
}

# Kubernetes cluster
resource "kubernetes_namespace" "{{PROJECT_NAME}}" {
  metadata {
    name = "{{PROJECT_NAME}}-${var.environment}"
  }
}

# Database
resource "kubernetes_deployment" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.{{PROJECT_NAME}}.metadata[0].name
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "postgres"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgres"
        }
      }

      spec {
        container {
          name  = "postgres"
          image = "postgres:15-alpine"

          env {
            name  = "POSTGRES_DB"
            value = var.db_name
          }

          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = "{{PROJECT_NAME}}-secrets"
                key  = "DB_PASSWORD"
              }
            }
          }
        }
      }
    }
  }
}

# Variables
variable "environment" {
  description = "Environment (staging/production)"
  type        = string
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "{{DB_NAME}}"
}
```

---

## 🎯 Ambientes de Deploy

### 🟡 Staging Environment
- **Branch**: `develop`
- **URL**: https://staging.{{DOMAIN}}
- **Auto-deploy**: Sim
- **Recursos**: 2 vCPU, 4GB RAM
- **Replicas**: 2

### 🟢 Production Environment
- **Branch**: `main`
- **URL**: https://{{DOMAIN}}
- **Auto-deploy**: Com aprovação manual
- **Recursos**: 4 vCPU, 8GB RAM
- **Replicas**: 3-20 (auto-scaling)

---

## 📊 Deploy Monitoring

### Deployment Metrics
```prometheus
# Deploy success rate
deployment_success_rate{environment="production"} 0.98

# Deployment duration
deployment_duration_seconds{environment="production"} 180

# Rollback frequency
deployment_rollback_total{environment="production"} 2
```

### Health Checks pós Deploy
```bash
# Automated post-deploy checks
./scripts/health-check.sh production

# Performance baseline
./scripts/performance-test.sh --env=production --baseline

# Security scan
./scripts/security-scan.sh --environment=production
```

---

## 🔄 Rollback Procedures

### Automatic Rollback Triggers
- **Health check failure** por >5 minutos
- **Error rate > 5%** por >2 minutos
- **P95 latency > 1s** por >3 minutos

### Manual Rollback
```bash
# K8s rollback
kubectl rollout undo deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# Verificar status
kubectl rollout status deployment/{{PROJECT_NAME}} -n {{NAMESPACE}}

# Specific revision rollback
kubectl rollout undo deployment/{{PROJECT_NAME}} --to-revision=2
```

---

## ✅ Deploy Checklist

### Pré-Deploy
- [ ] Testes passando 100%
- [ ] Code review aprovado
- [ ] Security scan limpo
- [ ] Performance tests OK
- [ ] Database migrations testadas

### Deploy
- [ ] Backup database realizado
- [ ] Deploy executado com sucesso
- [ ] Health checks passando
- [ ] Métricas normais
- [ ] Smoke tests executados

### Pós-Deploy
- [ ] Monitoramento ativo
- [ ] Logs sendo coletados
- [ ] Performance dentro do SLA
- [ ] Alertas configurados
- [ ] Documentação atualizada