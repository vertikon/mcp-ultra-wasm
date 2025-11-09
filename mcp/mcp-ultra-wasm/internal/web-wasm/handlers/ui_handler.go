package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UIHandler struct {
	staticPath string
	wasmPath   string
	logger     *zap.Logger
}

func NewUIHandler(staticPath, wasmPath string, logger *zap.Logger) *UIHandler {
	return &UIHandler{
		staticPath: staticPath,
		wasmPath:   wasmPath,
		logger:     logger,
	}
}

// ServeIndex serve a página HTML principal
func (h *UIHandler) ServeIndex(c *gin.Context) {
	indexPath := filepath.Join(h.staticPath, "..", "templates", "index.html")

	// Se o template não existir, serve um HTML básico
	if _, err := filepath.Glob(indexPath); err != nil {
		h.serveBasicHTML(c)
		return
	}

	c.File(indexPath)
}

// serveBasicHTML serve um HTML básico quando não há template
func (h *UIHandler) serveBasicHTML(c *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MCP Ultra WASM Web Interface</title>
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body>
    <div id="app">
        <header>
            <h1>MCP Ultra WASM</h1>
            <p>Interface Web para Componentes WASM</p>
        </header>
        
        <main>
            <section id="controls">
                <div class="control-group">
                    <label for="task-input">Task Input:</label>
                    <textarea id="task-input" placeholder="Digite sua task aqui..."></textarea>
                    <button id="execute-btn">Executar Task</button>
                </div>
                
                <div class="control-group">
                    <label for="wasm-function">Função WASM:</label>
                    <select id="wasm-function">
                        <option value="analyze">Analisar Projeto</option>
                        <option value="generate">Gerar Código</option>
                        <option value="validate">Validar Configuração</option>
                    </select>
                </div>
            </section>
            
            <section id="status">
                <h2>Status</h2>
                <div id="status-indicator" class="status-idle">Idle</div>
                <div id="task-id"></div>
            </section>
            
            <section id="output">
                <h2>Output</h2>
                <pre id="output-content"></pre>
            </section>
            
            <section id="logs">
                <h2>Logs</h2>
                <div id="log-content"></div>
            </section>
        </main>
        
        <footer>
            <p>Powered by MCP Ultra WASM & WebAssembly</p>
        </footer>
    </div>
    
    <script src="/static/js/wasm-loader.js"></script>
    <script src="/static/js/websocket-client.js"></script>
    <script src="/static/js/main.js"></script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// ServeWASM serve arquivos WASM com os headers corretos
func (h *UIHandler) ServeWASM(c *gin.Context) {
	filePath := c.Param("filepath")

	// Remove o prefixo /
	if strings.HasPrefix(filePath, "/") {
		filePath = filePath[1:]
	}

	fullPath := filepath.Join(h.wasmPath, filePath)

	// Verifica se o arquivo existe
	if _, err := filepath.Glob(fullPath); err != nil {
		h.logger.Warn("Arquivo WASM não encontrado",
			zap.String("path", fullPath),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo WASM não encontrado"})
		return
	}

	// Configura headers para WASM
	c.Header("Content-Type", "application/wasm")
	c.Header("Cross-Origin-Opener-Policy", "same-origin")
	c.Header("Cross-Origin-Embedder-Policy", "require-corp")

	h.logger.Debug("Servindo arquivo WASM",
		zap.String("path", fullPath),
		zap.String("content-type", "application/wasm"))

	c.File(fullPath)
}

// ServeStatic serve arquivos estáticos genéricos
func (h *UIHandler) ServeStatic(c *gin.Context) {
	filePath := c.Param("filepath")
	fullPath := filepath.Join(h.staticPath, filePath)

	// Verifica se o arquivo existe
	if _, err := filepath.Glob(fullPath); err != nil {
		h.logger.Warn("Arquivo estático não encontrado",
			zap.String("path", fullPath),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Arquivo não encontrado"})
		return
	}

	// Configura Content-Type baseado na extensão
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".css":
		c.Header("Content-Type", "text/css")
	case ".js":
		c.Header("Content-Type", "application/javascript")
	case ".html":
		c.Header("Content-Type", "text/html; charset=utf-8")
	case ".json":
		c.Header("Content-Type", "application/json")
	case ".png":
		c.Header("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
	case ".svg":
		c.Header("Content-Type", "image/svg+xml")
	}

	c.File(fullPath)
}

// GetWasmFiles retorna lista de arquivos WASM disponíveis
func (h *UIHandler) GetWasmFiles(c *gin.Context) {
	files, err := filepath.Glob(filepath.Join(h.wasmPath, "*.wasm"))
	if err != nil {
		h.logger.Error("Erro ao listar arquivos WASM", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar arquivos WASM"})
		return
	}

	var wasmFiles []string
	for _, file := range files {
		wasmFiles = append(wasmFiles, filepath.Base(file))
	}

	c.JSON(http.StatusOK, gin.H{
		"files": wasmFiles,
		"count": len(wasmFiles),
	})
}
