package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type KlaudiaFileClusters struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

type KlaudiaFile struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Size           int64                `json:"size"`
	UploadedAt     string               `json:"uploadedAt"`
	CreatedByEmail string               `json:"createdByEmail"`
	Clusters       *KlaudiaFileClusters `json:"clusters,omitempty"`
}

type KlaudiaFileListResponse struct {
	Files []KlaudiaFile `json:"files"`
}

type klaudiaFileDeleteRequest struct {
	FileIDs []string `json:"fileIDs"`
}

type KlaudiaFileDeleteResponse struct {
	DeletedFiles []string `json:"deletedFiles"`
	FailedFiles  []string `json:"failedFiles"`
}

func (c *Client) ListKlaudiaFiles(fileType string) (*KlaudiaFileListResponse, int, error) {
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, c.GetKlaudiaFilesUrl(fileType), nil)
	if err != nil {
		return nil, statusCode, err
	}

	var files KlaudiaFileListResponse
	if err := json.Unmarshal(res, &files); err != nil {
		return nil, statusCode, err
	}
	return &files, statusCode, nil
}

func (c *Client) UploadKlaudiaFile(fileType string, file klaudiaFilePayload, clusters *KlaudiaFileClusters) (*KlaudiaFileListResponse, error) {
	body, contentType, err := buildKlaudiaFileMultipartBody([]klaudiaFilePayload{file}, "files", clusters, true)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeMultipartRequest(http.MethodPost, c.GetKlaudiaFilesUrl(fileType), body, contentType)
	if err != nil {
		return nil, err
	}

	var files KlaudiaFileListResponse
	if err := json.Unmarshal(res, &files); err != nil {
		return nil, err
	}
	return &files, nil
}

func (c *Client) UpdateKlaudiaFile(fileType string, fileID string, file *klaudiaFilePayload, clusters *KlaudiaFileClusters) (*KlaudiaFile, int, error) {
	var files []klaudiaFilePayload
	if file != nil {
		files = append(files, *file)
	}

	body, contentType, err := buildKlaudiaFileMultipartBody(files, "file", clusters, false)
	if err != nil {
		return nil, 0, err
	}

	res, statusCode, err := c.executeMultipartRequest(http.MethodPut, fmt.Sprintf("%s/%s", c.GetKlaudiaFilesUrl(fileType), fileID), body, contentType)
	if err != nil {
		return nil, statusCode, err
	}

	var updated KlaudiaFile
	if err := json.Unmarshal(res, &updated); err != nil {
		return nil, statusCode, err
	}
	return &updated, statusCode, nil
}

func (c *Client) DeleteKlaudiaFile(fileType string, fileID string) (*KlaudiaFileDeleteResponse, error) {
	req := klaudiaFileDeleteRequest{FileIDs: []string{fileID}}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodDelete, c.GetKlaudiaFilesUrl(fileType), &body)
	if err != nil {
		return nil, err
	}

	var deleted KlaudiaFileDeleteResponse
	if err := json.Unmarshal(res, &deleted); err != nil {
		return nil, err
	}
	return &deleted, nil
}

type klaudiaFilePayload struct {
	Filename string
	Content  []byte
}

func buildKlaudiaFileMultipartBody(files []klaudiaFilePayload, fileFieldName string, clusters *KlaudiaFileClusters, clustersAsArray bool) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range files {
		part, err := writer.CreateFormFile(fileFieldName, file.Filename)
		if err != nil {
			return nil, "", err
		}
		if _, err := io.Copy(part, bytes.NewReader(file.Content)); err != nil {
			return nil, "", err
		}
	}

	if clusters != nil {
		var clustersValue interface{} = clusters
		if clustersAsArray {
			clustersValue = []KlaudiaFileClusters{*clusters}
		}
		clustersJSON, err := json.Marshal(clustersValue)
		if err != nil {
			return nil, "", err
		}
		if err := writer.WriteField("clusters", string(clustersJSON)); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func (c *Client) executeMultipartRequest(method string, url string, body *bytes.Buffer, contentType string) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "Terraform (terraform-provider-komodor); Go-http-client/1.1")

	return c.executeWithRetry(req, 3, 5*time.Second)
}
