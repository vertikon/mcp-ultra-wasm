package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketHandler struct {
	logger   *zap.Logger
	upgrader websocket.Upgrader
	clients  map[string]*WebSocketClient
	mutex    sync.RWMutex
	done     chan struct{}
}

type WebSocketClient struct {
	ID            string
	Connection    *websocket.Conn
	CorrelationID string
	LastPing      time.Time
	SendChan      chan []byte
	mutex         sync.Mutex
}

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewWebSocketHandler(logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implementar validação de origem
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients: make(map[string]*WebSocketClient),
		done:    make(chan struct{}),
	}
}

// HandleWebSocket gerencia conexões WebSocket
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Upgrade HTTP para WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Erro ao fazer upgrade WebSocket", zap.Error(err))
		return
	}
	defer conn.Close()

	// Criar cliente
	clientID := uuid.New().String()
	client := &WebSocketClient{
		ID:         clientID,
		Connection: conn,
		LastPing:   time.Now(),
		SendChan:   make(chan []byte, 256),
	}

	// Registrar cliente
	h.mutex.Lock()
	h.clients[clientID] = client
	h.mutex.Unlock()

	h.logger.Info("Cliente WebSocket conectado",
		zap.String("client_id", clientID),
		zap.String("remote_addr", c.Request.RemoteAddr))

	// Cleanup ao desconectar
	defer func() {
		h.mutex.Lock()
		delete(h.clients, clientID)
		h.mutex.Unlock()
		close(client.SendChan)
		h.logger.Info("Cliente WebSocket desconectado", zap.String("client_id", clientID))
	}()

	// Goroutine para enviar mensagens
	go h.writeMessages(client)

	// Loop para ler mensagens
	h.readMessages(client)
}

// readMessages lê mensagens do cliente WebSocket
func (h *WebSocketHandler) readMessages(client *WebSocketClient) {
	defer client.Connection.Close()

	client.Connection.SetReadLimit(512)
	client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Connection.SetPongHandler(func(appData string) error {
		client.LastPing = time.Now()
		client.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WebSocketMessage
		err := client.Connection.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("Erro inesperado no WebSocket",
					zap.String("client_id", client.ID),
					zap.Error(err))
			}
			break
		}

		h.logger.Debug("Mensagem recebida do cliente",
			zap.String("client_id", client.ID),
			zap.String("type", msg.Type))

		// Processar mensagem
		h.handleMessage(client, msg)
	}
}

// writeMessages envia mensagens para o cliente WebSocket
func (h *WebSocketHandler) writeMessages(client *WebSocketClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-client.SendChan:
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Connection.WriteMessage(websocket.TextMessage, message); err != nil {
				h.logger.Error("Erro ao enviar mensagem WebSocket",
					zap.String("client_id", client.ID),
					zap.Error(err))
				return
			}

		case <-ticker.C:
			client.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processa mensagens recebidas do cliente
func (h *WebSocketHandler) handleMessage(client *WebSocketClient, msg WebSocketMessage) {
	switch msg.Type {
	case "ping":
		response := WebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		h.sendToClient(client, response)

	case "subscribe_task":
		if correlationID, ok := msg.Data["correlation_id"].(string); ok {
			client.CorrelationID = correlationID
			h.logger.Info("Cliente inscrito na task",
				zap.String("client_id", client.ID),
				zap.String("correlation_id", correlationID))
		}

	case "unsubscribe_task":
		client.CorrelationID = ""
		h.logger.Info("Cliente desinscrito da task", zap.String("client_id", client.ID))

	default:
		h.logger.Warn("Tipo de mensagem desconhecido",
			zap.String("client_id", client.ID),
			zap.String("type", msg.Type))
	}
}

// sendToClient envia mensagem para um cliente específico
func (h *WebSocketHandler) sendToClient(client *WebSocketClient, msg WebSocketMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Erro ao serializar mensagem", zap.Error(err))
		return
	}

	select {
	case client.SendChan <- data:
	default:
		h.logger.Warn("Canal do cliente cheio, descartando mensagem",
			zap.String("client_id", client.ID))
	}
}

// BroadcastToClients envia mensagem para todos os clientes
func (h *WebSocketHandler) BroadcastToClients(msg WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Erro ao serializar broadcast", zap.Error(err))
		return
	}

	for _, client := range h.clients {
		select {
		case client.SendChan <- data:
		default:
			h.logger.Warn("Canal do cliente cheio durante broadcast",
				zap.String("client_id", client.ID))
		}
	}
}

// SendToClientByCorrelationID envia mensagem para cliente por correlation ID
func (h *WebSocketHandler) SendToClientByCorrelationID(correlationID string, msg WebSocketMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error("Erro ao serializar mensagem por correlation ID", zap.Error(err))
		return
	}

	for _, client := range h.clients {
		if client.CorrelationID == correlationID {
			select {
			case client.SendChan <- data:
				h.logger.Debug("Mensagem enviada para cliente por correlation ID",
					zap.String("client_id", client.ID),
					zap.String("correlation_id", correlationID))
			default:
				h.logger.Warn("Canal do cliente cheio",
					zap.String("client_id", client.ID))
			}
		}
	}
}

// GetConnectedClients retorna número de clientes conectados
func (h *WebSocketHandler) GetConnectedClients() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// Close fecha todas as conexões WebSocket
func (h *WebSocketHandler) Close() {
	close(h.done)

	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, client := range h.clients {
		client.Connection.Close()
		close(client.SendChan)
	}
	h.clients = make(map[string]*WebSocketClient)
}
