// MCP Ultra WASM Loader

(function() {
    'use strict';

    class WasmLoader {
        constructor() {
            this.wasmModule = null;
            this.isReady = false;
            this.loadPromise = null;
        }

        async load() {
            if (this.loadPromise) {
                return this.loadPromise;
            }

            this.loadPromise = this._loadWasm();
            return this.loadPromise;
        }

        async _loadWasm() {
            try {
                console.log('Loading WebAssembly module...');
                
                // Verificar suporte a WebAssembly
                if (!this._isWebAssemblySupported()) {
                    throw new Error('WebAssembly is not supported in this browser');
                }

                // Carregar wasm_exec.js do Go
                await this._loadGoWasmExec();

                // Carregar o arquivo .wasm
                const wasmBytes = await this._fetchWasmFile();
                
                // Inicializar o módulo Go
                const go = new Go();
                
                // Configurar ambiente Go
                this._setupGoEnvironment(go);

                // Compilar e instanciar o WASM
                const result = await WebAssembly.instantiate(wasmBytes, go.importObject);
                
                // Executar o módulo Go
                go.run(result.instance);

                // Aguardar o módulo estar pronto
                await this._waitForWasmReady();

                this.isReady = true;
                console.log('WebAssembly module loaded successfully');
                
                // Disparar evento de ready
                window.dispatchEvent(new CustomEvent('wasm-ready', {
                    detail: { module: this.wasmModule }
                }));

                return this.wasmModule;

            } catch (error) {
                console.error('Failed to load WebAssembly module:', error);
                throw error;
            }
        }

        _isWebAssemblySupported() {
            return typeof WebAssembly === 'object' && 
                   typeof WebAssembly.instantiate === 'function';
        }

        async _loadGoWasmExec() {
            return new Promise((resolve, reject) => {
                const script = document.createElement('script');
                script.src = '/wasm/wasm_exec.js';
                script.onload = resolve;
                script.onerror = () => reject(new Error('Failed to load wasm_exec.js'));
                document.head.appendChild(script);
            });
        }

        async _fetchWasmFile() {
            const response = await fetch('/wasm/main.wasm', {
                headers: {
                    'Content-Type': 'application/wasm'
                }
            });

            if (!response.ok) {
                throw new Error(`Failed to fetch WASM file: ${response.status}`);
            }

            return response.arrayBuffer();
        }

        _setupGoEnvironment(go) {
            // Configurar environment variables se necessário
            go.env = Object.assign({
                'GOOS': 'js',
                'GOARCH': 'wasm',
            }, go.env);

            // Configurar argumentos se necessário
            go.args = [];

            // Configurar saída para debugging
            if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
                // Capturar stdout/stderr para debugging
                const originalLog = console.log;
                console.log = (...args) => {
                    originalLog.apply(console, args);
                    // Enviar logs para UI se necessário
                    if (window.app && window.app.addLog) {
                        window.app.addLog('info', args.join(' '));
                    }
                };
            }
        }

        async _waitForWasmReady() {
            return new Promise((resolve) => {
                const checkReady = () => {
                    // Verificar se as funções do módulo estão disponíveis
                    if (typeof window.analyzeProject === 'function' &&
                        typeof window.generateCode === 'function' &&
                        typeof window.validateConfig === 'function' &&
                        typeof window.processTask === 'function' &&
                        typeof window.getVersion === 'function' &&
                        typeof window.initialize === 'function' &&
                        typeof window.cleanup === 'function') {
                        
                        // Criar wrapper para o módulo
                        this.wasmModule = {
                            analyzeProject: window.analyzeProject,
                            generateCode: window.generateCode,
                            validateConfig: window.validateConfig,
                            processTask: window.processTask,
                            getVersion: window.getVersion,
                            initialize: window.initialize,
                            cleanup: window.cleanup
                        };

                        // Expor globalmente para compatibilidade
                        window.wasmModule = this.wasmModule;
                        
                        resolve();
                    } else {
                        setTimeout(checkReady, 100);
                    }
                };
                
                checkReady();
            });
        }

        getVersion() {
            if (!this.isReady) {
                throw new Error('WASM module is not loaded');
            }
            
            return this.wasmModule.getVersion();
        }

        initialize(config) {
            if (!this.isReady) {
                throw new Error('WASM module is not loaded');
            }
            
            return this.wasmModule.initialize(config);
        }

        cleanup() {
            if (!this.isReady) {
                return;
            }
            
            try {
                this.wasmModule.cleanup();
            } catch (error) {
                console.error('Error during WASM cleanup:', error);
            }
            
            this.wasmModule = null;
            this.isReady = false;
        }
    }

    // Criar e expor o loader globalmente
    window.wasmLoader = new WasmLoader();
    
    // Função conveniente para carregar o WASM
    window.loadWasm = () => window.wasmLoader.load();
    
    // Auto-carregar quando o script for executado
    console.log('WASM Loader initialized, auto-loading module...');
    
    // Iniciar carregamento assíncrono
    window.wasmLoader.load().catch(error => {
        console.error('Auto-loading WASM failed:', error);
        
        // Exibir erro na UI se disponível
        const outputContent = document.getElementById('output-content');
        if (outputContent) {
            outputContent.textContent = `Error loading WASM module: ${error.message}`;
            outputContent.style.color = '#e53e3e';
        }
    });

})();