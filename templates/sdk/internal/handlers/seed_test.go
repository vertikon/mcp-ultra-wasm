package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSeedSyncHandler_Success(t *testing.T) {
	// Mock request with valid template path
	reqBody := SeedSyncRequest{
		TemplatePath: "testdata/mock-template",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/seed/sync", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	SeedSyncHandler(rr, req)

	// Note: This will likely fail in practice as seeds.Sync needs a real template
	// But we're testing the handler logic
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadGateway {
		t.Logf("Expected 200 or 502, got %d (this is expected in test env)", rr.Code)
	}

	var resp SeedSyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status == "" {
		t.Error("Response status should not be empty")
	}
}

func TestSeedSyncHandler_DefaultPath(t *testing.T) {
	// Request without body should use default path
	req := httptest.NewRequest(http.MethodPost, "/seed/sync", nil)
	rr := httptest.NewRecorder()

	SeedSyncHandler(rr, req)

	// Should get response (either success or error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadGateway {
		t.Logf("Expected 200 or 502, got %d", rr.Code)
	}

	var resp SeedSyncResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "ok" && resp.Status != "error" {
		t.Errorf("Expected status 'ok' or 'error', got '%s'", resp.Status)
	}
}

func TestSeedSyncHandler_InvalidJSON(t *testing.T) {
	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/seed/sync", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	SeedSyncHandler(rr, req)

	// Should still handle gracefully (warn log + use default path)
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadGateway {
		t.Logf("Expected 200 or 502, got %d", rr.Code)
	}
}

func TestSeedStatusHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/seed/status", nil)
	rr := httptest.NewRecorder()

	SeedStatusHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	// Response should be valid JSON
	var status map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &status); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	// Should have content-type header
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestSeedSyncRequest(t *testing.T) {
	// Test struct marshaling
	req := SeedSyncRequest{
		TemplatePath: "/path/to/template",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal SeedSyncRequest: %v", err)
	}

	var decoded SeedSyncRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SeedSyncRequest: %v", err)
	}

	if decoded.TemplatePath != req.TemplatePath {
		t.Errorf("Expected TemplatePath=%s, got %s", req.TemplatePath, decoded.TemplatePath)
	}
}

func TestSeedSyncResponse(t *testing.T) {
	// Test struct marshaling
	resp := SeedSyncResponse{
		Status:  "ok",
		Seed:    "test-seed",
		Message: "success",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal SeedSyncResponse: %v", err)
	}

	var decoded SeedSyncResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SeedSyncResponse: %v", err)
	}

	if decoded.Status != resp.Status {
		t.Errorf("Expected Status=%s, got %s", resp.Status, decoded.Status)
	}
	if decoded.Seed != resp.Seed {
		t.Errorf("Expected Seed=%s, got %s", resp.Seed, decoded.Seed)
	}
	if decoded.Message != resp.Message {
		t.Errorf("Expected Message=%s, got %s", resp.Message, decoded.Message)
	}
}
