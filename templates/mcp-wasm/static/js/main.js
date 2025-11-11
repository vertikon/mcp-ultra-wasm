// MCP Ultra WASM Web Interface - Main JavaScript

class MCPWebWasmApp {
    constructor() {
        this.wasmModule = null;
        this.wasmLoaded = false;
        this.websocket = null;
        this.currentTaskId = null;
        this.logs = [];
        
        this.init();
    }

    async init() {
        try {
            await this.loadWASM();
            this.setupEventListeners();
            this.connectWebSocket();
            this.addLog('info', 'MCP Ultra WASM Web Interface initialized');
        } catch (error) {
            this.addLog('error', `Failed to initialize: ${error.message}`);
        }
    }

    async loadWASM() {
        try {
            this.addLog('info', 'Loading WASM module...');
            
            // O wasm-loader.js já deve ter carregado o WASM
            if (typeof wasmModule !== 'undefined') {
                this.wasmModule = wasmModule;
                this.wasmLoaded = true;
                
                // Inicializar o módulo WASM
                const initResult = this.wasmModule.initialize(JSON.stringify({
                    version: '1.0.0',
                    debug: true
                }));
                
                if (initResult.error) {
                    throw new Error(initResult.error);
                }
                
                this.addLog('success', 'WASM module loaded and initialized successfully');
            } else {
                throw new Error('WASM module not available');
            }
        } catch (error) {
            this.addLog('error', `Failed to load WASM: ${error.message}`);
            throw error;
        }
    }

    setupEventListeners() {
        // Botão de executar
        const executeBtn = document.getElementById('execute-btn');
        if (executeBtn) {
            executeBtn.addEventListener('click', () => this.executeTask());
        }

        // TextArea para detectar Ctrl+Enter
        const taskInput = document.getElementById('task-input');
        if (taskInput) {
            taskInput.addEventListener('keydown', (e) => {
                if (e.ctrlKey && e.key === 'Enter') {
                    e.preventDefault();
                    this.executeTask();
                }
            });
        }

        // Seleção de função WASM
        const wasmFunction = document.getElementById('wasm-function');
        if (wasmFunction) {
            wasmFunction.addEventListener('change', (e) => {
                this.updateTaskPlaceholder(e.target.value);
            });
        }
    }

    connectWebSocket() {
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                this.addLog('success', 'WebSocket connected');
                this.updateConnectionStatus('online');
            };
            
            this.websocket.onmessage = (event) => {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            };
            
            this.websocket.onclose = () => {
                this.addLog('warning', 'WebSocket disconnected');
                this.updateConnectionStatus('offline');
                
                // Tentar reconectar após 5 segundos
                setTimeout(() => this.connectWebSocket(), 5000);
            };
            
            this.websocket.onerror = (error) => {
                this.addLog('error', `WebSocket error: ${error.message}`);
                this.updateConnectionStatus('offline');
            };
        } catch (error) {
            this.addLog('error', `Failed to connect WebSocket: ${error.message}`);
        }
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'task_progress':
                this.handleTaskProgress(message.data);
                break;
            case 'task_completed':
                this.handleTaskCompleted(message.data);
                break;
            case 'task_error':
                this.handleTaskError(message.data);
                break;
            case 'system_status':
                this.handleSystemStatus(message.data);
                break;
            default:
                this.addLog('info', `Received message type: ${message.type}`);
        }
    }

    async executeTask() {
        if (!this.wasmLoaded) {
            this.addLog('error', 'WASM module not loaded');
            return;
        }

        const taskInput = document.getElementById('task-input');
        const wasmFunction = document.getElementById('wasm-function');
        const executeBtn = document.getElementById('execute-btn');
        
        const taskData = taskInput.value.trim();
        const functionName = wasmFunction.value;
        
        if (!taskData) {
            this.addLog('warning', 'Please enter task data');
            return;
        }

        try {
            // Desabilitar botão
            executeBtn.disabled = true;
            executeBtn.textContent = 'Executing...';
            
            // Limpar output anterior
            this.clearOutput();
            this.updateStatus('running');
            
            // Preparar configuração para a função WASM
            const config = {
                type: functionName,
                data: taskData,
                timestamp: new Date().toISOString()
            };

            this.addLog('info', `Executing ${functionName}...`);
            
            // Gerar correlation ID para rastreamento
            const correlationId = this.generateCorrelationId();
            this.currentTaskId = correlationId;
            
            // Inscrever para atualizações via WebSocket
            if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
                this.websocket.send(JSON.stringify({
                    type: 'subscribe_task',
                    data: { correlation_id: correlationId }
                }));
            }
            
            // Executar função WASM
            await this.executeWasmFunction(functionName, config, correlationId);
            
        } catch (error) {
            this.addLog('error', `Execution failed: ${error.message}`);
            this.updateStatus('error');
        } finally {
            // Reabilitar botão
            executeBtn.disabled = false;
            executeBtn.textContent = 'Execute Task';
        }
    }

    async executeWasmFunction(functionName, config, correlationId) {
        return new Promise((resolve, reject) => {
            // Callback para processar resultado
            const callback = (result) => {
                if (result.error) {
                    this.addLog('error', `WASM Error: ${result.error}`);
                    this.updateStatus('error');
                    reject(new Error(result.error));
                } else {
                    this.handleWasmResult(result, correlationId);
                    resolve(result);
                }
            };

            // Executar função baseada no tipo
            switch (functionName) {
                case 'analyze':
                    this.wasmModule.analyzeProject(JSON.stringify(config), callback);
                    break;
                case 'generate':
                    this.wasmModule.generateCode(JSON.stringify(config), callback);
                    break;
                case 'validate':
                    this.wasmModule.validateConfig(JSON.stringify(config), callback);
                    break;
                default:
                    this.wasmModule.processTask(JSON.stringify(config), callback);
                    break;
            }
        });
    }

    handleWasmResult(result, correlationId) {
        this.addLog('success', `Task completed: ${result.status}`);
        
        // Exibir resultado no output
        const outputContent = document.getElementById('output-content');
        if (outputContent) {
            outputContent.textContent = JSON.stringify(result, null, 2);
            outputContent.classList.add('fade-in');
        }
        
        this.updateStatus('completed');
        
        // Atualizar progresso para 100%
        this.updateProgress(100);
    }

    handleTaskProgress(data) {
        if (data.correlation_id === this.currentTaskId) {
            this.addLog('info', `Progress: ${data.progress}% - ${data.message}`);
            this.updateProgress(data.progress);
        }
    }

    handleTaskCompleted(data) {
        if (data.correlation_id === this.currentTaskId) {
            this.addLog('success', 'Task completed via WebSocket');
            this.updateStatus('completed');
            this.updateProgress(100);
            
            // Exibir resultado
            const outputContent = document.getElementById('output-content');
            if (outputContent) {
                outputContent.textContent = JSON.stringify(data.result, null, 2);
            }
        }
    }

    handleTaskError(data) {
        if (data.correlation_id === this.currentTaskId) {
            this.addLog('error', `Task failed: ${data.error}`);
            this.updateStatus('error');
        }
    }

    handleSystemStatus(data) {
        this.addLog('info', `System status: ${JSON.stringify(data)}`);
    }

    updateStatus(status) {
        const statusIndicator = document.getElementById('status-indicator');
        if (statusIndicator) {
            statusIndicator.className = `status-${status}`;
            statusIndicator.textContent = status.charAt(0).toUpperCase() + status.slice(1);
        }
    }

    updateProgress(percentage) {
        const progressFill = document.querySelector('.progress-fill');
        if (progressFill) {
            progressFill.style.width = `${percentage}%`;
            progressFill.textContent = `${percentage}%`;
        }
    }

    updateConnectionStatus(status) {
        const indicator = document.querySelector('.status-indicator');
        if (indicator) {
            indicator.className = `status-indicator ${status}`;
        }
    }

    updateTaskPlaceholder(functionName) {
        const taskInput = document.getElementById('task-input');
        const placeholders = {
            'analyze': 'Enter project path or configuration to analyze...',
            'generate': 'Enter code generation specifications...',
            'validate': 'Enter configuration to validate...'
        };
        
        if (taskInput) {
            taskInput.placeholder = placeholders[functionName] || 'Enter task data...';
        }
    }

    clearOutput() {
        const outputContent = document.getElementById('output-content');
        if (outputContent) {
            outputContent.textContent = '';
        }
        
        const taskId = document.getElementById('task-id');
        if (taskId) {
            taskId.textContent = '';
        }
        
        this.updateProgress(0);
    }

    addLog(type, message) {
        const timestamp = new Date().toLocaleTimeString();
        const logEntry = {
            timestamp,
            type,
            message
        };
        
        this.logs.push(logEntry);
        
        // Manter apenas os últimos 100 logs
        if (this.logs.length > 100) {
            this.logs = this.logs.slice(-100);
        }
        
        this.renderLogs();
    }

    renderLogs() {
        const logContent = document.getElementById('log-content');
        if (!logContent) return;
        
        const html = this.logs.map(log => `
            <div class="log-entry log-${log.type}">
                <span class="log-time">[${log.timestamp}]</span> ${log.message}
            </div>
        `).join('');
        
        logContent.innerHTML = html;
        logContent.scrollTop = logContent.scrollHeight;
    }

    generateCorrelationId() {
        return `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    // Cleanup quando a página for fechada
    cleanup() {
        if (this.websocket) {
            this.websocket.close();
        }
        
        if (this.wasmModule && this.wasmLoaded) {
            try {
                this.wasmModule.cleanup();
            } catch (error) {
                console.error('Error cleaning up WASM module:', error);
            }
        }
    }
}

// Inicializar aplicação quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', () => {
    window.app = new MCPWebWasmApp();
    
    // Cleanup quando a página for fechada
    window.addEventListener('beforeunload', () => {
        if (window.app) {
            window.app.cleanup();
        }
    });
});

// Excluir funções para depuração
window.debugApp = () => {
    console.log('App state:', window.app);
    console.log('Logs:', window.app?.logs);
    console.log('WASM loaded:', window.app?.wasmLoaded);
    console.log('WebSocket status:', window.app?.websocket?.readyState);
};

window.clearLogs = () => {
    if (window.app) {
        window.app.logs = [];
        window.app.renderLogs();
    }
};