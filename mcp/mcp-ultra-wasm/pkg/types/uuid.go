package types

import (
	"github.com/google/uuid"
)

// UUID type re-exported from google/uuid for centralized dependency management
type UUID = uuid.UUID

// UUID generator functions
var (
	// New generates a new random UUID v4
	New = uuid.New
	// NewString generates a new random UUID v4 as a string
	NewString = uuid.NewString
	// Parse parses a UUID string
	Parse = uuid.Parse
	// MustParse parses a UUID string and panics on error
	MustParse = uuid.MustParse
	// Nil is the nil UUID (00000000-0000-0000-0000-000000000000)
	Nil = uuid.Nil
)
