package aptabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/exp/rand"
)

type Client struct {
	APIKey         string        // Public field
	BaseURL        string        // Public field
	HTTPClient     *http.Client  // Public field
	SessionID      string        // Public field
	LastTouch      time.Time     // Public field
	SessionTimeout time.Duration // Public field
}

// NewClient creates a new Client with the specified API key and optional base URL.
func NewClient(apiKey string, baseURL ...string) *Client {
	client := &Client{
		APIKey:         apiKey,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		SessionTimeout: 1 * time.Hour, // default session timeout
	}

	if len(baseURL) > 0 {
		client.BaseURL = baseURL[0]
	} else {
		client.BaseURL = "https://api.aptabase.com/v1"
	}

	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()

	return client
}

// NewSessionID generates a new session ID (you can use a better method here).
func (c *Client) NewSessionID() string {
	return RandomString()
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
func (c *Client) TrackEvent(eventName string, props map[string]interface{}) error {
	if c.APIKey == "" {
		log.Println("Tracking is disabled: API key is empty.")
		return nil
	}

	body := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"sessionId": c.EvalSessionID(),
		"eventName": eventName,
		"props":     props,
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/events", bytes.NewBuffer(data))
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

	if resp.StatusCode >= 300 {
		log.Printf("TrackEvent failed with status code %d", resp.StatusCode)
		return fmt.Errorf("failed with status code %d", resp.StatusCode)
	}

	log.Println("Event tracked successfully!")
	return nil
}

// RandomString generates a random string (replace with a better method if needed).
func RandomString() string {
	length := 12
	rand.Seed(uint64(time.Now().UnixNano()))
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
