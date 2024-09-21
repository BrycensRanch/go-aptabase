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

type Session struct {
	SessionID string
	StartTime time.Time
	EndTime   time.Time
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
	
	// Track the event
	err := client.TrackEvent(event.EventName, event.EventParams)
	if err != nil {
		fmt.Printf("Error tracking event: %v\n", err)
	} else {
		fmt.Println("Event tracked successfully!")
	}

	// Create a session to track
	session := Session{
		SessionID: "session_001",
		StartTime: time.Now().Add(-30 * time.Minute),
		EndTime:   time.Now(),
	}

	// Track the session (assuming you have a TrackSession method)
	err = client.TrackSession(session.SessionID, session.StartTime, session.EndTime)
	if err != nil {
		fmt.Printf("Error tracking session: %v\n", err)
	} else {
		fmt.Println("Session tracked successfully!")
	}
}
