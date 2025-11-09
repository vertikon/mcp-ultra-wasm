// pkg/contracts/job.go
package contracts

import "context"

// Job representa trabalho agendado ou background
// SemVer: v1.0.0 - Interface estável
type Job interface {
	// Name identifica o job
	Name() string

	// Schedule retorna expressão cron (ex: "0 */5 * * *")
	// Retorne "" para jobs manuais
	Schedule() string

	// Run executa o job
	Run(ctx context.Context) error
}
