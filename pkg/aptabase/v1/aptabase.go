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
	APIKey           string
	BaseURL          string
	HTTPClient       *http.Client
	SessionID        string
	LastTouch        time.Time
	SessionTimeout   time.Duration
	eventChan        chan EventData
	mu               sync.Mutex
	AppVersion       string
	AppBuildNumber   uint64
	DebugMode        bool
	quitChan         chan struct{}
	processQueueOnce sync.Once
}

// NewClient creates a new Client with the specified parameters.
func NewClient(apiKey, appVersion string, appBuildNumber uint64, debugMode bool, baseURL string) *Client {
	client := &Client{
		APIKey:           apiKey,
		HTTPClient:       &http.Client{Timeout: 10 * time.Second},
		SessionTimeout:   1 * time.Hour,
		eventChan:        make(chan EventData, 10), // Buffered channel for events
		AppVersion:       appVersion,
		AppBuildNumber:   appBuildNumber,
		DebugMode:        debugMode,
		quitChan:         make(chan struct{}),
		processQueueOnce: sync.Once{},
	}

	client.BaseURL = client.determineHost(apiKey)
	if strings.Contains(client.APIKey, "SH") {
		client.BaseURL = baseURL
	}
	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()

	client.processQueueOnce.Do(client.processQueue)

	log.Printf("NewClient created with APIKey=%s, BaseURL=%s, SessionID=%s", client.APIKey, client.BaseURL, client.SessionID)

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
	log.Printf("NewSessionID called")
	epochSeconds := time.Now().UTC().Unix()
	randomNumber := rand.Intn(100000000)
	return fmt.Sprintln("%d%08d", epochSeconds, randomNumber)
}

// EvalSessionID evaluates and updates the session ID if the session has expired.
func (c *Client) EvalSessionID() string {
	log.Println("EvalSessionID called")
	now := time.Now().UTC()
	if now.Sub(c.LastTouch) > c.SessionTimeout {
		c.SessionID = c.NewSessionID()
	}
	c.LastTouch = now
	return c.SessionID
}

// processQueue processes the queued events periodically, batching them into a single request.

func (c *Client) processQueue() {
	log.Printf("processQueue started")
	batch := make([]EventData, 0, 10) // Pre-allocate a slice to hold up to 10 events

	for {
		select {
		case event := <-c.eventChan:
			log.Printf("processQueue received event: %+v", event)
			batch = append(batch, event)
			if len(batch) == cap(batch) {
				// Batch is full, send it
				go func() {
					err := c.sendEvents(batch)
					if err != nil {
						log.Printf("Error sending events: %v", err)
					}
				}()
				batch = make([]EventData, 0, 10) // Reset the batch for next events
			}
		case <-c.quitChan:
			log.Printf("processQueue received quitChan")
			// Drain any remaining events before exiting
			if len(batch) > 0 {
				go func() {
					err := c.sendEvents(batch)
					if err != nil {
						log.Printf("Error sending events: %v", err)
					}
				}()
			}
			return
		}
	}
}

// Stop gracefully stops the event processing and sends any remaining events.
func (c *Client) Stop() {
	log.Println("Stop called")
	close(c.quitChan)
}

// sendEvents sends a batch of events to the tracking service in a single request.
func (c *Client) sendEvents(events []EventData) error {
	systemProps, err := c.systemProps()
	if err != nil {
		log.Println("Error getting system properties: %v", err)
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

	log.Printf("Sending events to %s", c.BaseURL)

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
	log.Printf("TrackEvent called with event: %+v", event)
	c.eventChan <- event
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
	fmt.Printf("systemProps: %v", props)

	return props, nil
}
