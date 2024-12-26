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
	MaxBatchSize   int
}

type Builder struct {
	APIKey         string
	AppVersion     string
	AppBuildNumber uint64
	DebugMode      bool
	MaxBatchSize   int
	BaseURL        string
}

// NewClient Initializes a new client and begins processing events automagically.
func NewClient(data Builder) *Client {

	client := &Client{
		APIKey:         data.APIKey,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		SessionTimeout: 1 * time.Hour,
		eventChan:      make(chan EventData, 100),
		AppVersion:     data.AppVersion,
		AppBuildNumber: data.AppBuildNumber,
		DebugMode:      data.DebugMode,
		quitChan:       make(chan struct{}),
		Quit:           false,
		Logger:         log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		batch:          make([]EventData, 0, 999),
		MaxBatchSize:   data.MaxBatchSize,
	}

	client.BaseURL = client.determineHost(data.APIKey)
	if strings.Contains(client.APIKey, "SH") {
		client.BaseURL = data.BaseURL
	}
	client.SessionID = client.NewSessionID()
	client.LastTouch = time.Now().UTC()
	client.Logger.Printf("Aptabase Go is ready to go! SDK Version: %s", GetVersion())
	client.Logger.Printf("NewClient created with APIKey=%s, BaseURL=%s, SessionID=%s", client.APIKey, client.BaseURL, client.SessionID)
	go client.processQueue()

	return client
}
