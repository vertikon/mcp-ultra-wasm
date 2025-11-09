package main

import future.keywords.contains
import future.keywords.if
import future.keywords.in

# MCP Ultra Security Policies - Kubernetes & Docker

# =============================================================
# Kubernetes Security Policies
# =============================================================

# Deny pods without runAsNonRoot
deny[msg] {
  input.kind == "Deployment"
  not input.spec.template.spec.securityContext.runAsNonRoot
  msg := sprintf("Deployment '%s' must set runAsNonRoot: true", [input.metadata.name])
}

deny[msg] {
  input.kind == "StatefulSet"
  not input.spec.template.spec.securityContext.runAsNonRoot
  msg := sprintf("StatefulSet '%s' must set runAsNonRoot: true", [input.metadata.name])
}

# Deny containers without readOnlyRootFilesystem
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  not container.securityContext.readOnlyRootFilesystem
  msg := sprintf("Container '%s' must set readOnlyRootFilesystem: true", [container.name])
}

# Deny containers with allowPrivilegeEscalation
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  container.securityContext.allowPrivilegeEscalation
  msg := sprintf("Container '%s' must not allow privilege escalation", [container.name])
}

# Deny containers without resource limits
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  not container.resources.limits.memory
  msg := sprintf("Container '%s' must specify memory limits", [container.name])
}

deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  not container.resources.limits.cpu
  msg := sprintf("Container '%s' must specify CPU limits", [container.name])
}

# Deny containers running as root (UID 0)
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  container.securityContext.runAsUser == 0
  msg := sprintf("Container '%s' must not run as root (UID 0)", [container.name])
}

# Deny pods without NetworkPolicy
deny[msg] {
  input.kind == "Deployment"
  not input.metadata.labels["network-policy"]
  msg := sprintf("Deployment '%s' must have network-policy label", [input.metadata.name])
}

# Deny services without TLS
deny[msg] {
  input.kind == "Service"
  input.spec.type == "LoadBalancer"
  not input.metadata.annotations["service.beta.kubernetes.io/aws-load-balancer-ssl-cert"]
  msg := sprintf("LoadBalancer service '%s' must use TLS/SSL", [input.metadata.name])
}

# Deny Ingress without TLS
deny[msg] {
  input.kind == "Ingress"
  not input.spec.tls
  msg := sprintf("Ingress '%s' must use TLS", [input.metadata.name])
}

# Deny PodDisruptionBudget with maxUnavailable > 1
deny[msg] {
  input.kind == "PodDisruptionBudget"
  input.spec.maxUnavailable > 1
  msg := sprintf("PodDisruptionBudget '%s' maxUnavailable should not exceed 1", [input.metadata.name])
}

# =============================================================
# Docker/Container Security Policies
# =============================================================

# Deny latest tags
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  endswith(container.image, ":latest")
  msg := sprintf("Container '%s' must not use :latest tag", [container.name])
}

deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  not contains(container.image, ":")
  msg := sprintf("Container '%s' must specify image tag", [container.name])
}

# Deny containers from untrusted registries
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "DaemonSet"]
  container := input.spec.template.spec.containers[_]
  not trusted_registry(container.image)
  msg := sprintf("Container '%s' must use trusted registry", [container.name])
}

trusted_registry(image) {
  prefixes := [
    "gcr.io/",
    "docker.io/",
    "quay.io/",
    "registry.hub.docker.com/",
    "ghcr.io/",
    "localhost:5000/"
  ]
  startswith(image, prefixes[_])
}

# =============================================================
# RBAC Security Policies
# =============================================================

# Deny overly permissive RBAC
deny[msg] {
  input.kind == "ClusterRole"
  rule := input.rules[_]
  rule.verbs[_] == "*"
  rule.resources[_] == "*"
  msg := sprintf("ClusterRole '%s' is too permissive with wildcard verbs and resources", [input.metadata.name])
}

deny[msg] {
  input.kind == "Role"
  rule := input.rules[_]
  rule.verbs[_] == "*"
  msg := sprintf("Role '%s' should not use wildcard verbs", [input.metadata.name])
}

# =============================================================
# ConfigMap/Secret Security Policies
# =============================================================

# Warn about sensitive data in ConfigMaps
deny[msg] {
  input.kind == "ConfigMap"
  data_keys := object.keys(input.data)
  sensitive_keys := ["password", "secret", "token", "key", "api_key", "apikey"]
  some key in data_keys
  some sensitive in sensitive_keys
  contains(lower(key), sensitive)
  msg := sprintf("ConfigMap '%s' may contain sensitive data in key '%s', use Secret instead", [input.metadata.name, key])
}

# Deny Secrets without encryption
deny[msg] {
  input.kind == "Secret"
  not input.metadata.annotations["kubernetes.io/encryption"]
  msg := sprintf("Secret '%s' should be encrypted at rest", [input.metadata.name])
}

# =============================================================
# Namespace Security Policies
# =============================================================

# Deny resources in default namespace
deny[msg] {
  input.kind in ["Deployment", "StatefulSet", "Service", "DaemonSet"]
  input.metadata.namespace == "default"
  msg := sprintf("%s '%s' should not be deployed to default namespace", [input.kind, input.metadata.name])
}

# Require namespace labels
deny[msg] {
  input.kind == "Namespace"
  not input.metadata.labels.environment
  msg := sprintf("Namespace '%s' must have environment label", [input.metadata.name])
}

# =============================================================
# HPA/Autoscaling Policies
# =============================================================

# Require PodDisruptionBudget for deployments with HPA
deny[msg] {
  input.kind == "HorizontalPodAutoscaler"
  not input.metadata.annotations["pdb-configured"]
  msg := sprintf("HPA '%s' should have corresponding PodDisruptionBudget", [input.metadata.name])
}