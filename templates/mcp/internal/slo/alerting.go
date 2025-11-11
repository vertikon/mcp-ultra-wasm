package slo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AlertSeverity represents different alert severity levels
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// AlertChannel represents different alerting channels
type AlertChannel string

const (
	ChannelSlack     AlertChannel = "slack"
	ChannelEmail     AlertChannel = "email"
	ChannelPagerDuty AlertChannel = "pagerduty"
	ChannelWebhook   AlertChannel = "webhook"
	ChannelDiscord   AlertChannel = "discord"
	ChannelMSTeams   AlertChannel = "msteams"
)

// AlertingConfig holds configuration for the alerting system
type AlertingConfig struct {
	Enabled            bool                      `json:"enabled"`
	DefaultChannels    []AlertChannel            `json:"default_channels"`
	SeverityRouting    map[string][]AlertChannel `json:"severity_routing"`
	ChannelConfigs     map[string]ChannelConfig  `json:"channel_configs"`
	RateLimiting       RateLimitConfig           `json:"rate_limiting"`
	EscalationPolicies []EscalationPolicy        `json:"escalation_policies"`
	SilenceRules       []SilenceRule             `json:"silence_rules"`
}

// ChannelConfig holds configuration for specific alert channels
type ChannelConfig struct {
	Type       AlertChannel      `json:"type"`
	Endpoint   string            `json:"endpoint"`
	Headers    map[string]string `json:"headers"`
	Templates  TemplateConfig    `json:"templates"`
	Enabled    bool              `json:"enabled"`
	Timeout    time.Duration     `json:"timeout"`
	RetryCount int               `json:"retry_count"`
}

// TemplateConfig holds message templates for different channels
type TemplateConfig struct {
	Title     string `json:"title"`
	Message   string `json:"message"`
	Color     string `json:"color"`
	IconEmoji string `json:"icon_emoji"`
	Username  string `json:"username"`
}

// RateLimitConfig configures rate limiting for alerts
type RateLimitConfig struct {
	Enabled    bool          `json:"enabled"`
	WindowSize time.Duration `json:"window_size"`
	MaxAlerts  int           `json:"max_alerts"`
	BurstLimit int           `json:"burst_limit"`
}

// EscalationPolicy defines how alerts should be escalated
type EscalationPolicy struct {
	Name       string           `json:"name"`
	Conditions []string         `json:"conditions"`
	Steps      []EscalationStep `json:"steps"`
	Enabled    bool             `json:"enabled"`
}

// EscalationStep defines a single step in an escalation policy
type EscalationStep struct {
	Duration time.Duration  `json:"duration"`
	Channels []AlertChannel `json:"channels"`
	Message  string         `json:"message"`
}

// SilenceRule defines when alerts should be silenced
type SilenceRule struct {
	Name       string            `json:"name"`
	Conditions map[string]string `json:"conditions"`
	StartTime  string            `json:"start_time"`
	EndTime    string            `json:"end_time"`
	Weekdays   []time.Weekday    `json:"weekdays"`
	Enabled    bool              `json:"enabled"`
}

// AlertManager manages SLO-based alerting
type AlertManager struct {
	config     AlertingConfig
	logger     *zap.Logger
	httpClient *http.Client

	// State management
	alertHistory map[string][]AlertEvent
	rateLimiter  map[string][]time.Time
	silences     map[string]time.Time
	mu           sync.RWMutex

	// Channels
	alertChan chan AlertEvent
	stopChan  chan struct{}
}

// NewAlertManager creates a new alert manager
func NewAlertManager(config AlertingConfig, logger *zap.Logger) *AlertManager {
	return &AlertManager{
		config:       config,
		logger:       logger,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		alertHistory: make(map[string][]AlertEvent),
		rateLimiter:  make(map[string][]time.Time),
		silences:     make(map[string]time.Time),
		alertChan:    make(chan AlertEvent, 1000),
		stopChan:     make(chan struct{}),
	}
}

// Start begins the alert processing
func (am *AlertManager) Start(ctx context.Context) error {
	if !am.config.Enabled {
		am.logger.Info("Alert manager is disabled")
		return nil
	}

	am.logger.Info("Starting alert manager")

	// Start alert processing goroutine
	go am.processAlerts(ctx)

	// Start cleanup goroutine
	go am.cleanup(ctx)

	return nil
}

// Stop stops the alert manager
func (am *AlertManager) Stop() {
	close(am.stopChan)
}

// SendAlert queues an alert for processing
func (am *AlertManager) SendAlert(alert AlertEvent) {
	select {
	case am.alertChan <- alert:
	default:
		am.logger.Warn("Alert channel full, dropping alert",
			zap.String("slo", alert.SLOName),
			zap.String("type", alert.Type))
	}
}

// processAlerts processes incoming alerts
func (am *AlertManager) processAlerts(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			am.logger.Info("Alert processing stopped by context")
			return
		case <-am.stopChan:
			am.logger.Info("Alert processing stopped")
			return
		case alert := <-am.alertChan:
			if err := am.processAlert(alert); err != nil {
				am.logger.Error("Failed to process alert",
					zap.String("slo", alert.SLOName),
					zap.String("type", alert.Type),
					zap.Error(err))
			}
		}
	}
}

// processAlert processes a single alert
func (am *AlertManager) processAlert(alert AlertEvent) error {
	// Check if alert should be silenced
	if am.shouldSilence(alert) {
		am.logger.Debug("Alert silenced",
			zap.String("slo", alert.SLOName),
			zap.String("type", alert.Type))
		return nil
	}

	// Apply rate limiting
	if am.isRateLimited(alert) {
		am.logger.Debug("Alert rate limited",
			zap.String("slo", alert.SLOName),
			zap.String("type", alert.Type))
		return nil
	}

	// Store in history
	am.storeAlertHistory(alert)

	// Determine channels based on severity
	channels := am.getChannelsForSeverity(alert.Severity)

	// Send to each channel
	var errors []error
	for _, channel := range channels {
		if err := am.sendToChannel(alert, channel); err != nil {
			errors = append(errors, fmt.Errorf("channel %s: %w", channel, err))
		}
	}

	// Start escalation if configured
	am.startEscalation(alert)

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to some channels: %v", errors)
	}

	return nil
}

// shouldSilence checks if an alert should be silenced
func (am *AlertManager) shouldSilence(alert AlertEvent) bool {
	now := time.Now()

	for _, rule := range am.config.SilenceRules {
		if !rule.Enabled {
			continue
		}

		// Check conditions
		matches := true
		for key, value := range rule.Conditions {
			switch key {
			case "slo_name":
				if alert.SLOName != value {
					matches = false
				}
			case "severity":
				if alert.Severity != value {
					matches = false
				}
			case "type":
				if alert.Type != value {
					matches = false
				}
			}
		}

		if !matches {
			continue
		}

		// Check time windows
		if len(rule.Weekdays) > 0 {
			found := false
			for _, weekday := range rule.Weekdays {
				if now.Weekday() == weekday {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Check time of day
		if rule.StartTime != "" && rule.EndTime != "" {
			start, err := time.Parse("15:04", rule.StartTime)
			if err != nil {
				continue
			}
			end, err := time.Parse("15:04", rule.EndTime)
			if err != nil {
				continue
			}

			currentTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
			startTime := time.Date(0, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
			endTime := time.Date(0, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)

			if currentTime.Before(startTime) || currentTime.After(endTime) {
				continue
			}
		}

		// Rule matches, silence the alert
		return true
	}

	return false
}

// isRateLimited checks if an alert is rate limited
func (am *AlertManager) isRateLimited(alert AlertEvent) bool {
	if !am.config.RateLimiting.Enabled {
		return false
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	key := fmt.Sprintf("%s:%s", alert.SLOName, alert.Type)
	now := time.Now()
	windowStart := now.Add(-am.config.RateLimiting.WindowSize)

	// Clean old entries
	var recentAlerts []time.Time
	for _, alertTime := range am.rateLimiter[key] {
		if alertTime.After(windowStart) {
			recentAlerts = append(recentAlerts, alertTime)
		}
	}

	// Check if we're at the limit
	if len(recentAlerts) >= am.config.RateLimiting.MaxAlerts {
		return true
	}

	// Add current alert
	recentAlerts = append(recentAlerts, now)
	am.rateLimiter[key] = recentAlerts

	return false
}

// storeAlertHistory stores alert in history
func (am *AlertManager) storeAlertHistory(alert AlertEvent) {
	am.mu.Lock()
	defer am.mu.Unlock()

	key := alert.SLOName
	am.alertHistory[key] = append(am.alertHistory[key], alert)

	// Keep only recent history (last 100 alerts per SLO)
	if len(am.alertHistory[key]) > 100 {
		am.alertHistory[key] = am.alertHistory[key][len(am.alertHistory[key])-100:]
	}
}

// getChannelsForSeverity returns channels for a given severity
func (am *AlertManager) getChannelsForSeverity(severity string) []AlertChannel {
	if channels, exists := am.config.SeverityRouting[severity]; exists {
		return channels
	}
	return am.config.DefaultChannels
}

// sendToChannel sends an alert to a specific channel
func (am *AlertManager) sendToChannel(alert AlertEvent, channel AlertChannel) error {
	channelKey := string(channel)
	config, exists := am.config.ChannelConfigs[channelKey]
	if !exists || !config.Enabled {
		return fmt.Errorf("channel %s not configured or disabled", channel)
	}

	switch channel {
	case ChannelSlack:
		return am.sendToSlack(alert, config)
	case ChannelEmail:
		return am.sendToEmail(alert, config)
	case ChannelPagerDuty:
		return am.sendToPagerDuty(alert, config)
	case ChannelWebhook:
		return am.sendToWebhook(alert, config)
	case ChannelDiscord:
		return am.sendToDiscord(alert, config)
	case ChannelMSTeams:
		return am.sendToMSTeams(alert, config)
	default:
		return fmt.Errorf("unsupported channel: %s", channel)
	}
}

// sendToSlack sends alert to Slack
func (am *AlertManager) sendToSlack(alert AlertEvent, config ChannelConfig) error {
	color := am.getSeverityColor(alert.Severity)
	if config.Templates.Color != "" {
		color = config.Templates.Color
	}

	title := am.renderTemplate(config.Templates.Title, alert)
	message := am.renderTemplate(config.Templates.Message, alert)

	payload := map[string]interface{}{
		"username":   config.Templates.Username,
		"icon_emoji": config.Templates.IconEmoji,
		"attachments": []map[string]interface{}{
			{
				"color":     color,
				"title":     title,
				"text":      message,
				"timestamp": alert.Timestamp.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "SLO",
						"value": alert.SLOName,
						"short": true,
					},
					{
						"title": "Severity",
						"value": alert.Severity,
						"short": true,
					},
					{
						"title": "Type",
						"value": alert.Type,
						"short": true,
					},
				},
			},
		},
	}

	return am.sendHTTPPayload(config.Endpoint, payload, config.Headers, config.Timeout)
}

// sendToDiscord sends alert to Discord
func (am *AlertManager) sendToDiscord(alert AlertEvent, config ChannelConfig) error {
	color := am.getSeverityColorInt(alert.Severity)
	title := am.renderTemplate(config.Templates.Title, alert)
	message := am.renderTemplate(config.Templates.Message, alert)

	payload := map[string]interface{}{
		"username": config.Templates.Username,
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": message,
				"color":       color,
				"timestamp":   alert.Timestamp.Format(time.RFC3339),
				"fields": []map[string]interface{}{
					{
						"name":   "SLO",
						"value":  alert.SLOName,
						"inline": true,
					},
					{
						"name":   "Severity",
						"value":  alert.Severity,
						"inline": true,
					},
					{
						"name":   "Type",
						"value":  alert.Type,
						"inline": true,
					},
				},
			},
		},
	}

	return am.sendHTTPPayload(config.Endpoint, payload, config.Headers, config.Timeout)
}

// sendToWebhook sends alert to a generic webhook
func (am *AlertManager) sendToWebhook(alert AlertEvent, config ChannelConfig) error {
	payload := map[string]interface{}{
		"slo_name":    alert.SLOName,
		"type":        alert.Type,
		"severity":    alert.Severity,
		"message":     alert.Message,
		"timestamp":   alert.Timestamp,
		"labels":      alert.Labels,
		"annotations": alert.Annotations,
	}

	return am.sendHTTPPayload(config.Endpoint, payload, config.Headers, config.Timeout)
}

// sendToEmail sends alert via email (placeholder implementation)
func (am *AlertManager) sendToEmail(alert AlertEvent, config ChannelConfig) error {
	am.logger.Info("Email alert sent (placeholder)",
		zap.String("slo", alert.SLOName),
		zap.String("severity", alert.Severity))
	return nil // TODO: Implement actual email sending
}

// sendToPagerDuty sends alert to PagerDuty (placeholder implementation)
func (am *AlertManager) sendToPagerDuty(alert AlertEvent, config ChannelConfig) error {
	am.logger.Info("PagerDuty alert sent (placeholder)",
		zap.String("slo", alert.SLOName),
		zap.String("severity", alert.Severity))
	return nil // TODO: Implement actual PagerDuty integration
}

// sendToMSTeams sends alert to Microsoft Teams (placeholder implementation)
func (am *AlertManager) sendToMSTeams(alert AlertEvent, config ChannelConfig) error {
	am.logger.Info("MS Teams alert sent (placeholder)",
		zap.String("slo", alert.SLOName),
		zap.String("severity", alert.Severity))
	return nil // TODO: Implement actual MS Teams integration
}

// sendHTTPPayload sends a JSON payload via HTTP POST
func (am *AlertManager) sendHTTPPayload(endpoint string, payload interface{}, headers map[string]string, timeout time.Duration) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			am.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	return nil
}

// startEscalation starts escalation process for an alert
func (am *AlertManager) startEscalation(alert AlertEvent) {
	for _, policy := range am.config.EscalationPolicies {
		if !policy.Enabled {
			continue
		}

		// Check if alert matches escalation conditions
		matches := false
		for _, condition := range policy.Conditions {
			if strings.Contains(condition, alert.SLOName) ||
				strings.Contains(condition, alert.Severity) ||
				strings.Contains(condition, alert.Type) {
				matches = true
				break
			}
		}

		if matches {
			go am.executeEscalation(alert, policy)
			break // Only one escalation policy per alert
		}
	}
}

// executeEscalation executes an escalation policy
func (am *AlertManager) executeEscalation(alert AlertEvent, policy EscalationPolicy) {
	for i, step := range policy.Steps {
		// Wait for step duration (except for first step)
		if i > 0 {
			time.Sleep(step.Duration)
		}

		// Check if alert is resolved
		// TODO: Implement resolution checking

		// Send escalation alert
		escalationAlert := alert
		escalationAlert.Message = step.Message
		escalationAlert.Type = fmt.Sprintf("%s_escalation_step_%d", alert.Type, i+1)

		for _, channel := range step.Channels {
			if err := am.sendToChannel(escalationAlert, channel); err != nil {
				am.logger.Error("Failed to send escalation alert",
					zap.String("policy", policy.Name),
					zap.Int("step", i+1),
					zap.String("channel", string(channel)),
					zap.Error(err))
			}
		}
	}
}

// cleanup performs periodic cleanup of old data
func (am *AlertManager) cleanup(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-am.stopChan:
			return
		case <-ticker.C:
			am.performCleanup()
		}
	}
}

// performCleanup cleans up old rate limiter and history data
func (am *AlertManager) performCleanup() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Keep 24 hours of data

	// Cleanup rate limiter
	for key, alerts := range am.rateLimiter {
		var recent []time.Time
		for _, alertTime := range alerts {
			if alertTime.After(cutoff) {
				recent = append(recent, alertTime)
			}
		}
		am.rateLimiter[key] = recent
	}

	// Cleanup silences
	for key, silenceTime := range am.silences {
		if silenceTime.Before(cutoff) {
			delete(am.silences, key)
		}
	}
}

// Helper methods

func (am *AlertManager) renderTemplate(template string, alert AlertEvent) string {
	if template == "" {
		return fmt.Sprintf("[%s] %s: %s", alert.Severity, alert.SLOName, alert.Message)
	}

	// Simple template rendering - replace placeholders
	result := template
	result = strings.ReplaceAll(result, "{{.SLOName}}", alert.SLOName)
	result = strings.ReplaceAll(result, "{{.Type}}", alert.Type)
	result = strings.ReplaceAll(result, "{{.Severity}}", alert.Severity)
	result = strings.ReplaceAll(result, "{{.Message}}", alert.Message)
	result = strings.ReplaceAll(result, "{{.Timestamp}}", alert.Timestamp.Format(time.RFC3339))

	return result
}

func (am *AlertManager) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case string(SeverityCritical):
		return "danger"
	case string(SeverityWarning):
		return string(SeverityWarning)
	case "info":
		return "good"
	default:
		return "#808080"
	}
}

func (am *AlertManager) getSeverityColorInt(severity string) int {
	switch strings.ToLower(severity) {
	case string(SeverityCritical):
		return 0xFF0000 // Red
	case string(SeverityWarning):
		return 0xFFA500 // Orange
	case "info":
		return 0x00FF00 // Green
	default:
		return 0x808080 // Gray
	}
}

// GetAlertHistory returns alert history for an SLO
func (am *AlertManager) GetAlertHistory(sloName string) []AlertEvent {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.alertHistory[sloName]
}

// GetAllAlertHistory returns all alert history
func (am *AlertManager) GetAllAlertHistory() map[string][]AlertEvent {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make(map[string][]AlertEvent)
	for key, alerts := range am.alertHistory {
		result[key] = alerts
	}
	return result
}
