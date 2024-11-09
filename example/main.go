package main

import (
	"github.com/brycensranch/go-aptabase/pkg/aptabase/v1"
	"log"
)

type Event struct {
	EventName   string
	EventParams map[string]string
}

func main() {
	// Initialize the tracking client
	apiKey := "A-US-your-api-key" // Replace with your actual API key
	appVersion := "1.0.0"
	appBuildNumber := uint64(123)
	debugMode := false

	client := aptabase.NewClient(apiKey, appVersion, appBuildNumber, debugMode, "")

	event := aptabase.EventData{
		EventName: "UserSignUp",
		Props: map[string]interface{}{
			"username": "johndoe",
			"email":    "johndoe@example.com",
		},
	}
	client.TrackEvent(event)
	event2 := aptabase.EventData{
		EventName: "UserSignIn",
		Props: map[string]interface{}{
			"username": "johndoe",
			"email":    "johndoe@example.com",
		},
	}
	client.TrackEvent(event2)
	// You need to flush all the events at the end of your application otherwise they will not be sent.
	client.Stop()
	log.Printf("I am done running the example and I've called stop")
}
