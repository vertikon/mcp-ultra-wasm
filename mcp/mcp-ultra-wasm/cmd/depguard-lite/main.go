package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/analyzers/depguardlite"
)

func main() {
	singlechecker.Main(depguardlite.Analyzer)
}
