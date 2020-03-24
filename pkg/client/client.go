package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PagerDuty/pagerduty-agent/pkg/eventsapi"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	HTTPClient    *http.Client
	ServerAddress string

	secret string
}

func NewClient(serverAddress, secret string) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   5 * time.Second,
		},
		ServerAddress: serverAddress,
		secret:        secret,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("token %v", c.secret))
	return c.HTTPClient.Do(req)
}

// Send an event to the agent daemon server.
func (c *Client) Send(event eventsapi.EventV2) (*http.Response, error) {
	url := generateURL(c.ServerAddress, "/send")

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) QueueRetry(routingKey string) (*http.Response, error) {
	url := generateURL(c.ServerAddress, "/queue/retry")
	url.RawQuery = fmt.Sprintf("rk=%v", routingKey)

	req, err := http.NewRequest("POST", url.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) QueueStatus(routingKey string) (*http.Response, error) {
	url := generateURL(c.ServerAddress, "/queue/status")
	url.RawQuery = fmt.Sprintf("rk=%v", routingKey)

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func generateURL(serverAddress, path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   serverAddress,
		Path:   path,
	}
}
