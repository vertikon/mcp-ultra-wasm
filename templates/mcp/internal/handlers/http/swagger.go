package http

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// SwaggerUIHandler serves the Swagger UI
func SwaggerUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove /docs prefix to get the file path
		uiPath := strings.TrimPrefix(r.URL.Path, "/docs")

		// Default to index.html if no path specified
		if uiPath == "" || uiPath == "/" {
			uiPath = "/index.html"
		}

		// Security: prevent directory traversal
		cleanPath := filepath.Clean(uiPath)
		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		switch cleanPath {
		case "/index.html", "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(swaggerUIHTML))
		case "/swagger-ui-bundle.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte("// Swagger UI bundle would be served here\n// In production, serve actual Swagger UI assets"))
		case "/swagger-ui.css":
			w.Header().Set("Content-Type", "text/css")
			_, _ = w.Write([]byte("/* Swagger UI styles would be served here */"))
		case "/openapi.yaml", "/openapi.yml":
			http.ServeFile(w, r, "./api/openapi.yaml")
		case "/openapi.json":
			w.Header().Set("Content-Type", "application/json")
			// In production, you'd convert YAML to JSON or serve a JSON version
			_, _ = w.Write([]byte(`{"info": {"title": "See /docs/openapi.yaml for full spec"}}`))
		default:
			http.NotFound(w, r)
		}
	})
}

// RegisterSwaggerRoutes registers Swagger UI routes
func RegisterSwaggerRoutes(router chi.Router) {
	// Swagger UI routes
	router.Handle("/docs/*", http.StripPrefix("/docs", SwaggerUIHandler()))

	// Direct OpenAPI spec access
	router.Get("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api/openapi.yaml")
	})

	router.Get("/api/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// In production, serve actual JSON conversion
		_, _ = w.Write([]byte(`{"info": {"title": "See /api/openapi.yaml for full spec"}}`))
	})
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>MCP Ultra v21 API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.15.5/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.15.5/favicon-16x16.png" sizes="16x16" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1976d2;
        }
        .swagger-ui .topbar .download-url-wrapper .download-url-button {
            background-color: #4caf50;
        }
        .custom-header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            text-align: center;
            margin-bottom: 0;
        }
        .custom-header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .custom-header p {
            margin: 10px 0 0 0;
            font-size: 1.1em;
            opacity: 0.9;
        }
        .version-badge {
            background: rgba(255,255,255,0.2);
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.9em;
            margin-left: 10px;
        }
        .environment-links {
            margin-top: 15px;
        }
        .environment-links a {
            color: white;
            text-decoration: none;
            margin: 0 10px;
            padding: 5px 15px;
            border: 1px solid rgba(255,255,255,0.3);
            border-radius: 15px;
            transition: background-color 0.3s;
        }
        .environment-links a:hover {
            background-color: rgba(255,255,255,0.1);
        }
        .feature-highlights {
            display: flex;
            justify-content: center;
            gap: 30px;
            margin-top: 20px;
            flex-wrap: wrap;
        }
        .feature-item {
            display: flex;
            align-items: center;
            font-size: 0.9em;
        }
        .feature-item::before {
            content: "✓";
            margin-right: 5px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="custom-header">
        <h1>MCP Ultra v21 API
            <span class="version-badge">v21.0.0</span>
        </h1>
        <p>Enterprise-grade microservice with Clean Architecture, DDD patterns, and comprehensive security</p>
        
        <div class="feature-highlights">
            <div class="feature-item">JWT Authentication</div>
            <div class="feature-item">OPA Authorization</div>
            <div class="feature-item">Feature Flags</div>
            <div class="feature-item">Multi-tenant</div>
            <div class="feature-item">Event-driven</div>
            <div class="feature-item">Production-ready</div>
        </div>
        
        <div class="environment-links">
            <a href="https://api.vertikon.com/v1" target="_blank">Production</a>
            <a href="https://staging-api.vertikon.com/v1" target="_blank">Staging</a>
            <a href="http://localhost:9655/api/v1" target="_blank">Local</a>
        </div>
    </div>

    <div id="swagger-ui"></div>

    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            // Begin Swagger UI call region
            const ui = SwaggerUIBundle({
                url: '/docs/openapi.yaml',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: "https://validator.swagger.io/validator",
                docExpansion: "list",
                operationsSorter: "alpha",
                tagsSorter: "alpha",
                filter: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log("Swagger UI loaded for MCP Ultra v21");
                },
                requestInterceptor: function(request) {
                    // Add default headers
                    request.headers['X-API-Version'] = 'v21';
                    return request;
                },
                responseInterceptor: function(response) {
                    // Log API responses for debugging
                    if (response.status >= 400) {
                        console.warn('API Error:', response.status, response.statusText);
                    }
                    return response;
                }
            });
            // End Swagger UI call region

            window.ui = ui;
            
            // Custom enhancements
            setTimeout(() => {
                // Add custom styling after load
                const style = document.createElement('style');
                style.textContent = '.swagger-ui .info .title { display: none; }' +
                    '.swagger-ui .scheme-container { background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 4px; padding: 10px; margin: 10px 0; }' +
                    '.swagger-ui .servers-title { font-weight: bold; color: #495057; }';
                document.head.appendChild(style);
            }, 1000);
        };
    </script>
</body>
</html>`
