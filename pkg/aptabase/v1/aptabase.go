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
	apiKey         string
	baseURL        string
	httpClient     *http.Client
	sessionID      string
	lastTouch      time.Time
	sessionTimeout time.Duration
}

func NewClient(apiKey string, baseURL ...string) *Client {
	client := &Client{
		apiKey:         apiKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		sessionTimeout: 1 * time.Hour, // default session timeout
	}

	if len(baseURL) > 0 {
		client.baseURL = baseURL[0]
	} else {
		client.baseURL = "https://api.aptabase.com/v1"
	}

	client.sessionID = client.newSessionID()
	client.lastTouch = time.Now().UTC()

	return client
}

func (c *Client) newSessionID() string {
	// Generate a new session ID (you can use a better method here)
	return RandomString()
}

func (c *Client) evalSessionID() string {
	now := time.Now().UTC()
	if now.Sub(c.lastTouch) > c.sessionTimeout {
		c.sessionID = c.newSessionID()
		log.Printf("New session ID generated: %s", c.sessionID)
	}
	c.lastTouch = now
	return c.sessionID
}

func (c *Client) trackEvent(eventName string, props map[string]interface{}) error {
	if c.apiKey == "" {
		log.Println("Tracking is disabled: API key is empty.")
		return nil
	}

	body := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"sessionId": c.evalSessionID(),
		"eventName": eventName,
		"props":     props,
	}

	data, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/events", bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("App-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

// RandomString generates a random string (replace with a better method if needed)
func RandomString() string {
	length := 12
	rand.Seed(uint64(time.Now().UnixNano()))
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
