package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
)

// NATSEventBus implements EventBus using NATS
type NATSEventBus struct {
	conn   *nats.Conn
	logger *zap.Logger
}

// NewNATSEventBus creates a new NATS event bus
func NewNATSEventBus(natsURL string, logger *zap.Logger) (*NATSEventBus, error) {
	conn, err := nats.Connect(natsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
		nats.ReconnectBufSize(-1),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				logger.Error("NATS disconnected", zap.Error(err))
			} else {
				logger.Info("NATS disconnected")
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS reconnected", zap.String("url", nc.ConnectedUrl()))
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	return &NATSEventBus{
		conn:   conn,
		logger: logger,
	}, nil
}

// Publish publishes an event to NATS
func (bus *NATSEventBus) Publish(_ context.Context, event *domain.Event) error {
	subject := fmt.Sprintf("events.%s", event.Type)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	// Create NATS message with headers
	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header: nats.Header{
			"Event-ID":      []string{event.ID.String()},
			"Event-Type":    []string{event.Type},
			"Aggregate-ID":  []string{event.AggregateID.String()},
			"Event-Version": []string{fmt.Sprintf("%d", event.Version)},
			"Content-Type":  []string{"application/json"},
			"Timestamp":     []string{event.OccurredAt.Format(time.RFC3339)},
		},
	}

	if err := bus.conn.PublishMsg(msg); err != nil {
		return fmt.Errorf("publishing event to NATS: %w", err)
	}

	bus.logger.Debug("Event published",
		zap.String("event_id", event.ID.String()),
		zap.String("event_type", event.Type),
		zap.String("subject", subject))

	return nil
}

// Subscribe subscribes to events of a specific type
func (bus *NATSEventBus) Subscribe(eventType string, handler EventHandler) (*nats.Subscription, error) {
	subject := fmt.Sprintf("events.%s", eventType)

	sub, err := bus.conn.Subscribe(subject, func(msg *nats.Msg) {
		var event domain.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			bus.logger.Error("Failed to unmarshal event",
				zap.Error(err),
				zap.String("subject", msg.Subject))
			return
		}

		ctx := context.Background()
		if err := handler.Handle(ctx, &event); err != nil {
			bus.logger.Error("Failed to handle event",
				zap.Error(err),
				zap.String("event_id", event.ID.String()),
				zap.String("event_type", event.Type))
		}
	})

	if err != nil {
		return nil, fmt.Errorf("subscribing to events: %w", err)
	}

	bus.logger.Info("Subscribed to events", zap.String("subject", subject))
	return sub, nil
}

// SubscribeQueue subscribes to events with queue group
func (bus *NATSEventBus) SubscribeQueue(eventType, queue string, handler EventHandler) (*nats.Subscription, error) {
	subject := fmt.Sprintf("events.%s", eventType)

	sub, err := bus.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var event domain.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			bus.logger.Error("Failed to unmarshal event",
				zap.Error(err),
				zap.String("subject", msg.Subject))
			return
		}

		ctx := context.Background()
		if err := handler.Handle(ctx, &event); err != nil {
			bus.logger.Error("Failed to handle event",
				zap.Error(err),
				zap.String("event_id", event.ID.String()),
				zap.String("event_type", event.Type))
		}
	})

	if err != nil {
		return nil, fmt.Errorf("subscribing to events with queue: %w", err)
	}

	bus.logger.Info("Subscribed to events with queue",
		zap.String("subject", subject),
		zap.String("queue", queue))
	return sub, nil
}

// Close closes the NATS connection
func (bus *NATSEventBus) Close() error {
	if bus.conn != nil && !bus.conn.IsClosed() {
		bus.conn.Close()
	}
	return nil
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event *domain.Event) error
}

// EventHandlerFunc is an adapter to allow using regular functions as EventHandler
type EventHandlerFunc func(ctx context.Context, event *domain.Event) error

// Handle implements EventHandler interface
func (f EventHandlerFunc) Handle(ctx context.Context, event *domain.Event) error {
	return f(ctx, event)
}

// TaskEventHandler handles task-related events
type TaskEventHandler struct {
	logger *zap.Logger
}

// NewTaskEventHandler creates a new task event handler
func NewTaskEventHandler(logger *zap.Logger) *TaskEventHandler {
	return &TaskEventHandler{logger: logger}
}

// Handle handles task events
func (h *TaskEventHandler) Handle(ctx context.Context, event *domain.Event) error {
	switch event.Type {
	case "task.created":
		return h.handleTaskCreated(ctx, event)
	case "task.updated":
		return h.handleTaskUpdated(ctx, event)
	case "task.completed":
		return h.handleTaskCompleted(ctx, event)
	case "task.deleted":
		return h.handleTaskDeleted(ctx, event)
	default:
		h.logger.Warn("Unknown task event type", zap.String("event_type", event.Type))
		return nil
	}
}

func (h *TaskEventHandler) handleTaskCreated(_ context.Context, event *domain.Event) error {
	h.logger.Info("Task created event handled",
		zap.String("event_id", event.ID.String()),
		zap.String("aggregate_id", event.AggregateID.String()))

	// Implement business logic here (notifications, analytics, etc.)
	return nil
}

func (h *TaskEventHandler) handleTaskUpdated(_ context.Context, event *domain.Event) error {
	h.logger.Info("Task updated event handled",
		zap.String("event_id", event.ID.String()),
		zap.String("aggregate_id", event.AggregateID.String()))

	// Implement business logic here
	return nil
}

func (h *TaskEventHandler) handleTaskCompleted(_ context.Context, event *domain.Event) error {
	h.logger.Info("Task completed event handled",
		zap.String("event_id", event.ID.String()),
		zap.String("aggregate_id", event.AggregateID.String()))

	// Implement business logic here (notifications, metrics, etc.)
	return nil
}

func (h *TaskEventHandler) handleTaskDeleted(_ context.Context, event *domain.Event) error {
	h.logger.Info("Task deleted event handled",
		zap.String("event_id", event.ID.String()),
		zap.String("aggregate_id", event.AggregateID.String()))

	// Implement business logic here
	return nil
}
