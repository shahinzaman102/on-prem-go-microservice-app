package unit

import (
	"authentication/api"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONHelper(t *testing.T) { // Renamed from TestWriteJSON
	app := &api.Config{}
	rec := httptest.NewRecorder()

	// Use a public method to check response
	payload := map[string]string{"message": "test"}
	err := app.WriteJSON(rec, http.StatusOK, payload)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestReadJSONHelper(t *testing.T) { // Renamed from TestReadJSON
	app := &api.Config{}
	body := map[string]string{"key": "value"}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/", bytes.NewReader(jsonBody))

	var target map[string]string
	err := app.ReadJSON(nil, req, &target)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}

	if target["key"] != "value" {
		t.Errorf("expected 'value', got '%v'", target["key"])
	}
}
