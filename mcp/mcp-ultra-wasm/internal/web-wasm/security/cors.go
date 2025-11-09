package security

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORSManager gerencia configurações CORS
type CORSManager struct {
	logger *zap.Logger
	config *CORSConfig
}

type CORSConfig struct {
	Enabled                 bool          `json:"enabled"`
	AllowedOrigins          []string      `json:"allowed_origins"`
	AllowedMethods          []string      `json:"allowed_methods"`
	AllowedHeaders          []string      `json:"allowed_headers"`
	ExposedHeaders          []string      `json:"exposed_headers"`
	AllowCredentials        bool          `json:"allow_credentials"`
	MaxAge                  time.Duration `json:"max_age"`
	OptionsPassthrough      bool          `json:"options_passthrough"`
	UnsafeWildcardOrigin    bool          `json:"unsafe_wildcard_origin"`
	IgnoreCredentialOptions bool          `json:"ignore_credential_options"`
}

func NewCORSManager(config *CORSConfig, logger *zap.Logger) *CORSManager {
	if config == nil {
		config = &CORSConfig{
			Enabled: true,
			AllowedOrigins: []string{
				"http://localhost:8080",
				"http://localhost:3000",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:3000",
			},
			AllowedMethods: []string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodHead,
				http.MethodOptions,
			},
			AllowedHeaders: []string{
				"Origin",
				"Content-Type",
				"Accept",
				"Authorization",
				"X-Requested-With",
				"X-Trace-ID",
				"X-Span-ID",
			},
			ExposedHeaders: []string{
				"X-Trace-ID",
				"X-Span-ID",
				"Content-Length",
				"Content-Type",
			},
			AllowCredentials:        false,
			MaxAge:                  12 * time.Hour,
			OptionsPassthrough:      false,
			UnsafeWildcardOrigin:    false,
			IgnoreCredentialOptions: false,
		}
	}

	manager := &CORSManager{
		logger: logger.Named("cors"),
		config: config,
	}

	logger.Info("CORS manager initialized",
		zap.Bool("enabled", config.Enabled),
		zap.Strings("allowed_origins", config.AllowedOrigins),
		zap.Bool("allow_credentials", config.AllowCredentials))

	return manager
}

// CORSMiddleware middleware CORS
func (cm *CORSManager) CORSMiddleware() gin.HandlerFunc {
	if !cm.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method

		// Verificar se é uma requisição de pré-voo (pre-flight)
		if method == http.MethodOptions {
			cm.handlePreflight(c, origin)
			return
		}

		// Processar requisição normal
		cm.handleRequest(c, origin)
	}
}

// handlePreflight processa requisições de pré-voo OPTIONS
func (cm *CORSManager) handlePreflight(c *gin.Context, origin string) {
	// Verificar se origin é permitido
	if !cm.isOriginAllowed(origin) {
		cm.logger.Warn("Origin não permitida para pré-voo", zap.String("origin", origin))
		c.Status(http.StatusForbidden)
		return
	}

	// Configurar headers CORS
	cm.setCORSHeaders(c, origin)

	// Verificar método solicitado
	method := c.Request.Header.Get("Access-Control-Request-Method")
	if method != "" && !cm.isMethodAllowed(method) {
		cm.logger.Warn("Método não permitido para pré-voo",
			zap.String("method", method),
			zap.String("origin", origin))
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	// Verificar headers solicitados
	headers := c.Request.Header.Get("Access-Control-Request-Headers")
	if headers != "" && !cm.areHeadersAllowed(headers) {
		cm.logger.Warn("Headers não permitidos para pré-voo",
			zap.String("headers", headers),
			zap.String("origin", origin))
		c.Status(http.StatusForbidden)
		return
	}

	// Retornar sucesso para pré-voo
	c.Status(http.StatusNoContent)
}

// handleRequest processa requisições normais
func (cm *CORSManager) handleRequest(c *gin.Context, origin string) {
	// Se não há origin, não é requisição cross-origin
	if origin == "" {
		c.Next()
		return
	}

	// Verificar se origin é permitida
	if !cm.isOriginAllowed(origin) {
		cm.logger.Warn("Origin não permitida", zap.String("origin", origin))
		c.Status(http.StatusForbidden)
		return
	}

	// Configurar headers CORS
	cm.setCORSHeaders(c, origin)

	// Se credentials não são permitidos mas origin tem credenciais, bloquear
	if !cm.config.AllowCredentials && cm.hasCredentials(c) {
		cm.logger.Warn("Credenciais não permitidas para origin",
			zap.String("origin", origin))
		c.Status(http.StatusForbidden)
		return
	}

	c.Next()
}

// setCORSHeaders configura os headers CORS na response
func (cm *CORSManager) setCORSHeaders(c *gin.Context, origin string) {
	// Origin
	if cm.config.UnsafeWildcardOrigin || len(cm.config.AllowedOrigins) == 0 {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	// Methods
	if len(cm.config.AllowedMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(cm.config.AllowedMethods, ", "))
	}

	// Headers
	if len(cm.config.AllowedHeaders) > 0 {
		c.Header("Access-Control-Allow-Headers", strings.Join(cm.config.AllowedHeaders, ", "))
	}

	// Exposed Headers
	if len(cm.config.ExposedHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(cm.config.ExposedHeaders, ", "))
	}

	// Credentials
	if cm.config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// Max Age
	if cm.config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", cm.config.MaxAge.String())
	}

	// Vary header para cache
	if len(cm.config.AllowedOrigins) > 0 && !cm.config.UnsafeWildcardOrigin {
		c.Header("Vary", "Origin")
	}
}

// isOriginAllowed verifica se origin está na lista de permitidas
func (cm *CORSManager) isOriginAllowed(origin string) bool {
	if len(cm.config.AllowedOrigins) == 0 {
		return true
	}

	if cm.config.UnsafeWildcardOrigin {
		return true
	}

	for _, allowedOrigin := range cm.config.AllowedOrigins {
		if cm.matchOrigin(allowedOrigin, origin) {
			return true
		}
	}

	return false
}

// matchOrigin verifica se origin corresponde ao padrão permitido
func (cm *CORSManager) matchOrigin(allowed, origin string) bool {
	// Match exato
	if allowed == origin {
		return true
	}

	// Wildcard no final (ex: https://*.example.com)
	if strings.HasSuffix(allowed, "*") {
		prefix := strings.TrimSuffix(allowed, "*")
		return strings.HasPrefix(origin, prefix)
	}

	// TODO: Implementar matching mais avançado se necessário
	// - Regex
	// - Subdomínios específicos

	return false
}

// isMethodAllowed verifica se método é permitido
func (cm *CORSManager) isMethodAllowed(method string) bool {
	if len(cm.config.AllowedMethods) == 0 {
		return true
	}

	for _, allowedMethod := range cm.config.AllowedMethods {
		if allowedMethod == method {
			return true
		}
	}

	return false
}

// areHeadersAllowed verifica se headers são permitidos
func (cm *CORSManager) areHeadersAllowed(headers string) bool {
	if len(cm.config.AllowedHeaders) == 0 {
		return true
	}

	requestedHeaders := strings.Split(headers, ",")
	for _, header := range requestedHeaders {
		header = strings.TrimSpace(header)
		if !cm.isHeaderAllowed(header) {
			return false
		}
	}

	return true
}

// isHeaderAllowed verifica se header específico é permitido
func (cm *CORSManager) isHeaderAllowed(header string) bool {
	// Case-insensitive comparison
	header = strings.ToLower(header)

	for _, allowedHeader := range cm.config.AllowedHeaders {
		if strings.ToLower(allowedHeader) == header || allowedHeader == "*" {
			return true
		}
	}

	return false
}

// hasCredentials verifica se requisição tem credenciais
func (cm *CORSManager) hasCredentials(c *gin.Context) bool {
	// Verificar cookie header
	if len(c.Request.Cookies()) > 0 {
		return true
	}

	// Verificar Authorization header
	if c.GetHeader("Authorization") != "" {
		return true
	}

	// Verificar outros headers que indicam credenciais
	if c.GetHeader("X-Requested-With") != "" {
		return true
	}

	return false
}

// UpdateConfig atualiza configuração CORS em runtime
func (cm *CORSManager) UpdateConfig(config *CORSConfig) {
	cm.config = config
	cm.logger.Info("CORS configuration updated",
		zap.Strings("allowed_origins", config.AllowedOrigins),
		zap.Bool("allow_credentials", config.AllowCredentials))
}

// GetConfig retorna configuração atual
func (cm *CORSManager) GetConfig() *CORSConfig {
	return cm.config
}

// AddOrigin adiciona origin à lista de permitidas
func (cm *CORSManager) AddOrigin(origin string) {
	// Verificar se origin já existe
	for _, existing := range cm.config.AllowedOrigins {
		if existing == origin {
			return
		}
	}

	cm.config.AllowedOrigins = append(cm.config.AllowedOrigins, origin)
	cm.logger.Info("Origin adicionada", zap.String("origin", origin))
}

// RemoveOrigin remove origin da lista de permitidas
func (cm *CORSManager) RemoveOrigin(origin string) {
	for i, existing := range cm.config.AllowedOrigins {
		if existing == origin {
			cm.config.AllowedOrigins = append(
				cm.config.AllowedOrigins[:i],
				cm.config.AllowedOrigins[i+1:]...,
			)
			cm.logger.Info("Origin removida", zap.String("origin", origin))
			return
		}
	}
}

// CORSInfo retorna informações sobre configuração CORS
func (cm *CORSManager) CORSInfo() map[string]interface{} {
	return map[string]interface{}{
		"enabled":                cm.config.Enabled,
		"allowed_origins":        cm.config.AllowedOrigins,
		"allowed_methods":        cm.config.AllowedMethods,
		"allowed_headers":        cm.config.AllowedHeaders,
		"exposed_headers":        cm.config.ExposedHeaders,
		"allow_credentials":      cm.config.AllowCredentials,
		"max_age":                cm.config.MaxAge.String(),
		"options_passthrough":    cm.config.OptionsPassthrough,
		"unsafe_wildcard_origin": cm.config.UnsafeWildcardOrigin,
	}
}

// Middleware customizado para CORS mais granular
func (cm *CORSManager) CustomCORSMiddleware(customConfig *CORSConfig) gin.HandlerFunc {
	if customConfig == nil {
		customConfig = cm.config
	}

	originalConfig := cm.config
	cm.config = customConfig

	return func(c *gin.Context) {
		cm.CORSMiddleware()(c)
		cm.config = originalConfig
	}
}

// CORSErrorHandler middleware para tratar erros CORS
func (cm *CORSManager) CORSErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Se houve erro e é requisição cross-origin
		if len(c.Errors) > 0 {
			origin := c.Request.Header.Get("Origin")
			if origin != "" && cm.isOriginAllowed(origin) {
				// Adicionar headers CORS mesmo em erro
				cm.setCORSHeaders(c, origin)
			}
		}
	}
}
