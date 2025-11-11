package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Configurar handlers para servir arquivos estáticos
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/wasm/", http.StripPrefix("/wasm/", http.FileServer(http.Dir("wasm/"))))
	
	// Handler para WebSocket
	http.HandleFunc("/ws", handleWebSocket)
	
	// Handler principal para servir o template
	http.HandleFunc("/", serveTemplate)

	// Handler de health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\": \"ok\", \"service\": \"mcp-ultra-wasm\"}")
	})

	// Iniciar servidor
	log.Printf("MCP Ultra WASM Server starting on port %s", port)
	log.Printf("Static files: ./static/")
	log.Printf("WASM files: ./wasm/")
	log.Printf("Open: http://localhost:%s", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	// Se for uma requisição para a raiz, servir o index.html
	if r.URL.Path == "/" {
		templatePath := filepath.Join("templates", "index.html")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			// Se o template não existir, criar um HTML básico
			serveBasicHTML(w)
			return
		}
		
		http.ServeFile(w, r, templatePath)
		return
	}
	
	// Para outras rotas, tentar servir arquivo estático
	http.FileServer(http.Dir(".")).ServeHTTP(w, r)
}

func serveBasicHTML(w http.ResponseWriter) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>MCP Ultra WASM</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .status.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .status.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .status.info { background: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
        textarea { width: 100%; height: 150px; font-family: monospace; margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        button { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; margin: 5px; }
        button:hover { background: #0056b3; }
        button:disabled { background: #6c757d; cursor: not-allowed; }
        .output { background: #1e1e1e; color: #f8f9fa; padding: 15px; border-radius: 4px; font-family: monospace; min-height: 200px; white-space: pre-wrap; overflow-x: auto; }
        .logs { background: #000; color: #00ff00; padding: 15px; border-radius: 4px; font-family: monospace; height: 200px; overflow-y: auto; margin-top: 20px; }
        .function-select { margin: 10px 0; padding: 10px; border: 1px solid #ddd; border-radius: 4px; width: 100%; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>MCP Ultra WASM Interface</h1>
            <p>Execute Go code directly in your browser with WebAssembly</p>
            <div class="status info" id="connection-status">
                Connecting to WebAssembly...
            </div>
        </div>

        <div class="controls">
            <h3>Control Panel</h3>
            
            <div class="function-select-container">
                <label for="wasm-function">Select Function:</label><br>
                <select id="wasm-function" class="function-select">
                    <option value="analyze">Analyze Project</option>
                    <option value="generate">Generate Code</option>
                    <option value="validate">Validate Config</option>
                    <option value="process">Process Task</option>
                </select>
            </div>

            <div class="task-input">
                <label for="task-input">Task Configuration (JSON):</label><br>
                <textarea id="task-input" placeholder='{"project_path": "/path/to/project", "analysis_type": "full"}'></textarea>
            </div>

            <div class="buttons">
                <button id="execute-btn" onclick="executeTask()">
                    Execute Task
                </button>
                <button onclick="clearOutput()">
                    Clear Output
                </button>
                <button onclick="clearLogs()">
                    Clear Logs
                </button>
                <button onclick="testConnection()">
                    Test Connection
                </button>
            </div>
        </div>

        <div class="output-section">
            <h3>Output</h3>
            <div class="output" id="output-content">
                Waiting for execution...
            </div>
        </div>

        <div class="logs-section">
            <h3>System Logs</h3>
            <div class="logs" id="log-content">
                [00:00:00] System initialized...<br>
                [00:00:01] Loading WebAssembly module...<br>
            </div>
        </div>
    </div>

    <script src="/static/js/wasm-loader.js"></script>
    <script>
        let wasmModule = null;
        let logs = [];
        
        function addLog(type, message) {
            const timestamp = new Date().toLocaleTimeString();
            const logEntry = "[${timestamp}] ${message}";
            logs.push({timestamp, type, message});
            
            const logContent = document.getElementById("log-content");
            logContent.innerHTML += logEntry + "<br>";
            logContent.scrollTop = logContent.scrollHeight;
            
            console.log("[" + type.toUpperCase() + "] " + message);
        }
        
        function updateOutput(content) {
            document.getElementById("output-content").textContent = content;
        }
        
        function updateConnectionStatus(status, message) {
            const statusEl = document.getElementById("connection-status");
            statusEl.className = "status " + status;
            statusEl.textContent = message;
        }
        
        function executeTask() {
            if (!wasmModule) {
                addLog("error", "WASM module not loaded");
                return;
            }
            
            const functionSelect = document.getElementById("wasm-function");
            const taskInput = document.getElementById("task-input");
            const executeBtn = document.getElementById("execute-btn");
            
            const functionName = functionSelect.value;
            const taskData = taskInput.value || "{}";
            
            executeBtn.disabled = true;
            executeBtn.textContent = "Executing...";
            
            addLog("info", "Executing " + functionName + " with data: " + taskData);
            
            try {
                const config = {
                    type: functionName,
                    data: JSON.parse(taskData),
                    timestamp: new Date().toISOString()
                };
                
                const callback = function(result) {
                    executeBtn.disabled = false;
                    executeBtn.textContent = "Execute Task";
                    
                    if (result.error) {
                        addLog("error", "WASM Error: " + result.error);
                        updateOutput("Error: " + result.error);
                    } else {
                        addLog("success", "Task completed successfully");
                        updateOutput(JSON.stringify(result, null, 2));
                    }
                };
                
                // Execute function based on selection
                switch (functionName) {
                    case "analyze":
                        wasmModule.analyzeProject(JSON.stringify(config), callback);
                        break;
                    case "generate":
                        wasmModule.generateCode(JSON.stringify(config), callback);
                        break;
                    case "validate":
                        wasmModule.validateConfig(JSON.stringify(config), callback);
                        break;
                    default:
                        wasmModule.processTask(JSON.stringify(config), callback);
                        break;
                }
                
            } catch (error) {
                executeBtn.disabled = false;
                executeBtn.textContent = "Execute Task";
                addLog("error", "JSON Parse Error: " + error.message);
                updateOutput("JSON Error: " + error.message);
            }
        }
        
        function clearOutput() {
            document.getElementById("output-content").textContent = "Output cleared...";
        }
        
        function clearLogs() {
            document.getElementById("log-content").innerHTML = "[00:00:00] Logs cleared...<br>";
            logs = [];
        }
        
        function testConnection() {
            if (wasmModule) {
                try {
                    const version = wasmModule.getVersion();
                    addLog("success", "WASM Connected! Version: " + version.version);
                    updateConnectionStatus("success", "WASM Connected");
                } catch (error) {
                    addLog("error", "WASM Test Failed: " + error.message);
                    updateConnectionStatus("error", "WASM Error");
                }
            } else {
                addLog("warning", "WASM module not loaded yet");
                updateConnectionStatus("info", "Loading WASM...");
            }
        }
        
        // Monitor WASM loading
        window.addEventListener("wasm-ready", function(event) {
            wasmModule = event.detail.module;
            addLog("success", "WebAssembly module loaded successfully!");
            updateConnectionStatus("success", "WASM Ready");
            
            // Auto-test
            setTimeout(testConnection, 1000);
        });
        
        // Update task placeholder based on function selection
        document.getElementById("wasm-function").addEventListener("change", function() {
            const taskInput = document.getElementById("task-input");
            const placeholders = {
                "analyze": '{"project_path": "/path/to/go/project", "analysis_type": "full"}',
                "generate": '{"component_type": "service", "name": "UserService", "language": "go"}',
                "validate": '{"type": "project", "project_name": "my-app", "module_path": "github.com/user/my-app"}',
                "process": '{"type": "custom", "action": "process", "parameters": {}}'
            };
            
            taskInput.placeholder = placeholders[this.value] || "{}";
        });
        
        // Keyboard shortcuts
        document.getElementById("task-input").addEventListener("keydown", function(e) {
            if (e.ctrlKey && e.key === "Enter") {
                e.preventDefault();
                executeTask();
            }
        });
        
        // Initialize
        addLog("info", "Basic HTML interface loaded");
        updateConnectionStatus("info", "Loading WebAssembly...");
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Simple handler for WebSocket (placeholder)
	log.Printf("WebSocket connection from %s", r.RemoteAddr)
	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "{\"error\": \"WebSocket not implemented yet\"}")
}