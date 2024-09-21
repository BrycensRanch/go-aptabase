package aptabase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")

	if client.apiKey != "test-api-key" {
		t.Errorf("Expected apiKey to be 'test-api-key', got '%s'", client.apiKey)
	}

	if client.baseURL != "https://api.aptabase.com/v1" {
		t.Errorf("Expected default baseURL, got '%s'", client.baseURL)
	}
}

func TestNewClientWithCustomBaseURL(t *testing.T) {
	client := NewClient("test-api-key", "https://custom.api.aptabase.com/v1")

	if client.baseURL != "https://custom.api.aptabase.com/v1" {
		t.Errorf("Expected baseURL to be 'https://custom.api.aptabase.com/v1', got '%s'", client.baseURL)
	}
}

func TestEvalSessionID(t *testing.T) {
	client := NewClient("test-api-key")
	oldSessionID := client.sessionID

	// Simulate a timeout
	time.Sleep(1 * time.Hour)

	newSessionID := client.evalSessionID()
	if oldSessionID == newSessionID {
		t.Error("Expected a new session ID to be generated after timeout")
	}
}

func TestTrackEventSuccess(t *testing.T) {
	client := NewClient("test-api-key")

	// Mock the HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method 'POST', got '%s'", r.Method)
		}
		if r.Header.Get("App-Key") != "test-api-key" {
			t.Errorf("Expected App-Key to be 'test-api-key', got '%s'", r.Header.Get("App-Key"))
		}

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			t.Fatalf("Error decoding body: %v", err)
		}

		if body["eventName"] != "test_event" {
			t.Errorf("Expected eventName to be 'test_event', got '%s'", body["eventName"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client.baseURL = ts.URL

	err := client.trackEvent("test_event", map[string]interface{}{"prop1": "value1"})
	if err != nil {
		t.Errorf("Expected no error, got '%v'", err)
	}
}

func TestTrackEventDisabled(t *testing.T) {
	client := NewClient("") // No API key

	err := client.trackEvent("test_event", nil)
	if err != nil {
		t.Error("Expected no error when tracking is disabled")
	}
}

func TestTrackEventHTTPError(t *testing.T) {
	client := NewClient("test-api-key")

	// Mock the HTTP server to return an error status
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client.baseURL = ts.URL

	err := client.trackEvent("test_event", nil)
	if err == nil {
		t.Error("Expected an error due to HTTP error response")
	}
}
