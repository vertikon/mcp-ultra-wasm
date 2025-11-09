// seed-examples/waba/internal/plugins/waba/plugin.go
package waba

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
)

func init() {
	_ = registry.Register("waba", &WABAPlugin{})
}

type WABAPlugin struct{}

func (p *WABAPlugin) Name() string {
	return "waba"
}

func (p *WABAPlugin) Version() string {
	return "1.0.0"
}

func (p *WABAPlugin) Routes() []contracts.Route {
	return []contracts.Route{
		{Method: "GET", Path: "/waba/webhook", Handler: p.verifyWebhook},
		{Method: "POST", Path: "/waba/webhook", Handler: p.handleWebhook},
		{Method: "POST", Path: "/waba/send", Handler: p.handleSend},
		{Method: "GET", Path: "/waba/templates", Handler: p.handleTemplates},
	}
}

// GET /waba/webhook?hub.mode=subscribe&hub.verify_token=XYZ&hub.challenge=123
func (p *WABAPlugin) verifyWebhook(w http.ResponseWriter, r *http.Request) {
	verify := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")
	expected := os.Getenv("WABA_VERIFY_TOKEN")

	if verify != "" && challenge != "" && verify == expected {
		w.WriteHeader(httpx.StatusOK)
		_, _ = w.Write([]byte(challenge))
		return
	}

	http.Error(w, "verification failed", httpx.StatusForbidden)
}

// POST webhook com X-Hub-Signature-256
func (p *WABAPlugin) handleWebhook(w http.ResponseWriter, r *http.Request) {
	sig := r.Header.Get("X-Hub-Signature-256")
	secret := os.Getenv("WABA_APP_SECRET")

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		if _, err := mac.Write(body); err != nil {
			log.Printf("Error writing to HMAC: %v", err)
			http.Error(w, "signature validation error", httpx.StatusInternalServerError)
			return
		}
		exp := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if sig != exp {
			http.Error(w, "invalid signature", httpx.StatusForbidden)
			return
		}
	}

	// TODO: parse eventos, enfileirar, etc.
	w.WriteHeader(httpx.StatusOK)
}

func (p *WABAPlugin) handleSend(w http.ResponseWriter, r *http.Request) {
	var req struct {
		To       string   `json:"to"`
		Template string   `json:"template"`
		Params   []string `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), httpx.StatusBadRequest)
		return
	}

	// TODO: chamar adapter Meta Graph
	if err := json.NewEncoder(w).Encode(map[string]any{
		"message_id": "wamid.mock",
		"status":     "queued",
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "encoding error", httpx.StatusInternalServerError)
		return
	}
}

func (p *WABAPlugin) handleTemplates(w http.ResponseWriter, r *http.Request) {
	// TODO: listar via Graph API
	if err := json.NewEncoder(w).Encode([]map[string]string{
		{"id": "1", "name": "welcome", "language": "pt_BR"},
		{"id": "2", "name": "order_confirmation", "language": "pt_BR"},
	}); err != nil {
		log.Printf("Error encoding templates response: %v", err)
		http.Error(w, "encoding error", httpx.StatusInternalServerError)
	}
}
