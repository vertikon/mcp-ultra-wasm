// MCP Ultra WASM WebSocket Client

(function() {
    'use strict';

    class WebSocketClient {
        constructor(url) {
            this.url = url || this._getWebSocketUrl();
            this.ws = null;
            this.reconnectAttempts = 0;
            this.maxReconnectAttempts = 5;
            this.reconnectDelay = 1000;
            this.isConnecting = false;
            this.isConnected = false;
            this.subscribers = new Map();
            this.messageQueue = [];
            this.heartbeatInterval = null;
            
            // Event callbacks
            this.onOpen = null;
            this.onClose = null;
            this.onMessage = null;
            this.onError = null;
        }

        connect() {
            if (this.isConnecting || this.isConnected) {
                return Promise.resolve();
            }

            return new Promise((resolve, reject) => {
                this.isConnecting = true;
                
                try {
                    console.log(`Connecting to WebSocket: ${this.url}`);
                    this.ws = new WebSocket(this.url);
                    
                    this.ws.onopen = () => {
                        this.isConnected = true;
                        this.isConnecting = false;
                        this.reconnectAttempts = 0;
                        
                        console.log('WebSocket connected');
                        this._startHeartbeat();
                        
                        // Processar fila de mensagens
                        this._processMessageQueue();
                        
                        // Disparar callbacks
                        if (this.onOpen) this.onOpen();
                        
                        // Disparar evento global
                        window.dispatchEvent(new CustomEvent('websocket-open'));
                        
                        resolve();
                    };
                    
                    this.ws.onmessage = (event) => {
                        try {
                            const message = JSON.parse(event.data);
                            this._handleMessage(message);
                        } catch (error) {
                            console.error('Failed to parse WebSocket message:', error);
                        }
                    };
                    
                    this.ws.onclose = (event) => {
                        this.isConnected = false;
                        this.isConnecting = false;
                        this._stopHeartbeat();
                        
                        console.log(`WebSocket closed: ${event.code} - ${event.reason}`);
                        
                        // Disparar callbacks
                        if (this.onClose) this.onClose(event);
                        
                        // Disparar evento global
                        window.dispatchEvent(new CustomEvent('websocket-close', {
                            detail: { code: event.code, reason: event.reason }
                        }));
                        
                        // Tentar reconectar se não foi um fechamento normal
                        if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                            this._scheduleReconnect();
                        }
                    };
                    
                    this.ws.onerror = (error) => {
                        console.error('WebSocket error:', error);
                        this.isConnecting = false;
                        
                        // Disparar callbacks
                        if (this.onError) this.onError(error);
                        
                        // Disparar evento global
                        window.dispatchEvent(new CustomEvent('websocket-error', {
                            detail: error
                        }));
                        
                        reject(error);
                    };
                    
                } catch (error) {
                    this.isConnecting = false;
                    reject(error);
                }
            });
        }

        disconnect() {
            if (this.ws) {
                this._stopHeartbeat();
                this.ws.close(1000, 'Client disconnect');
                this.ws = null;
            }
            this.isConnected = false;
            this.isConnecting = false;
        }

        send(message) {
            if (!this.isConnected) {
                console.warn('WebSocket not connected, queuing message');
                this.messageQueue.push(message);
                return false;
            }

            try {
                const data = typeof message === 'string' ? message : JSON.stringify(message);
                this.ws.send(data);
                return true;
            } catch (error) {
                console.error('Failed to send WebSocket message:', error);
                return false;
            }
        }

        subscribe(eventType, callback) {
            if (!this.subscribers.has(eventType)) {
                this.subscribers.set(eventType, new Set());
            }
            this.subscribers.get(eventType).add(callback);
            
            // Enviar mensagem de inscrição para o servidor
            this.send({
                type: 'subscribe',
                data: { event_type: eventType }
            });
        }

        unsubscribe(eventType, callback) {
            if (this.subscribers.has(eventType)) {
                this.subscribers.get(eventType).delete(callback);
                
                // Enviar mensagem de cancelamento de inscrição
                this.send({
                    type: 'unsubscribe',
                    data: { event_type: eventType }
                });
            }
        }

        // Métodos específicos para MCP Ultra WASM
        subscribeToTask(correlationId, callback) {
            const eventType = `task_${correlationId}`;
            this.subscribe(eventType, callback);
            
            // Enviar mensagem de inscrição específica
            this.send({
                type: 'subscribe_task',
                data: { correlation_id: correlationId }
            });
        }

        unsubscribeFromTask(correlationId, callback) {
            const eventType = `task_${correlationId}`;
            this.unsubscribe(eventType, callback);
        }

        sendPing() {
            this.send({ type: 'ping', timestamp: Date.now() });
        }

        sendPong() {
            this.send({ type: 'pong', timestamp: Date.now() });
        }

        // Métodos privados
        _getWebSocketUrl() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const host = window.location.host;
            return `${protocol}//${host}/ws`;
        }

        _handleMessage(message) {
            // Disparar callbacks genéricos
            if (this.onMessage) this.onMessage(message);
            
            // Disparar evento global
            window.dispatchEvent(new CustomEvent('websocket-message', {
                detail: message
            }));
            
            // Disparar callbacks específicos do tipo
            const eventType = message.type;
            if (this.subscribers.has(eventType)) {
                this.subscribers.get(eventType).forEach(callback => {
                    try {
                        callback(message);
                    } catch (error) {
                        console.error(`Error in WebSocket callback for ${eventType}:`, error);
                    }
                });
            }
            
            // Disparar callbacks específicos de task
            if (message.data && message.data.correlation_id) {
                const taskEventType = `task_${message.data.correlation_id}`;
                if (this.subscribers.has(taskEventType)) {
                    this.subscribers.get(taskEventType).forEach(callback => {
                        try {
                            callback(message);
                        } catch (error) {
                            console.error(`Error in task WebSocket callback:`, error);
                        }
                    });
                }
            }
        }

        _processMessageQueue() {
            while (this.messageQueue.length > 0) {
                const message = this.messageQueue.shift();
                this.send(message);
            }
        }

        _scheduleReconnect() {
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            console.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);
            
            setTimeout(() => {
                if (!this.isConnected) {
                    this.connect().catch(error => {
                        console.error(`Reconnect attempt ${this.reconnectAttempts} failed:`, error);
                    });
                }
            }, delay);
        }

        _startHeartbeat() {
            this.heartbeatInterval = setInterval(() => {
                if (this.isConnected) {
                    this.sendPing();
                }
            }, 30000); // Ping a cada 30 segundos
        }

        _stopHeartbeat() {
            if (this.heartbeatInterval) {
                clearInterval(this.heartbeatInterval);
                this.heartbeatInterval = null;
            }
        }
    }

    // Criar e expor o cliente WebSocket globalmente
    window.wsClient = new WebSocketClient();
    
    // Conectar automaticamente quando a página carregar
    document.addEventListener('DOMContentLoaded', () => {
        window.wsClient.connect().catch(error => {
            console.error('Failed to connect WebSocket:', error);
        });
    });
    
    // Expor funções para facilitar uso
    window.sendWebSocketMessage = (message) => window.wsClient.send(message);
    window.subscribeToWebSocket = (eventType, callback) => window.wsClient.subscribe(eventType, callback);
    window.subscribeToTask = (correlationId, callback) => window.wsClient.subscribeToTask(correlationId, callback);

})();