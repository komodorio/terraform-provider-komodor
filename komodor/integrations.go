package komodor

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const IntegrationsUrl string = DefaultEndpoint + "/integrations/kubernetes"

type Kubernetes struct {
	Id string `json:"apiKey"`
}

func (c *Client) GetKubernetesCluster(clusterName string) (*Kubernetes, int, error) {
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, fmt.Sprintf(IntegrationsUrl+"/%s", clusterName), nil)

	if err != nil {
		return nil, statusCode, err
	}

	var kubernetes Kubernetes

	err = json.Unmarshal(res, &kubernetes)
	if err != nil {
		return nil, statusCode, err
	}

	return &kubernetes, statusCode, nil
}

func (c *Client) CreateKubernetesCluster(name string) (*Kubernetes, error) {
	jsonPolicy, err := json.Marshal(map[string]string{"clusterName": name})

	if err != nil {
		return nil, err
	}
	res, _, err := c.executeHttpRequest(http.MethodPost, IntegrationsUrl, &jsonPolicy)

	if err != nil {
		return nil, err
	}

	var kubernetes Kubernetes
	err = json.Unmarshal(res, &kubernetes)
	if err != nil {
		return nil, err
	}

	return &kubernetes, nil
}

func (c *Client) DeleteKubernetesCluster(id string) error {
	_, _, err := c.executeHttpRequest(http.MethodDelete, fmt.Sprintf(IntegrationsUrl+"/%s", id), nil)
	if err != nil {
		return err
	}
	return nil
}
