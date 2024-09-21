package example

import (
	"fmt"
	"time"

	"github.com/brycensranch/go-aptabase/pkg/aptabase/v1"
)

type Event struct {
	EventName   string
	EventParams map[string]string
	Timestamp   time.Time
}

func main() {
	client := aptabase.NewClient("your-api-key")

	// Create an event to track
	event := Event{
		EventName: "user_login",
		EventParams: map[string]string{
			"user_id": "12345",
			"device":  "mobile",
		},
		Timestamp: time.Now(),
	}

	// Convert EventParams to map[string]interface{}
	eventParams := make(map[string]interface{})
	for k, v := range event.EventParams {
		eventParams[k] = v
	}

	// Track the event
	err := client.TrackEvent(event.EventName, eventParams)
	if err != nil {
		fmt.Printf("Error tracking event: %v\n", err)
	} else {
		fmt.Println("Event tracked successfully!")
	}
}
