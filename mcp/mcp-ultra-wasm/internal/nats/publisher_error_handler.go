package natsx

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

// Publisher publishes messages to NATS with retry and error handling
type Publisher struct {
	js         nats.JetStreamContext
	subjectErr string
	maxRetries int
	backoff    time.Duration
}

// NewPublisher creates a new NATS publisher with error handling
func NewPublisher(js nats.JetStreamContext, subjectErr string) *Publisher {
	return &Publisher{
		js:         js,
		subjectErr: subjectErr,
		maxRetries: 3,
		backoff:    250 * time.Millisecond,
	}
}

// PublishWithRetry publishes a message with retry logic and error reporting
func (p *Publisher) PublishWithRetry(ctx context.Context, subject string, payload []byte) error {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		_, err := p.js.Publish(subject, payload)
		if err == nil {
			return nil
		}

		lastErr = err
		slog.Error("nats publish failed",
			"subject", subject,
			"attempt", attempt,
			"err", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt+1) * p.backoff):
		}
	}

	// Best-effort: publish error event
	if p.subjectErr != "" {
		ev := []byte(`{"timestamp":"` + time.Now().UTC().Format(time.RFC3339Nano) +
			`","subject":"` + subject +
			`","error":"` + sanitizeErr(lastErr) + `"}`)
		_, _ = p.js.Publish(p.subjectErr, ev)
	}

	return lastErr
}

// sanitizeErr prevents leaking credentials in logs
func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline exceeded"
	}
	return "publish failed"
}
