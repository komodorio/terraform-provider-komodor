package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const MonitorsUrl = DefaultEndpoint + "/monitors/config"

type (
	ModelWorkflowConfigurationSensorFilters struct {
		Namespaces  []string `json:"namespaces,omitempty"`
		Annotations []string `json:"annotations,omitempty"`
		Labels      []string `json:"labels,omitempty"`
	}
	Sensor struct {
		Cluster string `json:"cluster"`
		ModelWorkflowConfigurationSensorFilters
		Exclude ModelWorkflowConfigurationSensorFilters `json:"exclude,omitempty"`
	}

	SinkOptions struct {
		ShouldSend *bool    `json:"shouldSend,omitempty"`
		NotifyOn   []string `json:"notifyOn,omitempty"`
	}

	PagerDutyModel struct {
		Channel              string `json:"channel"`
		IntegrationKey       string `json:"integrationKey"`
		PagerDutyAccountName string `json:"pagerDutyAccountName"`
	}

	Sinks struct {
		Slack          []string         `json:"slack,omitempty"`
		Teams          []string         `json:"teams,omitempty"`
		Opsgenie       []string         `json:"opsgenie,omitempty"`
		Pagerduty      []PagerDutyModel `json:"pagerduty,omitempty"`
		GenericWebhook []string         `json:"genericWebhook,omitempty"`
	}

	ModelWorkflowConfigurationVariables struct {
		MinDuration           *int      `json:"duration,omitempty"`
		MinAvailable          *string   `json:"minAvailable,omitempty"`
		CronJobCondition      *string   `json:"cronJobCondition,omitempty"`
		ResolveAfter          *int      `json:"resolveAfter,omitempty"`
		IgnoreAfter           *int      `json:"ignoreAfter,omitempty"`
		Reasons               *[]string `json:"reasons,omitempty"`
		NodeCreationThreshold *string   `json:"nodeCreationThreshold,omitempty"`
	}
)

type NewMonitor struct {
	Name        string                              `json:"name"`
	Type        string                              `json:"type"`
	Active      bool                                `json:"active"`
	Sensors     []Sensor                            `json:"sensors"`
	Variables   ModelWorkflowConfigurationVariables `json:"variables,omitempty"`
	Sinks       Sinks                               `json:"sinks,omitempty"`
	SinkOptions SinkOptions                         `json:"sinkOptions,omitempty"`
	IsDeleted   bool                                `json:"isDeleted"`
}

type Monitor struct {
	Id          string                              `json:"id"`
	Name        string                              `json:"name"`
	Type        string                              `json:"type"`
	Active      bool                                `json:"active"`
	Sensors     []Sensor                            `json:"sensors"`
	Variables   ModelWorkflowConfigurationVariables `json:"variables,omitempty"`
	Sinks       Sinks                               `json:"sinks,omitempty"`
	SinkOptions SinkOptions                         `json:"sinkOptions,omitempty"`
	CreatedAt   string                              `json:"createdAt"`
	UpdatedAt   string                              `json:"ureatedAt"`
	IsDeleted   bool                                `json:"isDeleted"`
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
