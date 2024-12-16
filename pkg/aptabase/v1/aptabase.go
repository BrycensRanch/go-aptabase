package aptabase

import (
	"fmt"
	"golang.org/x/exp/rand"
	"strings"
	"time"
)

var hosts = map[string]string{
	"EU":  "https://eu.aptabase.com",
	"US":  "https://us.aptabase.com",
	"SH":  "",
	"DEV": "http://localhost:3000",
}

// EventData represents the structure of the event data passed to TrackEvent.
type EventData struct {
	EventName   string                 `json:"eventName"`
	Props       map[string]interface{} `json:"props"`
	Timestamp   string                 `json:"Timestamp"`
	SessionId   string                 `json:"SessionId"`
	SystemProps map[string]interface{} `json:"SystemProps"`
}

// determineHost selects the host URL based on the AppKey.
func (c *Client) determineHost(apiKey string) string {
	parts := strings.Split(apiKey, "-")
	if len(parts) < 2 {
		 panic("Invalid API Key format")
	}

	regionCode := parts[1] // The second part should be the region code (e.g., "US" from "A-US-0343858461")

	host, exists := hosts[regionCode]
	if !exists {
		panic("Host not found for region " + regionCode)
	}

	return host
}

// NewSessionID generates a new session ID in the format of epochInSeconds + 8 random numbers.
func (c *Client) NewSessionID() string {
	rand.Seed(uint64(time.Now().UnixNano()))
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

// Stop gracefully stops the event processing and sends any remaining events.
func (c *Client) Stop() {
	c.Logger.Println("Stop called")
	c.Quit = true
	c.quitChan <- struct{}{}
	close(c.quitChan)
	if (c.DebugMode) { c.Logger.Printf("Starting to wait for goroutines to finish.") }

	timeout := time.After(5 * time.Second)

	// Use select to either wait for all goroutines to finish or a timeout
	done := make(chan struct{})

	go func() {
		// Wait for all goroutines to finish
		c.wg.Wait()
		if (finishedFlushing) {
			done <- struct{}{} // Signal that all goroutines are finished
		} 
	}()

	select {
	case <-done:
		if (c.DebugMode) { c.Logger.Printf("Finished waiting!") }
	case <-timeout:
		// Timeout occurred
		if (c.DebugMode) { c.Logger.Println("Timeout reached before all goroutines finished.") }
	}
}

// TrackEvent queues an event with the specified EventData for tracking.
func (c *Client) TrackEvent(event EventData) {
	if c.DebugMode {
		c.Logger.Printf("TrackEvent called with event: %+v", event)
	}
	c.eventChan <- event
}
