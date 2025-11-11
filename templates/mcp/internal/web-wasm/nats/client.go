package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Client struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger *zap.Logger
}

func NewClient(url string, logger *zap.Logger) (*Client, error) {
	// Conectar ao NATS
	conn, err := nats.Connect(url,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Error("Desconectado do NATS", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("Reconectado ao NATS", zap.String("server", nc.ConnectedUrl()))
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Error("Erro no NATS", zap.String("subject", sub.Subject), zap.Error(err))
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao NATS: %w", err)
	}

	// Inicializar JetStream
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao inicializar JetStream: %w", err)
	}

	logger.Info("Conectado ao NATS com sucesso", zap.String("url", url))

	return &Client{
		conn:   conn,
		js:     js,
		logger: logger,
	}, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.logger.Info("Conexão NATS fechada")
	}
}

func (c *Client) GetConnection() *nats.Conn {
	return c.conn
}

func (c *Client) GetJetStream() nats.JetStreamContext {
	return c.js
}

// CreateStream cria um stream se não existir
func (c *Client) CreateStream(config nats.StreamConfig) error {
	stream, err := c.js.StreamInfo(config.Name)
	if err == nil && stream != nil {
		c.logger.Debug("Stream já existe", zap.String("name", config.Name))
		return nil
	}

	_, err = c.js.AddStream(&config)
	if err != nil {
		return fmt.Errorf("erro ao criar stream %s: %w", config.Name, err)
	}

	c.logger.Info("Stream criado com sucesso",
		zap.String("name", config.Name),
		zap.Strings("subjects", config.Subjects))

	return nil
}

// Subscribe se inscreve em um subject
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := c.conn.Subscribe(subject, handler)
	if err != nil {
		return nil, fmt.Errorf("erro ao se inscrever em %s: %w", subject, err)
	}

	c.logger.Info("Inscrito no subject", zap.String("subject", subject))
	return sub, nil
}

// SubscribeQueue se inscreve em uma fila
func (c *Client) SubscribeQueue(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := c.conn.QueueSubscribe(subject, queue, handler)
	if err != nil {
		return nil, fmt.Errorf("erro ao se inscrever na fila %s: %w", subject, err)
	}

	c.logger.Info("Inscrito na fila",
		zap.String("subject", subject),
		zap.String("queue", queue))

	return sub, nil
}

// Publish publica uma mensagem
func (c *Client) Publish(subject string, data interface{}) error {
	msgData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	err = c.conn.Publish(subject, msgData)
	if err != nil {
		return fmt.Errorf("erro ao publicar em %s: %w", subject, err)
	}

	c.logger.Debug("Mensagem publicada",
		zap.String("subject", subject),
		zap.Int("size", len(msgData)))

	return nil
}

// PublishAsync publica mensagem assincronamente
func (c *Client) PublishAsync(subject string, data interface{}) error {
	msgData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao serializar mensagem: %w", err)
	}

	err = c.conn.PublishAsync(subject, msgData)
	if err != nil {
		return fmt.Errorf("erro ao publicar assincronamente em %s: %w", subject, err)
	}

	c.logger.Debug("Mensagem publicada assincronamente",
		zap.String("subject", subject),
		zap.Int("size", len(msgData)))

	return nil
}

// Request faz uma requisição com timeout
func (c *Client) Request(subject string, data interface{}, timeout time.Duration) (*nats.Msg, error) {
	msgData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	msg, err := c.conn.Request(subject, msgData, timeout)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição para %s: %w", subject, err)
	}

	return msg, nil
}

// Health verifica saúde da conexão
func (c *Client) Health() error {
	if c.conn == nil || c.conn.IsClosed() {
		return fmt.Errorf("conexão NATS está fechada")
	}

	if c.conn.Status() != nats.CONNECTED {
		return fmt.Errorf("conexão NATS não está conectada, status: %s", c.conn.Status())
	}

	return nil
}

// GetStats retorna estatísticas da conexão
func (c *Client) GetStats() map[string]interface{} {
	stats := c.conn.Stats()
	return map[string]interface{}{
		"connects":          stats.Connects,
		"reconnects":        stats.Reconnects,
		"errors":            stats.Errors,
		"in_msgs":           stats.InMsgs,
		"out_msgs":          stats.OutMsgs,
		"in_bytes":          stats.InBytes,
		"out_bytes":         stats.OutBytes,
		"status":            c.conn.Status().String(),
		"connected_server":  c.conn.ConnectedUrl(),
		"connected_servers": c.conn.Servers(),
	}
}
