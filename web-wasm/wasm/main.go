package main

import (
	"syscall/js"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm/web-wasm/wasm/functions"
)

func main() {
	// Registrar funções globais no JavaScript
	js.Global().Set("analyzeProject", js.FuncOf(functions.AnalyzeProject))
	js.Global().Set("generateCode", js.FuncOf(functions.GenerateCode))
	js.Global().Set("validateConfig", js.FuncOf(functions.ValidateConfig))
	js.Global().Set("processTask", js.FuncOf(functions.ProcessTask))
	js.Global().Set("getVersion", js.FuncOf(functions.GetVersion))
	js.Global().Set("initialize", js.FuncOf(functions.Initialize))
	js.Global().Set("cleanup", js.FuncOf(functions.Cleanup))

	// Enviar sinal de que o WASM foi carregado
	js.Global().Get("console").Call("log", "MCP Ultra WASM module loaded successfully")
	
	// Manter o processo WASM ativo
	select {}
}