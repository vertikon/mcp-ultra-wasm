package orchestrator

import "time"

// ═════════════════════════════════════════════════════════════════════════
// ORCHESTRATOR INTEGRATION TYPES
// Status: STUB - Pronto para implementação quando orquestrador estiver pronto
// ═════════════════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────────────────
// SYNC TYPES
// ─────────────────────────────────────────────────────────────────────────

// SyncRequest representa uma solicitação de sincronização
type SyncRequest struct {
	Timestamp  time.Time `json:"timestamp"`
	SDKVersion string    `json:"sdk_version"`
	Module     string    `json:"module"`
	Requester  string    `json:"requester"`
	Seeds      []string  `json:"seeds"`
}

// SyncResponse representa a resposta de sincronização
type SyncResponse struct {
	Timestamp         time.Time        `json:"timestamp"`
	SyncID            string           `json:"sync_id"`
	Status            string           `json:"status"`
	EstimatedDuration string           `json:"estimated_duration"`
	SeedsToUpdate     []SeedUpdate     `json:"seeds_to_update"`
	TemplateUpdates   []TemplateUpdate `json:"template_updates"`
}

// SeedUpdate representa uma atualização de seed
type SeedUpdate struct {
	Timestamp         time.Time `json:"timestamp"`
	SeedName          string    `json:"seed_name"`
	Version           string    `json:"version"`
	DownloadURL       string    `json:"download_url"`
	ChecksumSHA256    string    `json:"checksum_sha256"`
	Changelog         []string  `json:"changelog"`
	FilesUpdated      []string  `json:"files_updated"`
	MigrationRequired bool      `json:"migration_required"`
}

// TemplateUpdate representa uma atualização de template
type TemplateUpdate struct {
	Name            string   `json:"name"`
	CurrentVersion  string   `json:"current_version"`
	LatestVersion   string   `json:"latest_version"`
	Changelog       []string `json:"changelog"`
	BreakingChanges bool     `json:"breaking_changes"`
}

// SeedVersion representa informações de versão de um seed
type SeedVersion struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	SDKDependency string `json:"sdk_dependency"`
}

// ─────────────────────────────────────────────────────────────────────────
// AUDIT TYPES
// ─────────────────────────────────────────────────────────────────────────

// AuditRequest representa uma solicitação de auditoria
type AuditRequest struct {
	Timestamp  time.Time     `json:"timestamp"`
	SDKVersion string        `json:"sdk_version"`
	SDKModule  string        `json:"sdk_module"`
	Seeds      []SeedVersion `json:"seeds"`
}

// AuditResponse representa a resposta de auditoria
type AuditResponse struct {
	Timestamp          time.Time    `json:"timestamp"`
	AuditID            string       `json:"audit_id"`
	Status             string       `json:"status"`
	SDKVersion         string       `json:"sdk_version"`
	Issues             []AuditIssue `json:"issues"`
	Recommendations    []string     `json:"recommendations"`
	AuditDetails       AuditDetails `json:"audit_details"`
	CompatibilityScore int          `json:"compatibility_score"`
}

// AuditIssue representa um problema encontrado na auditoria
type AuditIssue struct {
	IssueID     string `json:"issue_id"`
	Severity    string `json:"severity"` // critical, warning, info
	Component   string `json:"component"`
	Description string `json:"description"`
	Resolution  string `json:"resolution"`
}

// AuditDetails contém detalhes da auditoria
type AuditDetails struct {
	SeedsAudited       int `json:"seeds_audited"`
	TemplatesAudited   int `json:"templates_audited"`
	VersionMismatches  int `json:"version_mismatches"`
	DeprecatedFeatures int `json:"deprecated_features"`
}

// VersionReport representa um relatório de versão
type VersionReport struct {
	Timestamp           time.Time              `json:"timestamp"`
	CompatibilityMatrix map[string]string      `json:"compatibility_matrix"`
	VersionGraph        map[string]interface{} `json:"version_graph"`
	ReportID            string                 `json:"report_id"`
	SDKVersion          string                 `json:"sdk_version"`
	TemplateVersion     string                 `json:"template_version"`
	Warnings            []string               `json:"warnings"`
	Errors              []string               `json:"errors"`
}

// ─────────────────────────────────────────────────────────────────────────
// MATRIX TYPES
// ─────────────────────────────────────────────────────────────────────────

// MatrixQuery representa uma consulta à matriz de compatibilidade
type MatrixQuery struct {
	Components map[string]string `json:"components"`
	Timestamp  time.Time         `json:"timestamp"`
	QueryType  string            `json:"query_type"`
}

// MatrixResponse representa a resposta da matriz
type MatrixResponse struct {
	Timestamp            time.Time              `json:"timestamp"`
	CompatibilityDetails map[string]interface{} `json:"compatibility_details"`
	QueryID              string                 `json:"query_id"`
	CompatibilityLevel   string                 `json:"compatibility_level"`
	MatrixVersion        string                 `json:"matrix_version"`
	Warnings             []string               `json:"warnings"`
	Compatible           bool                   `json:"compatible"`
}

// ValidationRequest representa uma solicitação de validação de deployment
type ValidationRequest struct {
	Components        map[string]interface{} `json:"components"`
	Timestamp         time.Time              `json:"timestamp"`
	ValidationType    string                 `json:"validation_type"`
	TargetEnvironment string                 `json:"target_environment"`
}

// ValidationResponse representa a resposta de validação
type ValidationResponse struct {
	Timestamp           time.Time         `json:"timestamp"`
	ValidationResults   map[string]string `json:"validation_results"`
	ValidationID        string            `json:"validation_id"`
	CompatibilityStatus string            `json:"compatibility_status"`
	RollbackPlan        RollbackPlan      `json:"rollback_plan"`
	Approved            bool              `json:"approved"`
	DeploymentSafe      bool              `json:"deployment_safe"`
}

// RollbackPlan contém informações do plano de rollback
type RollbackPlan struct {
	PreviousStable string `json:"previous_stable"`
	Available      bool   `json:"available"`
}

// ─────────────────────────────────────────────────────────────────────────
// MANAGER INTERFACES
// ─────────────────────────────────────────────────────────────────────────

// SyncManagerInterface define a interface do gerenciador de sincronização
// TODO: Uncomment when orchestrator is ready
// type SyncManagerInterface interface {
// 	RequestSync(ctx context.Context, seeds []string) (*SyncResponse, error)
// 	SubscribeSeedUpdates(handler func(*SeedUpdate)) error
// }

// AuditManagerInterface define a interface do gerenciador de auditoria
// TODO: Uncomment when orchestrator is ready
// type AuditManagerInterface interface {
// 	RequestAudit(ctx context.Context, seeds []SeedVersion) (*AuditResponse, error)
// 	SubscribeAuditReports(handler func(*VersionReport)) error
// }

// MatrixManagerInterface define a interface do gerenciador de matriz
// TODO: Uncomment when orchestrator is ready
// type MatrixManagerInterface interface {
// 	QueryCompatibility(ctx context.Context, components map[string]string) (*MatrixResponse, error)
// 	ValidateDeployment(ctx context.Context, components map[string]interface{}) (*ValidationResponse, error)
// }
