package aptabase

import (
	"time"
)

// processQueue processes the queued events periodically, batching them into a single request.
func (c *Client) processQueue() {
	c.Logger.Printf("processQueue started")
	batch := make([]EventData, 0, 999)

	for {
		select {
		case event := <-c.eventChan:
			c.Logger.Printf("processQueue received eventChan %s", event.EventName)
			c.handleEvent(&batch, event)
		case <-time.After(500 * time.Millisecond):
			if c.Quit {
				c.flushBatch(&batch)
				batch = make([]EventData, 0, 999)
			}
		}
	}
}

// handleEvent processes an incoming event by appending it to the current batch.
func (c *Client) handleEvent(batch *[]EventData, event EventData) {
	c.Logger.Printf("processQueue received event: %+v", event)
	*batch = append(*batch, event)
	c.Logger.Printf("processQueue current batch: %v", *batch)
	if len(*batch) >= 10 {
		c.sendBatch(*batch)
		*batch = make([]EventData, 0, 999)
	}
}

// checkAndFlushBatch checks if there are any remaining events in the batch and sends them if necessary.
func (c *Client) checkAndFlushBatch(batch *[]EventData) {
	if c.Quit && len(*batch) > 0 {
		c.Logger.Printf("You have held current batch: %v", *batch)
		c.sendBatch(*batch)
		*batch = make([]EventData, 0, 999)
	}
}

// sendBatch sends the events in the provided batch and waits for completion.
func (c *Client) sendBatch(batch []EventData) {
	c.wg.Add(1)
	go func(batchToSend []EventData) {
		defer c.wg.Done()

		err := c.sendEvents(batchToSend)
		if err != nil {
			c.Logger.Printf("Error sending events: %v", err)
		}
	}(batch)
}

// flushBatch sends any remaining events in the batch before quitting.
func (c *Client) flushBatch(batch *[]EventData) {
	if len(*batch) > 0 {
		c.Logger.Printf("Flushing events: %v", *batch)
		c.sendBatch(*batch)
	}
}
