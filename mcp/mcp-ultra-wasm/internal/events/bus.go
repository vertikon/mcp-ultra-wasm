package events

import "context"

// EventBus is a placeholder interface for event bus
// Replace with actual implementation (e.g., NATS, Kafka, etc.)
type EventBus interface {
	Publish(ctx context.Context, topic string, data interface{}) error
	Subscribe(ctx context.Context, topic string, handler func(interface{}) error) error
	Close() error
}
