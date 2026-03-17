package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// KnowledgebaseScopedClusters defines the cluster scoping for a knowledge base file.
type KnowledgebaseScopedClusters struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// KnowledgeBaseFile represents a file stored in the Komodor Klaudia Knowledge Base.
type KnowledgeBaseFile struct {
	Id             string                        `json:"id"`
	Name           string                        `json:"name"`
	Size           int64                         `json:"size"`
	Clusters       *KnowledgebaseScopedClusters  `json:"clusters,omitempty"`
	IsBlueprint    bool                          `json:"isBlueprint"`
	UploadedAt     string                        `json:"uploadedAt"`
	CreatedByEmail string                        `json:"createdByEmail"`
}

// KnowledgeBaseListResponse is the API response for listing knowledge base files.
type KnowledgeBaseListResponse struct {
	Files []KnowledgeBaseFile `json:"files"`
}

// KnowledgeBaseDeleteRequest is the request body for deleting knowledge base files.
type KnowledgeBaseDeleteRequest struct {
	FileIDs []string `json:"fileIDs"`
}

// KnowledgeBaseDeleteResponse is the API response for deleting knowledge base files.
type KnowledgeBaseDeleteResponse struct {
	DeletedFiles []string `json:"deletedFiles"`
	FailedFiles  []string `json:"failedFiles"`
}

// ListKnowledgeBaseFiles returns all files in the Knowledge Base.
func (c *Client) ListKnowledgeBaseFiles() ([]KnowledgeBaseFile, int, error) {
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, c.GetKnowledgeBaseUrl(), nil)
	if err != nil {
		return nil, statusCode, err
	}

	var listResp KnowledgeBaseListResponse
	if err := json.Unmarshal(res, &listResp); err != nil {
		return nil, statusCode, err
	}

	return listResp.Files, statusCode, nil
}

// GetKnowledgeBaseFile retrieves a single knowledge base file by its ID.
// Since there is no single-file GET endpoint, this lists all files and filters by ID.
func (c *Client) GetKnowledgeBaseFile(id string) (*KnowledgeBaseFile, int, error) {
	files, statusCode, err := c.ListKnowledgeBaseFiles()
	if err != nil {
		return nil, statusCode, err
	}

	for i := range files {
		if files[i].Id == id {
			return &files[i], http.StatusOK, nil
		}
	}

	return nil, http.StatusNotFound, fmt.Errorf("knowledge base file with id %q not found", id)
}

// UploadKnowledgeBaseFile uploads a single file to the Klaudia Knowledge Base.
func (c *Client) UploadKnowledgeBaseFile(filename string, content []byte, clusters *KnowledgebaseScopedClusters) (*KnowledgeBaseFile, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file content as a form file field
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file part: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	// Set custom filename for index 0
	if err := writer.WriteField("files[0][filename]", filename); err != nil {
		return nil, fmt.Errorf("failed to write filename field: %w", err)
	}

	// Add cluster scoping fields if provided
	if clusters != nil {
		for _, inc := range clusters.Include {
			if err := writer.WriteField("files[0][clusters][include][]", inc); err != nil {
				return nil, fmt.Errorf("failed to write cluster include field: %w", err)
			}
		}
		for _, exc := range clusters.Exclude {
			if err := writer.WriteField("files[0][clusters][exclude][]", exc); err != nil {
				return nil, fmt.Errorf("failed to write cluster exclude field: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	res, _, err := c.executeMultipartRequest(http.MethodPost, c.GetKnowledgeBaseUrl(), body, writer.FormDataContentType())
	if err != nil {
		return nil, fmt.Errorf("failed to upload knowledge base file: %w", err)
	}

	var listResp KnowledgeBaseListResponse
	if err := json.Unmarshal(res, &listResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}

	// Find the uploaded file by name in the response
	for i := range listResp.Files {
		if listResp.Files[i].Name == filename {
			return &listResp.Files[i], nil
		}
	}

	// Fallback: return the first file if only one was uploaded
	if len(listResp.Files) > 0 {
		return &listResp.Files[0], nil
	}

	return nil, fmt.Errorf("uploaded file not found in response")
}

// DeleteKnowledgeBaseFiles deletes one or more knowledge base files by their IDs.
func (c *Client) DeleteKnowledgeBaseFiles(ids []string) (*KnowledgeBaseDeleteResponse, error) {
	reqBody := &KnowledgeBaseDeleteRequest{FileIDs: ids}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodDelete, c.GetKnowledgeBaseUrl(), &body)
	if err != nil {
		return nil, err
	}

	var deleteResp KnowledgeBaseDeleteResponse
	if err := json.Unmarshal(res, &deleteResp); err != nil {
		return nil, err
	}

	return &deleteResp, nil
}
