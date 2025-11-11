package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Publisher struct {
	client *Client
	logger *zap.Logger
}

func NewPublisher(client *Client, logger *zap.Logger) *Publisher {
	return &Publisher{
		client: client,
		logger: logger,
	}
}

// PublishTask publica uma task para ser processada
func (p *Publisher) PublishTask(task map[string]interface{}) error {
	subject := "web.wasm.task.requested"

	if err := p.client.Publish(subject, task); err != nil {
		p.logger.Error("Erro ao publicar task",
			zap.String("subject", subject),
			zap.String("task_id", getString(task, "id")),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar task: %w", err)
	}

	p.logger.Info("Task publicada com sucesso",
		zap.String("subject", subject),
		zap.String("task_id", getString(task, "id")),
		zap.String("type", getString(task, "type")))

	return nil
}

// PublishTaskResult publica o resultado de uma task
func (p *Publisher) PublishTaskResult(taskID, correlationID string, result map[string]interface{}, err error) error {
	subject := "web.wasm.task.completed"

	data := map[string]interface{}{
		"id":             taskID,
		"correlation_id": correlationID,
		"completed_at":   time.Now().UTC(),
		"result":         result,
	}

	if err != nil {
		data["error"] = err.Error()
		data["status"] = "failed"
	} else {
		data["status"] = "completed"
	}

	if pubErr := p.client.Publish(subject, data); pubErr != nil {
		p.logger.Error("Erro ao publicar resultado da task",
			zap.String("subject", subject),
			zap.String("task_id", taskID),
			zap.Error(pubErr))
		return fmt.Errorf("erro ao publicar resultado: %w", pubErr)
	}

	p.logger.Info("Resultado da task publicado",
		zap.String("subject", subject),
		zap.String("task_id", taskID),
		zap.String("status", data["status"].(string)))

	return nil
}

// PublishTaskProgress publica progresso de uma task
func (p *Publisher) PublishTaskProgress(taskID, correlationID string, progress int, message string) error {
	subject := "web.wasm.task.progress"

	data := map[string]interface{}{
		"id":             taskID,
		"correlation_id": correlationID,
		"progress":       progress,
		"message":        message,
		"timestamp":      time.Now().UTC(),
	}

	if err := p.client.Publish(subject, data); err != nil {
		p.logger.Error("Erro ao publicar progresso",
			zap.String("subject", subject),
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar progresso: %w", err)
	}

	p.logger.Debug("Progresso publicado",
		zap.String("task_id", taskID),
		zap.Int("progress", progress))

	return nil
}

// PublishTaskCancel publica cancelamento de task
func (p *Publisher) PublishTaskCancel(taskID string, reason string) error {
	subject := "web.wasm.task.cancelled"

	data := map[string]interface{}{
		"id":           taskID,
		"reason":       reason,
		"cancelled_at": time.Now().UTC(),
	}

	if err := p.client.Publish(subject, data); err != nil {
		p.logger.Error("Erro ao publicar cancelamento",
			zap.String("subject", subject),
			zap.String("task_id", taskID),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar cancelamento: %w", err)
	}

	p.logger.Info("Cancelamento publicado",
		zap.String("task_id", taskID),
		zap.String("reason", reason))

	return nil
}

// PublishWASMEvent publica eventos relacionados ao WASM
func (p *Publisher) PublishWASMEvent(eventType string, data map[string]interface{}) error {
	subject := fmt.Sprintf("web.wasm.%s", eventType)

	eventData := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}

	if err := p.client.Publish(subject, eventData); err != nil {
		p.logger.Error("Erro ao publicar evento WASM",
			zap.String("subject", subject),
			zap.String("event_type", eventType),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar evento WASM: %w", err)
	}

	p.logger.Debug("Evento WASM publicado",
		zap.String("event_type", eventType),
		zap.String("subject", subject))

	return nil
}

// PublishSystemEvent publica eventos do sistema
func (p *Publisher) PublishSystemEvent(eventType string, data map[string]interface{}) error {
	subject := fmt.Sprintf("web.wasm.system.%s", eventType)

	eventData := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().UTC(),
	}

	if err := p.client.Publish(subject, eventData); err != nil {
		p.logger.Error("Erro ao publicar evento de sistema",
			zap.String("subject", subject),
			zap.String("event_type", eventType),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar evento de sistema: %w", err)
	}

	p.logger.Debug("Evento de sistema publicado",
		zap.String("event_type", eventType),
		zap.String("subject", subject))

	return nil
}

// Publish é um método genérico para publicar qualquer mensagem
func (p *Publisher) Publish(subject string, data interface{}) error {
	if err := p.client.Publish(subject, data); err != nil {
		p.logger.Error("Erro ao publicar mensagem",
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar mensagem: %w", err)
	}

	p.logger.Debug("Mensagem publicada", zap.String("subject", subject))
	return nil
}

// PublishAsync publica mensagem assincronamente
func (p *Publisher) PublishAsync(subject string, data interface{}) error {
	if err := p.client.PublishAsync(subject, data); err != nil {
		p.logger.Error("Erro ao publicar mensagem assincronamente",
			zap.String("subject", subject),
			zap.Error(err))
		return fmt.Errorf("erro ao publicar mensagem assincronamente: %w", err)
	}

	p.logger.Debug("Mensagem publicada assincronamente", zap.String("subject", subject))
	return nil
}

// Helper function para extrair string de map de forma segura
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// InitializeStreams cria os streams necessários para o publisher
func (p *Publisher) InitializeStreams() error {
	// Stream para tasks
	taskStream := nats.StreamConfig{
		Name:      "WEB_WASM_TASKS",
		Subjects:  []string{"web.wasm.task.>"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour,
		MaxBytes:  1 << 30, // 1GB
	}

	if err := p.client.CreateStream(taskStream); err != nil {
		return fmt.Errorf("erro ao criar stream de tasks: %w", err)
	}

	// Stream para eventos WASM
	wasmStream := nats.StreamConfig{
		Name:      "WEB_WASM_EVENTS",
		Subjects:  []string{"web.wasm.wasm.>"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    1 * time.Hour,
		MaxBytes:  100 << 20, // 100MB
	}

	if err := p.client.CreateStream(wasmStream); err != nil {
		return fmt.Errorf("erro ao criar stream de eventos WASM: %w", err)
	}

	// Stream para eventos do sistema
	systemStream := nats.StreamConfig{
		Name:      "WEB_WASM_SYSTEM",
		Subjects:  []string{"web.wasm.system.>"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    1 * time.Hour,
		MaxBytes:  50 << 20, // 50MB
	}

	if err := p.client.CreateStream(systemStream); err != nil {
		return fmt.Errorf("erro ao criar stream de eventos do sistema: %w", err)
	}

	p.logger.Info("Streams inicializados com sucesso")
	return nil
}
