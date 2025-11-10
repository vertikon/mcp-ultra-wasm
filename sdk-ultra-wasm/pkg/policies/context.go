// pkg/policies/context.go
package policies

import "context"

type identityKey struct{}

// Identity representa identidade autenticada
type Identity struct {
	Subject string   // User ID ou subject do token
	Roles   []string // Papéis/permissões do usuário
}

// WithIdentity adiciona identidade ao contexto
func WithIdentity(ctx context.Context, subject string, roles []string) context.Context {
	return context.WithValue(ctx, identityKey{}, &Identity{
		Subject: subject,
		Roles:   roles,
	})
}

// FromIdentity extrai identidade do contexto.
// Retorna nil se a identidade não estiver presente ou tiver tipo incorreto.
func FromIdentity(ctx context.Context) *Identity {
	if ctx == nil {
		return nil
	}
	identity, ok := ctx.Value(identityKey{}).(*Identity)
	if !ok {
		return nil
	}
	return identity
}
