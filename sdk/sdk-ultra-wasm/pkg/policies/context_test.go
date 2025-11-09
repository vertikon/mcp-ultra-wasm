// pkg/policies/context_test.go
package policies

import (
	"context"
	"testing"
)

func TestWithIdentity(t *testing.T) {
	ctx := context.Background()
	subject := "user123"
	roles := []string{"admin", "editor"}

	ctx = WithIdentity(ctx, subject, roles)

	identity := FromIdentity(ctx)
	if identity == nil {
		t.Fatal("FromIdentity returned nil")
	}

	if identity.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, identity.Subject)
	}

	if len(identity.Roles) != len(roles) {
		t.Errorf("Expected %d roles, got %d", len(roles), len(identity.Roles))
	}

	for i, role := range roles {
		if identity.Roles[i] != role {
			t.Errorf("Expected role %s at index %d, got %s", role, i, identity.Roles[i])
		}
	}
}

func TestFromIdentityNoIdentity(t *testing.T) {
	ctx := context.Background()

	identity := FromIdentity(ctx)
	if identity != nil {
		t.Error("Expected FromIdentity to return nil for context without identity")
	}
}

func TestFromIdentityWrongType(t *testing.T) {
	ctx := context.Background()
	// Add a value with the same key but wrong type
	ctx = context.WithValue(ctx, identityKey{}, "not an identity")

	identity := FromIdentity(ctx)
	if identity != nil {
		t.Error("Expected FromIdentity to return nil for wrong type")
	}
}

func TestIdentityEmptyRoles(t *testing.T) {
	ctx := context.Background()
	subject := "user456"
	roles := []string{}

	ctx = WithIdentity(ctx, subject, roles)

	identity := FromIdentity(ctx)
	if identity == nil {
		t.Fatal("FromIdentity returned nil")
	}

	if identity.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, identity.Subject)
	}

	if len(identity.Roles) != 0 {
		t.Errorf("Expected 0 roles, got %d", len(identity.Roles))
	}
}
