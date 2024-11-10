package aptabase

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// sendEvents sends a batch of events to the tracking service in a single request.
func (c *Client) sendEvents(events []EventData) error {
	if  len(events) == 0 && c.DebugMode {
		c.Logger.Printf("sendEvents called with no events to send! woah")
		return nil
	}
	c.wg.Add(1)
	defer c.wg.Done()
	systemProps, err := c.systemProps()
	if err != nil {
		c.Logger.Printf("Error getting system properties: %v\n", err)
		return err
	}

	// Prepare the batch of events
	var batch []map[string]interface{}
	for _, event := range events {
		if c.DebugMode {
			log.Printf("Event: %s\nData: %v\nSystemProps: %v\n", event.EventName, event.Props, systemProps)
		}

		// Add event to the batch
		batch = append(batch, map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"sessionId":   c.EvalSessionID(),
			"systemProps": systemProps,
			"eventName":   event.EventName,
			"props":       event.Props,
		})
	}
	data, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		c.Logger.Printf("Error marshalling event data: %v\n", err)
		return err
	}
	if string(data) == "null" {
		c.Logger.Printf("Event data is null!! Bug?\n")
		c.Logger.Printf("Batch %v\n", batch)
		c.Logger.Printf("Events %v\n", events)
		return nil
	}
	c.Logger.Printf("Sending data:\n%s", string(data))
	req, err := http.NewRequest("POST", c.BaseURL+"/api/v0/events", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("App-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	c.Logger.Printf("Sending events to %s", c.BaseURL)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			c.Logger.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Printf("Error reading response body: %v", err)
		return err
	}

	var respJSON map[string]interface{}
	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		c.Logger.Printf("Failed to unmarshal response body, logging raw body: %s", string(respBody))
		respJSON = map[string]interface{}{"raw_body": string(respBody)} // Store raw body as fallback for logging
		return err
	}

	if resp.StatusCode >= 300 {
		respJSONBytes, _ := json.Marshal(respJSON)
		c.Logger.Printf("TrackEvent failed with status code %d at %s: %s", resp.StatusCode, resp.Request.URL, respJSONBytes)
		return err
	}

	c.Logger.Println("Events tracked successfully!")
	return nil
}
