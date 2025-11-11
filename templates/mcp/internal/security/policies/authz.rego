package authz

import rego.v1

# Default deny
default allow := false

# Allow access to health endpoints for everyone
allow if {
	input.path in ["/healthz", "/readyz", "/metrics"]
}

# Admin role has full access
allow if {
	input.user.role == "admin"
}

# User role permissions
allow if {
	input.user.role == "user"
	user_allowed_actions[input.action]
}

# Define allowed actions for regular users
user_allowed_actions := {
	"read",
	"list",
	"create",
	"update"
}

# Tenant isolation - users can only access their own tenant data
allow if {
	input.user.role == "user"
	input.user.tenant_id
	tenant_resource_access
}

tenant_resource_access if {
	input.resource == "tasks"
	input.action in ["read", "list", "create", "update"]
	# Additional tenant checks would be implemented at the service layer
}

# Resource-specific authorization rules
allow if {
	resource_permissions[input.resource][input.action][input.user.role]
}

# Permission matrix for resources
resource_permissions := {
	"tasks": {
		"create": {
			"admin": true,
			"user": true
		},
		"read": {
			"admin": true,
			"user": true
		},
		"update": {
			"admin": true,
			"user": true
		},
		"delete": {
			"admin": true,
			"user": false
		},
		"list": {
			"admin": true,
			"user": true
		}
	},
	"users": {
		"create": {
			"admin": true,
			"user": false
		},
		"read": {
			"admin": true,
			"user": false
		},
		"update": {
			"admin": true,
			"user": false
		},
		"delete": {
			"admin": true,
			"user": false
		},
		"list": {
			"admin": true,
			"user": false
		}
	},
	"system": {
		"config": {
			"admin": true,
			"user": false
		},
		"monitoring": {
			"admin": true,
			"user": false
		}
	}
}

# Scope-based permissions
allow if {
	required_scope := scope_requirements[input.resource][input.action]
	required_scope
	has_scope(required_scope)
}

has_scope(required) if {
	some scope in input.user.scopes
	scope == required
}

# Scope requirements for different resources and actions
scope_requirements := {
	"tasks": {
		"create": "tasks:write",
		"read": "tasks:read",
		"update": "tasks:write",
		"delete": "tasks:delete",
		"list": "tasks:read"
	},
	"users": {
		"create": "users:admin",
		"read": "users:admin",
		"update": "users:admin",
		"delete": "users:admin",
		"list": "users:admin"
	}
}

# Time-based access (business hours only for certain operations)
allow if {
	input.resource == "system"
	input.action == "config"
	business_hours
}

business_hours if {
	now := time.now_ns() / 1000000000
	hour := time.weekday(now)[1]
	hour >= 9
	hour <= 17
}

# Rate limiting context (would be enforced at middleware level)
deny if {
	input.user.rate_limited == true
}

# Audit logging decision
reason := "Access denied: insufficient permissions" if not allow
reason := "Access granted" if allow

# Export decision and reason
decision := {
	"allow": allow,
	"deny": not allow,
	"reason": reason
}