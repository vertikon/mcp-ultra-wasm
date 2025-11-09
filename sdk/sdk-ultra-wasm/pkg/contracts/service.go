// pkg/contracts/service.go
package contracts

import "context"

// Service representa serviço customizado
// SemVer: v1.0.0 - Interface estável
type Service interface {
	// Name identifica o serviço
	Name() string

	// Start inicializa o serviço
	Start(ctx context.Context) error

	// Stop encerra gracefully
	Stop(ctx context.Context) error

	// Health retorna status de saúde
	Health() error
}
