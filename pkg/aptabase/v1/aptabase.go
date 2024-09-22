package aptabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/brycensranch/go-aptabase/pkg/locale"
	"github.com/brycensranch/go-aptabase/pkg/osinfo/v1"
	"golang.org/x/exp/rand"
)

var hosts = map[string]string{
	"EU":  "https://eu.aptabase.com",
	"US":  "https://us.aptabase.com",
	"SH":  "",
	"DEV": "http://localhost:3000",
}

// EventData represents the structure of the event data passed to TrackEvent.
type EventData struct {
	EventName string                 `json:"eventName"`
	Props     map[string]interface{} `json:"props"`
}

// Client represents a tracking client.
type Client struct {
	APIKey         string
	BaseURL        string
	HTTPClient     *http.Client
	SessionID      string
	LastTouch      time.Time
	SessionTimeout time.Duration
	eventQueue     []EventData
	mu             sync.Mutex
	AppVersion     string
	AppBuildNumber uint64
	DebugMode      bool
	quitChan       chan struct{}
}

// NewClient creates a new Client with the specified parameters.
func NewClient(apiKey, appVersion string, appBuildNumber uint64, debugMode bool, baseURL string) *Client {
	client := &Client{
		APIKey:         apiKey,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		SessionTimeout: 1 * time.Hour,
		eventQueue:     []EventData{},
		AppVersion:     appVersion,
		AppBuildNumber: appBuildNumber,
		DebugMode:      debugMode,
		quitChan:       make(chan struct{}),
	}

	client.BaseURL = client.determineHost(apiKey)
	if strings.Contains(client.APIKey, "SH") {
		client.BaseURL = baseURL
	}
	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()

	go client.processQueue()

	return client
}

// determineHost selects the host URL based on the AppKey.
func (c *Client) determineHost(apiKey string) string {
	if strings.Contains(apiKey, "EU") {
		return hosts["EU"]
	} else if strings.Contains(apiKey, "DEV") {
		return hosts["DEV"]
	}
	return hosts["US"]
}

// NewSessionID generates a new session ID in the format of epochInSeconds + 8 random numbers.
func (c *Client) NewSessionID() string {
	epochSeconds := time.Now().UTC().Unix()
	randomNumber := rand.Intn(100000000)
	return fmt.Sprintf("%d%08d", epochSeconds, randomNumber)
}

// EvalSessionID evaluates and updates the session ID if the session has expired.
func (c *Client) EvalSessionID() string {
	now := time.Now().UTC()
	if now.Sub(c.LastTouch) > c.SessionTimeout {
		c.SessionID = c.NewSessionID()
	}
	c.LastTouch = now
	return c.SessionID
}

// processQueue processes the queued events periodically, batching them into a single request.
func (c *Client) processQueue() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.sendQueuedEvents()
		case <-c.quitChan:
			// Process remaining events before quitting
			c.sendQueuedEvents()
			return
		}
	}
}

// sendQueuedEvents sends the queued events to the tracking service.
func (c *Client) sendQueuedEvents() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.eventQueue) == 0 {
		println("Event queue is 0 yet sendQueuedEvents was called! Bug?")
		return
	}

	// Copy and clear the event queue
	batch := make([]EventData, len(c.eventQueue))
	copy(batch, c.eventQueue)
	c.eventQueue = []EventData{}

	c.mu.Unlock()

	if err := c.sendEvents(batch); err != nil {
		log.Printf("Failed to send events: %v", err)
	}
}

// Stop gracefully stops the event processing and sends any remaining events.
func (c *Client) Stop() {
	close(c.quitChan)
}

// sendEvents sends a batch of events to the tracking service in a single request.
func (c *Client) sendEvents(events []EventData) error {
	systemProps, err := c.systemProps()
	if err != nil {
		log.Printf("Error getting system properties: %v", err)
		return err
	}

	var batch []map[string]interface{}
	for _, event := range events {
		if c.DebugMode {
			fmt.Printf("Event: %s\nData: %v\nSystemProps: %v", event.EventName, event.Props, systemProps)
		}
		batch = append(batch, map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"sessionId":   c.EvalSessionID(),
			"systemProps": systemProps,
			"eventName":   event.EventName,
			"props":       event.Props,
		})
	}

	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/v0/events", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("App-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		log.Printf("TrackEvent failed with status code %d: %v", resp.StatusCode, respBody)
		return nil
	}

	log.Println("Events tracked successfully!")
	return nil
}

// TrackEvent queues an event with the specified EventData for tracking.
func (c *Client) TrackEvent(event EventData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Printf("Queuing event: %s\nEvent data: %v", event.EventName, event.Props)
	c.eventQueue = append(c.eventQueue, event)

	if len(c.eventQueue) > 1000 { // Example limit, adjust as needed
		log.Println("Event queue size exceeds limit, consider sending events.")
	}
}

// systemProps retrieves system information using the osinfo package,
// and includes Client-specific details like AppVersion, AppBuildNumber, and DebugMode.
func (c *Client) systemProps() (map[string]interface{}, error) {
	osName, osVersion := osinfo.GetOSInfo()

	props := map[string]interface{}{
		"isDebug":        c.DebugMode,
		"osName":         osName,
		"osVersion":      osVersion,
		"engineName":     "go",
		"engineVersion":  runtime.Version(),
		"locale":         locale.GetLocale(),
		"appVersion":     c.AppVersion,
		"appBuildNumber": c.AppBuildNumber,
		// TODO: Embed VERSION file into code...
		"sdkVersion": "go-aptabase@0.0.0",
	}

	return props, nil
}
