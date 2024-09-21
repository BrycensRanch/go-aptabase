package aptabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/brycensranch/go-aptabase/pkg/osinfo/v1" // Import your osinfo package
	"golang.org/x/exp/rand"
)

var hosts = map[string]string{
	"EU":  "https://eu.aptabase.com",
	"US":  "https://us.aptabase.com",
	"DEV": "http://localhost:3000",
	"SH":  "",
}

type Client struct {
	APIKey         string        // Public field
	BaseURL        string        // Public field
	HTTPClient     *http.Client  // Public field
	SessionID      string        // Public field
	LastTouch      time.Time     // Public field
	SessionTimeout time.Duration // Public field
	eventQueue     []map[string]interface{}
	mu             sync.Mutex
}

// NewClient creates a new Client with the specified API key and optional base URL.
func NewClient(apiKey string, baseURL ...string) *Client {
	client := &Client{
		APIKey:         apiKey,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		SessionTimeout: 1 * time.Hour, // default session timeout
		eventQueue:     []map[string]interface{}{},
	}

	// Determine the host from the App-Key
	if apiKeyContains := func() string {
		if contains(apiKey, "EU") {
			return "EU"
		}
		return "US" // Default to US if "EU" is not found
	}(); apiKeyContains != "" {
		client.BaseURL = hosts[apiKeyContains]
	} else {
		client.BaseURL = hosts["US"] // Fallback if no valid region found
	}

	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()

	go client.processQueue() // Start the queue processing in a goroutine

	return client
}

// NewSessionID generates a new session ID in the format of epochInSeconds + 8 random numbers.
func (c *Client) NewSessionID() string {
	epochSeconds := time.Now().UTC().Unix()
	randomNumber := rand.Intn(100000000) // Generates a random number between 0 and 99,999,999
	return fmt.Sprintf("%d%08d", epochSeconds, randomNumber)
}

// EvalSessionID evaluates and potentially updates the session ID based on the last touch time.
func (c *Client) EvalSessionID() string {
	now := time.Now().UTC()
	if now.Sub(c.LastTouch) > c.SessionTimeout {
		c.SessionID = c.NewSessionID()
		log.Printf("New session ID generated: %s", c.SessionID)
	}
	c.LastTouch = now
	return c.SessionID
}

// TrackEvent tracks an event with the specified name and properties.
func (c *Client) TrackEvent(eventName string, props map[string]interface{}) {
	if c.APIKey == "" {
		log.Println("Tracking is disabled: API key is empty.")
		return
	}

	systemProps, err := systemProps()
	if err != nil {
		log.Printf("Error getting system properties: %v", err)
		return
	}

	body := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"sessionId":   c.EvalSessionID(),
		"eventName":   eventName,
		"systemProps": systemProps,
		"props":       props,
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Set("App-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		log.Printf("TrackEvent failed with status code %d: %s", resp.StatusCode, string(respBody))
		return
	}

	log.Println("Event tracked successfully!")
}

// processQueue processes the in-memory event queue.
func (c *Client) processQueue() {
	for {
		time.Sleep(10 * time.Second) // Wait before processing the queue
		c.mu.Lock()
		if len(c.eventQueue) > 0 {
			for _, event := range c.eventQueue {
				if err := c.sendEvent(event); err != nil {
					log.Printf("Failed to send event: %v", err)
				}
			}
			c.eventQueue = []map[string]interface{}{} // Clear the queue after processing
		}
		c.mu.Unlock()
	}
}

// sendEvent sends an individual event to the Aptabase API.
func (c *Client) sendEvent(event map[string]interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("App-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for additional context
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		log.Printf("TrackEvent failed with status code %d: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	log.Println("Event tracked successfully!")
	return nil
}

// systemProps retrieves system information using the osinfo package.
func systemProps() (map[string]interface{}, error) {
	osName, osVersion := osinfo.GetOSInfo()

	props := map[string]interface{}{
		"isDebug":        false, // Set to true if in debug mode
		"osName":         osName,
		"osVersion":      osVersion,
		"locale":         "en_US.UTF-8",       // You can update this as needed
		"appVersion":     "1.0.0",             // Replace with actual app version
		"appBuildNumber": "100",               // Replace with actual build number
		"sdkVersion":     "go-aptabase@0.0.0", // Assuming SDK version is available in sysInfo
	}

	return props, nil
}

// Helper function to check if a substring exists in a string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
