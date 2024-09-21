package aptabase

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	// Adjust this import as necessary
)

func TestNewClient(t *testing.T) {
	client := NewClient("EU-API-KEY")

	if client.APIKey != "EU-API-KEY" {
		t.Errorf("Expected API key to be 'EU-API-KEY', got '%s'", client.APIKey)
	}

	if client.BaseURL != "https://eu.aptabase.com" {
		t.Errorf("Expected BaseURL to be 'https://eu.aptabase.com', got '%s'", client.BaseURL)
	}
}

func TestNewSessionID(t *testing.T) {
	client := NewClient("EU-API-KEY")
	sessionID := client.NewSessionID()

	if len(sessionID) != 18 { // 10 digits + 8 random digits
		t.Errorf("Expected session ID length to be 18, got %d", len(sessionID))
	}
}

func TestEvalSessionID(t *testing.T) {
	client := NewClient("EU-API-KEY")
	oldSessionID := client.SessionID

	time.Sleep(2 * time.Second) // Ensure we exceed the default timeout

	client.EvalSessionID()
	if client.SessionID == oldSessionID {
		t.Errorf("Expected new session ID, got the same: %s", oldSessionID)
	}
}

func TestTrackEvent(t *testing.T) {
	client := NewClient("EU-API-KEY")

	// Mock HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			t.Errorf("Failed to decode JSON: %v", err)
		}

		if body["eventName"] != "test_event" {
			t.Errorf("Expected eventName to be 'test_event', got '%v'", body["eventName"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client.BaseURL = ts.URL

	// Track event
	props := map[string]interface{}{"key": "value"}
	client.TrackEvent("test_event", props)

	// Check that the queue is processed
	if len(client.eventQueue) != 0 {
		t.Errorf("Expected event queue to be empty after processing, got %d", len(client.eventQueue))
	}
}

func TestTrackEvent_Failure(t *testing.T) {
	client := NewClient("EU-API-KEY")

	// Mock HTTP server to return an error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest) // Simulate error
	}))
	defer ts.Close()

	client.BaseURL = ts.URL

	// Capture logs
	var buf bytes.Buffer
	log.SetOutput(&buf)

	// Track event
	client.TrackEvent("test_event", nil)

	// Check that error is logged
	if !strings.Contains(buf.String(), "TrackEvent failed with status code 400") {
		t.Errorf("Expected error log for status code 400, got: %s", buf.String())
	}
}

func TestSystemProps(t *testing.T) {
	props, err := systemProps()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if props["osName"] == "" || props["osVersion"] == "" {
		t.Error("Expected osName and osVersion to be set")
	}

	if props["locale"] == "" {
		t.Error("Expected locale to be set")
	}
}
