// cmd/ultra-sdk-cli/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const defaultDirPermissions = 0755

const pluginTemplate = `package {{.Name}}

import (
	"encoding/json"
	"net/http"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
)

func init() {
	_ = registry.Register("{{.Name}}", &Plugin{})
}

type Plugin struct{}

func (p *Plugin) Name() string {
	return "{{.Name}}"
}

func (p *Plugin) Version() string {
	return "0.1.0"
}

func (p *Plugin) Routes() []contracts.Route {
	return []contracts.Route{
		{
			Method:  "GET",
			Path:    "/{{.Name}}/status",
			Handler: p.handleStatus,
		},
	}
}

func (p *Plugin) handleStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "ok",
		"service": "{{.Name}}",
		"version": p.Version(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
`

func main() {
	name := flag.String("name", "", "Nome do plugin")
	kind := flag.String("kind", "generic", "Tipo: omnichannel|marketing|ia|generic")
	output := flag.String("output", "internal/plugins", "Diretório de saída")
	flag.Parse()

	if *name == "" {
		fmt.Println("Uso: ultra-sdk-cli --name my-plugin [--kind omnichannel] [--output internal/plugins]")
		os.Exit(1)
	}

	// Criar diretório
	dir := filepath.Join(*output, *name)
	if err := os.MkdirAll(dir, defaultDirPermissions); err != nil {
		fmt.Printf("Erro ao criar diretório: %v\n", err)
		os.Exit(1)
	}

	// Gerar arquivo
	file := filepath.Join(dir, "plugin.go")
	f, err := os.Create(file)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Erro ao fechar arquivo: %v\n", err)
		}
	}()

	tmpl := template.Must(template.New("plugin").Parse(pluginTemplate))
	if err := tmpl.Execute(f, map[string]string{"Name": *name}); err != nil {
		fmt.Printf("Erro ao gerar template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Plugin criado: %s\n", file)
	fmt.Printf("📝 Tipo: %s\n", *kind)
	fmt.Printf("📂 Edite o arquivo e implemente sua lógica\n")
}
