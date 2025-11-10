package security

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryTokenStore implementação de TokenStore em memória
type MemoryTokenStore struct {
	tokens map[string]*TokenEntry
	mu     sync.RWMutex
	logger *zap.Logger
	config *TokenStoreConfig
}

type TokenEntry struct {
	TokenID   string
	Claims    *CustomClaims
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
	Reason    string
}

type TokenStoreConfig struct {
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxTokens       int           `json:"max_tokens"`
	EnableCleanup   bool          `json:"enable_cleanup"`
}

func NewMemoryTokenStore(logger *zap.Logger) TokenStore {
	config := &TokenStoreConfig{
		CleanupInterval: 1 * time.Hour,
		MaxTokens:       10000,
		EnableCleanup:   true,
	}

	store := &MemoryTokenStore{
		tokens: make(map[string]*TokenEntry),
		logger: logger.Named("tokenstore.memory"),
		config: config,
	}

	// Iniciar cleanup se habilitado
	if config.EnableCleanup {
		go store.cleanupRoutine()
	}

	logger.Info("Memory token store initialized",
		zap.Duration("cleanup_interval", config.CleanupInterval),
		zap.Int("max_tokens", config.MaxTokens))

	return store
}

func (mts *MemoryTokenStore) StoreToken(tokenID string, claims *CustomClaims) error {
	mts.mu.Lock()
	defer mts.mu.Unlock()

	// Verificar limite de tokens
	if len(mts.tokens) >= mts.config.MaxTokens {
		mts.cleanupLocked()
		if len(mts.tokens) >= mts.config.MaxTokens {
			// Remover token mais antigo se ainda estiver no limite
			mts.removeOldestToken()
		}
	}

	entry := &TokenEntry{
		TokenID:   tokenID,
		Claims:    claims,
		CreatedAt: time.Now(),
		ExpiresAt: claims.ExpiresAt.Time,
	}

	mts.tokens[tokenID] = entry
	mts.logger.Debug("Token armazenado", zap.String("token_id", tokenID))

	return nil
}

func (mts *MemoryTokenStore) RevokeToken(tokenID string) error {
	mts.mu.Lock()
	defer mts.mu.Unlock()

	entry, exists := mts.tokens[tokenID]
	if !exists {
		return ErrTokenNotFound
	}

	now := time.Now()
	entry.RevokedAt = &now
	entry.Reason = "user_logout"

	mts.logger.Info("Token revogado",
		zap.String("token_id", tokenID),
		zap.String("reason", entry.Reason))

	return nil
}

func (mts *MemoryTokenStore) IsTokenRevoked(tokenID string) (bool, error) {
	mts.mu.RLock()
	defer mts.mu.RUnlock()

	entry, exists := mts.tokens[tokenID]
	if !exists {
		return false, nil // Token não encontrado, não está revogado
	}

	// Verificar se está explicitamente revogado
	if entry.RevokedAt != nil {
		return true, nil
	}

	// Verificar se expirou
	if time.Now().After(entry.ExpiresAt) {
		return true, nil
	}

	return false, nil
}

func (mts *MemoryTokenStore) CleanupExpired() error {
	mts.mu.Lock()
	defer mts.mu.Unlock()

	return mts.cleanupLocked()
}

func (mts *MemoryTokenStore) cleanupLocked() error {
	now := time.Now()
	var removed int

	for tokenID, entry := range mts.tokens {
		// Remover tokens expirados ou revogados há muito tempo
		shouldRemove := false

		if now.After(entry.ExpiresAt) {
			shouldRemove = true
		} else if entry.RevokedAt != nil {
			// Remover tokens revogados há mais de 1 hora
			if now.Sub(*entry.RevokedAt) > time.Hour {
				shouldRemove = true
			}
		}

		if shouldRemove {
			delete(mts.tokens, tokenID)
			removed++
		}
	}

	if removed > 0 {
		mts.logger.Debug("Tokens removidos no cleanup",
			zap.Int("removed", removed),
			zap.Int("remaining", len(mts.tokens)))
	}

	return nil
}

func (mts *MemoryTokenStore) removeOldestToken() {
	var oldestTokenID string
	var oldestTime time.Time

	for tokenID, entry := range mts.tokens {
		if oldestTokenID == "" || entry.CreatedAt.Before(oldestTime) {
			oldestTokenID = tokenID
			oldestTime = entry.CreatedAt
		}
	}

	if oldestTokenID != "" {
		delete(mts.tokens, oldestTokenID)
		mts.logger.Warn("Token mais antigo removido por limite",
			zap.String("token_id", oldestTokenID))
	}
}

func (mts *MemoryTokenStore) cleanupRoutine() {
	ticker := time.NewTicker(mts.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := mts.CleanupExpired(); err != nil {
			mts.logger.Error("Erro no cleanup de tokens", zap.Error(err))
		}
	}
}

func (mts *MemoryTokenStore) GetStats() map[string]interface{} {
	mts.mu.RLock()
	defer mts.mu.RUnlock()

	total := len(mts.tokens)
	revoked := 0
	expired := 0
	now := time.Now()

	for _, entry := range mts.tokens {
		if entry.RevokedAt != nil {
			revoked++
		} else if now.After(entry.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_tokens":   total,
		"active_tokens":  total - revoked - expired,
		"revoked_tokens": revoked,
		"expired_tokens": expired,
		"max_tokens":     mts.config.MaxTokens,
	}
}

// Erros
var (
	ErrTokenNotFound = fmt.Errorf("token not found")
)
