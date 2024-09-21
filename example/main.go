package main

import (
	"fmt"
	"time"
	"github.com/brycensranch/go-aptabase/pkg/aptabase/v1"
)

func main() {
	client := NewClient("your-api-key")

	event := Event{
		EventName: "user_login",
		EventParams: map[string]string{
			"user_id": "12345",
			"device":  "mobile",
		},
		Timestamp: time.Now(),
	}

	err := client.TrackEvent(event)
	if err != nil {
		fmt.Printf("Error tracking event: %v\n", err)
	} else {
		fmt.Println("Event tracked successfully!")
	}

	session := Session{
		SessionID: "session_001",
		StartTime: time.Now().Add(-30 * time.Minute),
		EndTime:   time.Now(),
	}

	err = client.TrackSession(session)
	if err != nil {
		fmt.Printf("Error tracking session: %v\n", err)
	} else {
		fmt.Println("Session tracked successfully!")
	}
}