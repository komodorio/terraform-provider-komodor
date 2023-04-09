package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const MonitorsUrl = DefaultEndpoint + "/monitors/config"

type Sensor struct {
	Cluster    string                 `json:"cluster"`
	Namespaces []string               `json:"namespaces"`
	Exclude    map[string]interface{} `json:"exclude"`
	Include    map[string]interface{} `json:"include"`
}

type Sinks struct {
	Slack    []string `json:"slack"`
	Teams    []string `json:"teams"`
	Opsgenie []string `json:"opsgenie"`
	//Pagerduty map[string]interface{} `json:"pagerduty"` // TODO: what is the type
	//Webhook   map[string]interface{} `json:"webhook"`   // TODO: what is the type
}

type SinkOptions struct {
	ShouldSend bool     `json:"shouldSend"`
	NotifyOn   []string `json:"notifyOn"`
}
type NewMonitor struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Active      bool                   `json:"active"`
	Sensors     []Sensor               `json:"sensors"`
	Variables   map[string]interface{} `json:"variables"`
	Sinks       Sinks                  `json:"sinks"`
	SinkOptions SinkOptions            `json:"sinkOptions"`
	IsDeleted   bool                   `json:"isDeleted"`
}

type Monitor struct {
	Id          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Active      bool                   `json:"active"`
	Sensors     []Sensor               `json:"sensors"`
	Variables   map[string]interface{} `json:"variables"`
	Sinks       Sinks                  `json:"sinks"`
	SinkOptions map[string]interface{} `json:"sinkOptions"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"ureatedAt"`
	IsDeleted   bool                   `json:"isDeleted"`
}

func (c *Client) GetMonitors() ([]Monitor, error) {
	res, err := c.executeHttpRequest(http.MethodGet, MonitorsUrl, nil)
	if err != nil {
		return nil, err
	}

	var monitors []Monitor

	err = json.Unmarshal(res, &monitors)
	if err != nil {
		return nil, err
	}

	return monitors, nil
}

func (c *Client) GetMonitor(id string) (*Monitor, error) {
	res, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(MonitorsUrl+"/%s", id), nil)
	if err != nil {
		return nil, err
	}
	var monitor Monitor
	err = json.Unmarshal(res, &monitor)
	if err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (c *Client) UpdateMonitor(id string, m *NewMonitor) (*Monitor, error) {
	jsonMonitor, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	res, err := c.executeHttpRequest(http.MethodPut, fmt.Sprintf(MonitorsUrl+"/%s", id), &jsonMonitor)
	if err != nil {
		return nil, err
	}
	var monitor Monitor
	err = json.Unmarshal(res, &monitor)
	if err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (c *Client) CreateMonitor(m *NewMonitor) (*Monitor, error) {
	jsonMonitor, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	res, err := c.executeHttpRequest(http.MethodPost, MonitorsUrl, &jsonMonitor)

	if err != nil {
		return nil, err
	}
	var monitor Monitor
	err = json.Unmarshal(res, &monitor)
	if err != nil {
		return nil, err
	}

	return &monitor, nil
}
