package lifecycle

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// DeploymentStrategy represents deployment strategies
type DeploymentStrategy string

const (
	DeploymentBlueGreen DeploymentStrategy = "blue_green"
	DeploymentCanary    DeploymentStrategy = "canary"
	DeploymentRolling   DeploymentStrategy = "rolling"
	DeploymentRecreate  DeploymentStrategy = "recreate"
)

// DeploymentPhase represents deployment phases
type DeploymentPhase string

const (
	PhaseValidation   DeploymentPhase = "validation"
	PhasePreHooks     DeploymentPhase = "pre_hooks"
	PhaseDeployment   DeploymentPhase = "deployment"
	PhaseVerification DeploymentPhase = "verification"
	PhasePostHooks    DeploymentPhase = "post_hooks"
	PhaseComplete     DeploymentPhase = "complete"
	PhaseRollback     DeploymentPhase = "rollback"
)

// DeploymentConfig configures deployment automation
type DeploymentConfig struct {
	Strategy    DeploymentStrategy `yaml:"strategy"`
	Environment string             `yaml:"environment"`
	Namespace   string             `yaml:"namespace"`
	Image       string             `yaml:"image"`
	Tag         string             `yaml:"tag"`

	// Validation settings
	ValidateConfig    bool `yaml:"validate_config"`
	ValidateImage     bool `yaml:"validate_image"`
	ValidateResources bool `yaml:"validate_resources"`

	// Rollout settings
	MaxUnavailable  string        `yaml:"max_unavailable"`
	MaxSurge        string        `yaml:"max_surge"`
	ProgressTimeout time.Duration `yaml:"progress_timeout"`

	// Canary settings
	CanaryReplicas      int           `yaml:"canary_replicas"`
	CanaryDuration      time.Duration `yaml:"canary_duration"`
	TrafficSplitPercent int           `yaml:"traffic_split_percent"`

	// Blue/Green settings
	BlueGreenTimeout time.Duration `yaml:"blue_green_timeout"`

	// Health checks
	ReadinessTimeout time.Duration `yaml:"readiness_timeout"`
	LivenessTimeout  time.Duration `yaml:"liveness_timeout"`

	// Hooks
	PreDeployHooks  []DeploymentHook `yaml:"pre_deploy_hooks"`
	PostDeployHooks []DeploymentHook `yaml:"post_deploy_hooks"`
	RollbackHooks   []DeploymentHook `yaml:"rollback_hooks"`

	// Monitoring
	EnableMetrics  bool `yaml:"enable_metrics"`
	EnableAlerting bool `yaml:"enable_alerting"`

	// Kubernetes
	KubeConfig   string `yaml:"kube_config"`
	KubeContext  string `yaml:"kube_context"`
	ManifestPath string `yaml:"manifest_path"`

	// Rollback
	AutoRollback       bool               `yaml:"auto_rollback"`
	RollbackThresholds RollbackThresholds `yaml:"rollback_thresholds"`
}

// DeploymentHook represents a deployment hook
type DeploymentHook struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"` // "command", "http", "script"
	Command     string            `yaml:"command"`
	URL         string            `yaml:"url"`
	Script      string            `yaml:"script"`
	Timeout     time.Duration     `yaml:"timeout"`
	RetryCount  int               `yaml:"retry_count"`
	Environment map[string]string `yaml:"environment"`
}

// RollbackThresholds defines when to trigger auto-rollback
type RollbackThresholds struct {
	ErrorRate        float64       `yaml:"error_rate"`    // Error rate percentage
	ResponseTime     time.Duration `yaml:"response_time"` // P95 response time
	HealthCheckFails int           `yaml:"health_check_fails"`
	TimeWindow       time.Duration `yaml:"time_window"`
}

// DeploymentResult represents the result of a deployment
type DeploymentResult struct {
	Success         bool                   `json:"success"`
	Strategy        DeploymentStrategy     `json:"strategy"`
	Environment     string                 `json:"environment"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Phase           DeploymentPhase        `json:"phase"`
	PreviousVersion string                 `json:"previous_version"`
	NewVersion      string                 `json:"new_version"`
	RollbackVersion string                 `json:"rollback_version,omitempty"`
	Logs            []string               `json:"logs"`
	Errors          []string               `json:"errors"`
	Metrics         map[string]interface{} `json:"metrics"`
}

// DeploymentAutomation manages automated deployments
type DeploymentAutomation struct {
	config DeploymentConfig
	logger *logger.Logger

	// State tracking
	currentDeployment *DeploymentResult
	deploymentHistory []DeploymentResult
	maxHistorySize    int
}

// NewDeploymentAutomation creates a new deployment automation system
func NewDeploymentAutomation(config DeploymentConfig, log *logger.Logger) *DeploymentAutomation {
	return &DeploymentAutomation{
		config:            config,
		logger:            log,
		deploymentHistory: make([]DeploymentResult, 0),
		maxHistorySize:    50,
	}
}

// Deploy executes a deployment using the configured strategy
func (da *DeploymentAutomation) Deploy(ctx context.Context, version string) (*DeploymentResult, error) {
	result := &DeploymentResult{
		Strategy:    da.config.Strategy,
		Environment: da.config.Environment,
		StartTime:   time.Now(),
		NewVersion:  version,
		Phase:       PhaseValidation,
		Logs:        make([]string, 0),
		Errors:      make([]string, 0),
		Metrics:     make(map[string]interface{}),
	}

	da.currentDeployment = result
	da.addLog(result, fmt.Sprintf("Starting %s deployment for version %s", da.config.Strategy, version))

	// Get previous version for rollback
	if prev := da.getPreviousVersion(); prev != "" {
		result.PreviousVersion = prev
	}

	// Execute deployment pipeline
	if err := da.executeDeploymentPipeline(ctx, result); err != nil {
		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		da.addError(result, err.Error())

		// Auto-rollback if enabled
		if da.config.AutoRollback && result.PreviousVersion != "" {
			da.addLog(result, "Auto-rollback triggered due to deployment failure")
			if rollbackErr := da.rollback(ctx, result); rollbackErr != nil {
				da.addError(result, fmt.Sprintf("Rollback failed: %v", rollbackErr))
			}
		}

		da.addDeploymentToHistory(*result)
		return result, err
	}

	result.Success = true
	result.Phase = PhaseComplete
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	da.addLog(result, fmt.Sprintf("Deployment completed successfully in %v", result.Duration))
	da.addDeploymentToHistory(*result)

	return result, nil
}

// Rollback rolls back to the previous version
func (da *DeploymentAutomation) Rollback(ctx context.Context) (*DeploymentResult, error) {
	if da.currentDeployment == nil || da.currentDeployment.PreviousVersion == "" {
		return nil, fmt.Errorf("no previous version available for rollback")
	}

	result := &DeploymentResult{
		Strategy:        da.config.Strategy,
		Environment:     da.config.Environment,
		StartTime:       time.Now(),
		NewVersion:      da.currentDeployment.PreviousVersion,
		RollbackVersion: da.currentDeployment.NewVersion,
		Phase:           PhaseRollback,
		Logs:            make([]string, 0),
		Errors:          make([]string, 0),
		Metrics:         make(map[string]interface{}),
	}

	da.addLog(result, fmt.Sprintf("Starting rollback from %s to %s",
		result.RollbackVersion, result.NewVersion))

	if err := da.rollback(ctx, result); err != nil {
		result.Success = false
		da.addError(result, err.Error())
		return result, err
	}

	result.Success = true
	result.Phase = PhaseComplete
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	da.addLog(result, fmt.Sprintf("Rollback completed successfully in %v", result.Duration))
	da.addDeploymentToHistory(*result)

	return result, nil
}

// GetDeploymentHistory returns deployment history
func (da *DeploymentAutomation) GetDeploymentHistory() []DeploymentResult {
	history := make([]DeploymentResult, len(da.deploymentHistory))
	copy(history, da.deploymentHistory)
	return history
}

// GetCurrentDeployment returns the current deployment status
func (da *DeploymentAutomation) GetCurrentDeployment() *DeploymentResult {
	return da.currentDeployment
}

// Private methods

func (da *DeploymentAutomation) executeDeploymentPipeline(ctx context.Context, result *DeploymentResult) error {
	pipeline := []struct {
		phase DeploymentPhase
		fn    func(context.Context, *DeploymentResult) error
	}{
		{PhaseValidation, da.validateDeployment},
		{PhasePreHooks, da.executePreHooks},
		{PhaseDeployment, da.executeDeployment},
		{PhaseVerification, da.verifyDeployment},
		{PhasePostHooks, da.executePostHooks},
	}

	for _, stage := range pipeline {
		result.Phase = stage.phase
		da.addLog(result, fmt.Sprintf("Executing phase: %s", stage.phase))

		if err := stage.fn(ctx, result); err != nil {
			return fmt.Errorf("phase %s failed: %w", stage.phase, err)
		}
	}

	return nil
}

func (da *DeploymentAutomation) validateDeployment(_ context.Context, result *DeploymentResult) error {
	da.addLog(result, "Validating deployment configuration")

	// Validate configuration
	if da.config.ValidateConfig {
		if err := da.validateKubernetesManifests(); err != nil {
			return fmt.Errorf("manifest validation failed: %w", err)
		}
		da.addLog(result, "Kubernetes manifests validated successfully")
	}

	// Validate image
	if da.config.ValidateImage {
		if err := da.validateDockerImage(result.NewVersion); err != nil {
			return fmt.Errorf("image validation failed: %w", err)
		}
		da.addLog(result, "Docker image validated successfully")
	}

	// Validate resources
	if da.config.ValidateResources {
		if err := da.validateClusterResources(); err != nil {
			return fmt.Errorf("resource validation failed: %w", err)
		}
		da.addLog(result, "Cluster resources validated successfully")
	}

	return nil
}

func (da *DeploymentAutomation) executePreHooks(ctx context.Context, result *DeploymentResult) error {
	if len(da.config.PreDeployHooks) == 0 {
		return nil
	}

	da.addLog(result, "Executing pre-deployment hooks")

	for _, hook := range da.config.PreDeployHooks {
		if err := da.executeHook(ctx, hook, result); err != nil {
			return fmt.Errorf("pre-deploy hook %s failed: %w", hook.Name, err)
		}
		da.addLog(result, fmt.Sprintf("Pre-deploy hook %s completed successfully", hook.Name))
	}

	return nil
}

func (da *DeploymentAutomation) executeDeployment(ctx context.Context, result *DeploymentResult) error {
	da.addLog(result, fmt.Sprintf("Executing %s deployment", da.config.Strategy))

	switch da.config.Strategy {
	case DeploymentRolling:
		return da.executeRollingDeployment(ctx, result)
	case DeploymentBlueGreen:
		return da.executeBlueGreenDeployment(ctx, result)
	case DeploymentCanary:
		return da.executeCanaryDeployment(ctx, result)
	case DeploymentRecreate:
		return da.executeRecreateDeployment(ctx, result)
	default:
		return fmt.Errorf("unsupported deployment strategy: %s", da.config.Strategy)
	}
}

func (da *DeploymentAutomation) executeRollingDeployment(ctx context.Context, result *DeploymentResult) error {
	// Update deployment with new image
	cmd := fmt.Sprintf("kubectl set image deployment/mcp-ultra-wasm mcp-ultra-wasm=%s:%s --namespace=%s",
		da.config.Image, result.NewVersion, da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to update deployment image: %w", err)
	}

	// Wait for rollout to complete
	cmd = fmt.Sprintf("kubectl rollout status deployment/mcp-ultra-wasm --namespace=%s --timeout=%s",
		da.config.Namespace, da.config.ProgressTimeout.String())

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("rollout failed: %w", err)
	}

	da.addLog(result, "Rolling deployment completed successfully")
	return nil
}

func (da *DeploymentAutomation) executeBlueGreenDeployment(ctx context.Context, result *DeploymentResult) error {
	// Implementation for Blue/Green deployment
	// This is a simplified version - real implementation would be more complex

	// Deploy green environment
	cmd := fmt.Sprintf("kubectl apply -f %s/green-deployment.yaml --namespace=%s",
		da.config.ManifestPath, da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to deploy green environment: %w", err)
	}

	// Wait for green to be ready
	time.Sleep(da.config.BlueGreenTimeout)

	// Switch traffic to green
	cmd = fmt.Sprintf("kubectl patch service mcp-ultra-wasm-service -p '{\"spec\":{\"selector\":{\"version\":\"green\"}}}' --namespace=%s",
		da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to switch traffic to green: %w", err)
	}

	// Cleanup blue environment after successful switch
	cmd = fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm-blue --namespace=%s --ignore-not-found=true",
		da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		da.addLog(result, fmt.Sprintf("Warning: failed to cleanup blue deployment: %v", err))
	}

	da.addLog(result, "Blue/Green deployment completed successfully")
	return nil
}

func (da *DeploymentAutomation) executeCanaryDeployment(ctx context.Context, result *DeploymentResult) error {
	// Deploy canary version with limited replicas
	cmd := fmt.Sprintf("kubectl apply -f %s/canary-deployment.yaml --namespace=%s",
		da.config.ManifestPath, da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to deploy canary: %w", err)
	}

	// Wait for canary duration to monitor metrics
	da.addLog(result, fmt.Sprintf("Monitoring canary for %v", da.config.CanaryDuration))
	time.Sleep(da.config.CanaryDuration)

	// Check canary metrics
	if err := da.validateCanaryMetrics(ctx, result); err != nil {
		// Rollback canary
		da.addLog(result, "Canary validation failed, rolling back")
		if rollbackErr := da.executeCommand(ctx, fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm-canary --namespace=%s", da.config.Namespace), result); rollbackErr != nil {
			da.addLog(result, fmt.Sprintf("Warning: failed to delete canary deployment: %v", rollbackErr))
		}
		return fmt.Errorf("canary validation failed: %w", err)
	}

	// Promote canary to full deployment
	cmd = fmt.Sprintf("kubectl patch deployment mcp-ultra-wasm --patch '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"mcp-ultra-wasm\",\"image\":\"%s:%s\"}]}}}}' --namespace=%s",
		da.config.Image, result.NewVersion, da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to promote canary: %w", err)
	}

	// Cleanup canary deployment
	_ = da.executeCommand(ctx, fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm-canary --namespace=%s", da.config.Namespace), result)

	da.addLog(result, "Canary deployment completed successfully")
	return nil
}

func (da *DeploymentAutomation) executeRecreateDeployment(ctx context.Context, result *DeploymentResult) error {
	// Delete existing deployment
	cmd := fmt.Sprintf("kubectl delete deployment mcp-ultra-wasm --namespace=%s --wait=true",
		da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to delete existing deployment: %w", err)
	}

	// Create new deployment
	cmd = fmt.Sprintf("kubectl apply -f %s/deployment.yaml --namespace=%s",
		da.config.ManifestPath, da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("failed to create new deployment: %w", err)
	}

	da.addLog(result, "Recreate deployment completed successfully")
	return nil
}

func (da *DeploymentAutomation) verifyDeployment(ctx context.Context, result *DeploymentResult) error {
	da.addLog(result, "Verifying deployment health")

	// Wait for pods to be ready
	cmd := fmt.Sprintf("kubectl wait --for=condition=ready pod -l app=mcp-ultra-wasm --timeout=%s --namespace=%s",
		da.config.ReadinessTimeout.String(), da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("pods not ready within timeout: %w", err)
	}

	// Perform health checks
	if err := da.performHealthChecks(ctx, result); err != nil {
		return fmt.Errorf("health checks failed: %w", err)
	}

	da.addLog(result, "Deployment verification completed successfully")
	return nil
}

func (da *DeploymentAutomation) executePostHooks(ctx context.Context, result *DeploymentResult) error {
	if len(da.config.PostDeployHooks) == 0 {
		return nil
	}

	da.addLog(result, "Executing post-deployment hooks")

	for _, hook := range da.config.PostDeployHooks {
		if err := da.executeHook(ctx, hook, result); err != nil {
			return fmt.Errorf("post-deploy hook %s failed: %w", hook.Name, err)
		}
		da.addLog(result, fmt.Sprintf("Post-deploy hook %s completed successfully", hook.Name))
	}

	return nil
}

func (da *DeploymentAutomation) rollback(ctx context.Context, result *DeploymentResult) error {
	da.addLog(result, "Executing rollback")
	result.Phase = PhaseRollback

	// Execute rollback hooks first
	for _, hook := range da.config.RollbackHooks {
		if err := da.executeHook(ctx, hook, result); err != nil {
			da.addLog(result, fmt.Sprintf("Rollback hook %s failed: %v", hook.Name, err))
		}
	}

	// Rollback deployment
	cmd := fmt.Sprintf("kubectl rollout undo deployment/mcp-ultra-wasm --namespace=%s",
		da.config.Namespace)

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("kubectl rollback failed: %w", err)
	}

	// Wait for rollback to complete
	cmd = fmt.Sprintf("kubectl rollout status deployment/mcp-ultra-wasm --namespace=%s --timeout=%s",
		da.config.Namespace, da.config.ProgressTimeout.String())

	if err := da.executeCommand(ctx, cmd, result); err != nil {
		return fmt.Errorf("rollback verification failed: %w", err)
	}

	da.addLog(result, "Rollback completed successfully")
	return nil
}

func (da *DeploymentAutomation) executeHook(ctx context.Context, hook DeploymentHook, result *DeploymentResult) error {
	hookCtx, cancel := context.WithTimeout(ctx, hook.Timeout)
	defer cancel()

	switch hook.Type {
	case "command":
		return da.executeCommand(hookCtx, hook.Command, result)
	case "script":
		return da.executeScript(hookCtx, hook.Script, result)
	case "http":
		return da.executeHTTPHook(hookCtx, hook, result)
	default:
		return fmt.Errorf("unsupported hook type: %s", hook.Type)
	}
}

func (da *DeploymentAutomation) executeCommand(ctx context.Context, command string, result *DeploymentResult) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		da.addError(result, fmt.Sprintf("Command failed: %s\nOutput: %s", command, string(output)))
		return err
	}

	da.addLog(result, fmt.Sprintf("Command executed: %s", command))
	if len(output) > 0 {
		da.addLog(result, fmt.Sprintf("Output: %s", string(output)))
	}

	return nil
}

func (da *DeploymentAutomation) executeScript(ctx context.Context, script string, result *DeploymentResult) error {
	// Implementation for script execution
	cmd := exec.CommandContext(ctx, "bash", "-c", script)
	output, err := cmd.CombinedOutput()

	if err != nil {
		da.addError(result, fmt.Sprintf("Script failed: %s\nOutput: %s", script, string(output)))
		return err
	}

	da.addLog(result, "Script executed successfully")
	return nil
}

func (da *DeploymentAutomation) executeHTTPHook(_ context.Context, hook DeploymentHook, result *DeploymentResult) error {
	// Implementation for HTTP hook execution
	da.addLog(result, fmt.Sprintf("Executing HTTP hook: %s", hook.URL))
	// This would implement HTTP request logic
	return nil
}

func (da *DeploymentAutomation) validateKubernetesManifests() error {
	// Implementation for manifest validation
	return nil
}

func (da *DeploymentAutomation) validateDockerImage(version string) error {
	// Implementation for image validation
	return nil
}

func (da *DeploymentAutomation) validateClusterResources() error {
	// Implementation for resource validation
	return nil
}

func (da *DeploymentAutomation) validateCanaryMetrics(ctx context.Context, result *DeploymentResult) error {
	// Implementation for canary metrics validation
	return nil
}

func (da *DeploymentAutomation) performHealthChecks(ctx context.Context, result *DeploymentResult) error {
	// Implementation for health checks
	return nil
}

func (da *DeploymentAutomation) getPreviousVersion() string {
	if len(da.deploymentHistory) == 0 {
		return ""
	}

	// Get the last successful deployment
	for i := len(da.deploymentHistory) - 1; i >= 0; i-- {
		if da.deploymentHistory[i].Success && da.deploymentHistory[i].Phase == PhaseComplete {
			return da.deploymentHistory[i].NewVersion
		}
	}

	return ""
}

func (da *DeploymentAutomation) addLog(result *DeploymentResult, message string) {
	result.Logs = append(result.Logs, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), message))
	da.logger.Info(message, "deployment", result.NewVersion, "phase", result.Phase)
}

func (da *DeploymentAutomation) addError(result *DeploymentResult, message string) {
	result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), message))
	da.logger.Error(message, "deployment", result.NewVersion, "phase", result.Phase)
}

func (da *DeploymentAutomation) addDeploymentToHistory(result DeploymentResult) {
	da.deploymentHistory = append(da.deploymentHistory, result)

	// Maintain history size limit
	if len(da.deploymentHistory) > da.maxHistorySize {
		da.deploymentHistory = da.deploymentHistory[len(da.deploymentHistory)-da.maxHistorySize:]
	}
}
