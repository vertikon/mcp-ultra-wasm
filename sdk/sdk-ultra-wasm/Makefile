# Makefile para sdk-ultra-wasm

.PHONY: help build test lint clean cli example-waba

# Variáveis
GO := E:\go1.25.0\go\bin\go.exe
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod

help: ## Mostra esta mensagem de ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila todos os pacotes
	$(GOBUILD) ./...

test: ## Executa todos os testes
	$(GOTEST) -v -count=1 ./pkg/...

test-coverage: ## Executa testes com cobertura
	$(GOTEST) -v -coverprofile=coverage.out ./pkg/...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Executa linter (se instalado)
	golangci-lint run || echo "golangci-lint não instalado"

clean: ## Limpa arquivos gerados
	rm -rf bin/ dist/ coverage.out coverage.html
	$(GO) clean -cache -testcache

cli: ## Compila a CLI de scaffolding
	cd cmd/ultra-sdk-cli && $(GOBUILD) -o ../../bin/ultra-sdk-cli.exe .
	@echo "✅ CLI compilada em bin/ultra-sdk-cli.exe"

example-waba: ## Executa o exemplo WABA
	cd seed-examples/waba && $(GO) run ./cmd/main.go

deps: ## Baixa dependências
	$(GOMOD) download
	$(GOMOD) tidy

verify: ## Verifica integridade dos módulos
	$(GOMOD) verify
