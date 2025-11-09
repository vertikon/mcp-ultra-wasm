package fsx

import "os"

// File mode constants for consistent and secure file permissions.
// These constants follow security best practices for different file types.
const (
	// FileModeUserRW provides read/write permissions for owner only (0600).
	// Use for sensitive files like credentials, secrets, or configuration files
	// containing API keys or passwords.
	FileModeUserRW os.FileMode = 0o600

	// FileModeUserRWGroupR provides read/write for owner, read for group (0640).
	// Use for application configuration files that need to be readable by
	// the application group but writable only by the owner.
	FileModeUserRWGroupR os.FileMode = 0o640

	// FileModeUserRWXGroupRX provides read/write/execute for owner,
	// read/execute for group (0750).
	// Use for executable scripts or binaries that need group execution.
	FileModeUserRWXGroupRX os.FileMode = 0o750

	// FileModePublicRead provides read permissions for owner/group/others (0644).
	// Use for non-sensitive public files like documentation or public assets.
	// WARNING: Use only for non-sensitive data.
	FileModePublicRead os.FileMode = 0o644

	// FileModeDirUserRWX provides full permissions for owner on directories (0700).
	// Use for sensitive directories containing configuration or data files.
	FileModeDirUserRWX os.FileMode = 0o700

	// FileModeDirPublic provides standard directory permissions (0755).
	// Use for non-sensitive directories that need to be readable by others.
	FileModeDirPublic os.FileMode = 0o755
)
