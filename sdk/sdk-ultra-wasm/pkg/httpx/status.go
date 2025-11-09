package httpx

// HTTP status codes used throughout the application.
// These constants provide a centralized, type-safe way to reference status codes
// and eliminate magic numbers in the codebase.
const (
	// Success status codes
	StatusOK        = 200 // StatusOK indicates the request succeeded
	StatusNoContent = 204 // StatusNoContent indicates success with no response body

	// Client error status codes
	StatusBadRequest   = 400 // StatusBadRequest indicates invalid request
	StatusUnauthorized = 401 // StatusUnauthorized indicates missing or invalid authentication
	StatusForbidden    = 403 // StatusForbidden indicates insufficient permissions

	// Server error status codes
	StatusInternalServerError = 500 // StatusInternalServerError indicates an internal error
	StatusBadGateway          = 502 // StatusBadGateway indicates upstream server error
	StatusServiceUnavailable  = 503 // StatusServiceUnavailable indicates service is temporarily unavailable
)
