// Package services define contratos mínimos usados por handlers e testes.
package services

// HealthStatus e HealthChecker são referenciados por testes do router.
type HealthStatus struct {
	Status string // ex.: "ok", "degraded"
}

type HealthChecker interface {
	Status() HealthStatus
}
