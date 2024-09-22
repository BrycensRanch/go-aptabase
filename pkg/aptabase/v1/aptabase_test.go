package aptabase

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

// MockRoundTripper is a mock implementation of http.RoundTripper.
type MockRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

// TestNewClient verifies that a new Client is created with the correct initial values.
func TestNewClient(t *testing.T) {
	client := NewClient("A-US-LOL", "1.0.0", 42, true, "http://localhost:3000")

	if client.APIKey != "A-US-LOL" {
		t.Errorf("expected APIKey to be 'API_KEY', got '%s'", client.APIKey)
	}
	if client.AppVersion != "1.0.0" {
		t.Errorf("expected AppVersion to be '1.0.0', got '%s'", client.AppVersion)
	}
	if client.AppBuildNumber != 42 {
		t.Errorf("expected AppBuildNumber to be 42, got %d", client.AppBuildNumber)
	}
	if client.DebugMode != true {
		t.Errorf("expected DebugMode to be true, got false")
	}
	if client.BaseURL != "https://us.aptabase.com" {
		t.Errorf("expected BaseURL to be 'https://us.aptabase.com', got '%s'", client.BaseURL)
	}
	if client.SessionID == "" {
		t.Errorf("expected SessionID to be set")
	}
	if client.LastTouch.IsZero() {
		t.Errorf("expected LastTouch to be set")
	}
	if client.SessionTimeout != time.Hour {
		t.Errorf("expected SessionTimeout to be 1 hour, got %v", client.SessionTimeout)
	}
	if client.eventQueue == nil {
		t.Errorf("expected eventQueue to be initialized")
	}
}

// TestTrackEvent verifies that TrackEvent adds the event to the queue.
func TestTrackEvent(t *testing.T) {
	client := NewClient("A-US-LOL", "1.0.0", 42, true, "http://localhost:3000")
	event := EventData{EventName: "TestEvent", Props: map[string]interface{}{"testKey": "testValue"}}

	client.TrackEvent(event)

	if len(client.eventQueue) != 1 {
		t.Errorf("expected eventQueue length to be 1, got %d", len(client.eventQueue))
	}
	if client.eventQueue[0].EventName != "TestEvent" {
		t.Errorf("expected EventName to be 'TestEvent', got '%s'", client.eventQueue[0].EventName)
	}
}

// TestSendEvents verifies that the events are sent as a batch in a single HTTP request.
func TestSendEvents(t *testing.T) {
	client := NewClient("A-US-LOL", "1.0.0", 42, true, "http://localhost:3000")

	mockRoundTripper := &MockRoundTripper{}

	// Create an HTTP client with the mock RoundTripper
	client.HTTPClient = &http.Client{
		Transport: mockRoundTripper,
	}

	events := []EventData{
		{EventName: "TestEvent1", Props: map[string]interface{}{"key1": "value1"}},
		{EventName: "TestEvent2", Props: map[string]interface{}{"key2": "value2"}},
	}

	mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
		// Check that the request contains the correct events
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var receivedEvents []map[string]interface{}
		if err := json.Unmarshal(body, &receivedEvents); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if len(receivedEvents) != 2 {
			t.Errorf("expected 2 events, got %d", len(receivedEvents))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		}, nil
	}

	err := client.sendEvents(events)
	if err != nil {
		t.Fatalf("sendEvents failed: %v", err)
	}
}

// TestProcessQueue verifies that processQueue sends the batch of events.
func TestProcessQueue(t *testing.T) {
	client := NewClient("A-US-LOL", "1.0.0", 42, true, "http://localhost:3000")

	mockRoundTripper := &MockRoundTripper{}

	// Create an HTTP client with the mock RoundTripper
	client.HTTPClient = &http.Client{
		Transport: mockRoundTripper,
	}

	// Queue events
	event1 := EventData{EventName: "Event1", Props: map[string]interface{}{"key1": "value1"}}
	event2 := EventData{EventName: "Event2", Props: map[string]interface{}{"key2": "value2"}}

	client.TrackEvent(event1)
	client.TrackEvent(event2)

	mockRoundTripper.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
		// Verify that the request is being made with the correct data
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var receivedEvents []map[string]interface{}
		if err := json.Unmarshal(body, &receivedEvents); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if len(receivedEvents) != 2 {
			t.Errorf("expected 2 events, got %d", len(receivedEvents))
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		}, nil
	}

	// Run the process queue once
	go client.processQueue()

	time.Sleep(15 * time.Second) // Wait for the ticker to fire

	// Verify that the queue was processed and is now empty
	if len(client.eventQueue) != 0 {
		t.Errorf("expected eventQueue to be empty, got %d", len(client.eventQueue))
	}
}

// TestSystemProps verifies that systemProps returns correct system information.
func TestSystemProps(t *testing.T) {
	client := NewClient("API_KEY", "1.0.0", 42, true, "http://localhost:3000")
	props, err := client.systemProps()

	if err != nil {
		t.Fatalf("systemProps returned an error: %v", err)
	}

	// Ensure that the required system properties are included
	if props["isDebug"] != true {
		t.Errorf("expected isDebug to be true")
	}
	if props["appVersion"] != "1.0.0" {
		t.Errorf("expected appVersion to be '1.0.0', got '%s'", props["appVersion"])
	}
	if props["appBuildNumber"] != uint64(42) {
		t.Errorf("expected appBuildNumber to be 42, got %v", props["appBuildNumber"])
	}
	if !strings.HasPrefix(props["engineVersion"].(string), "go") {
		t.Errorf("expected engineVersion to start with 'go', got '%s'", props["engineVersion"])
	}
}
