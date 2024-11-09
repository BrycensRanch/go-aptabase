package aptabase

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	APIKey         string
	BaseURL        string
	HTTPClient     *http.Client
	SessionID      string
	LastTouch      time.Time
	SessionTimeout time.Duration
	eventChan      chan EventData
	AppVersion     string
	AppBuildNumber uint64
	DebugMode      bool
	quitChan       chan struct{}
	wg             sync.WaitGroup
	Quit           bool
	Logger         *log.Logger // Logger field added
	batch          []EventData
}

// NewClient Initializes a new client and begins processing events automagically.
func NewClient(apiKey, appVersion string, appBuildNumber uint64, debugMode bool, baseURL string) *Client {

	client := &Client{
		APIKey:         apiKey,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		SessionTimeout: 1 * time.Hour,
		eventChan:      make(chan EventData, 100),
		AppVersion:     appVersion,
		AppBuildNumber: appBuildNumber,
		DebugMode:      debugMode,
		quitChan:       make(chan struct{}),
		Quit:           false,
		Logger:         log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		batch:          make([]EventData, 0, 999),
	}

	client.BaseURL = client.determineHost(apiKey)
	// Blackhole
	if strings.Contains(client.APIKey, "SH") {
		client.BaseURL = baseURL
	}
	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()
	client.Logger.Printf("Aptabase Go is ready to go! SDK Version: %s", GetVersion())
	client.Logger.Printf("NewClient created with APIKey=%s, BaseURL=%s, SessionID=%s", client.APIKey, client.BaseURL, client.SessionID)
	go client.processQueue()

	return client
}
