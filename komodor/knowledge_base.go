package komodor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// KnowledgeBaseScopedClusters defines the cluster scoping for a knowledge base file.
type KnowledgeBaseScopedClusters struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// KnowledgeBaseFile represents a file stored in the Komodor Klaudia Knowledge Base.
type KnowledgeBaseFile struct {
	Id             string                       `json:"id"`
	Name           string                       `json:"name"`
	Size           int64                        `json:"size"`
	Clusters       *KnowledgeBaseScopedClusters `json:"clusters,omitempty"`
	UploadedAt     string                       `json:"uploadedAt"`
	CreatedByEmail string                       `json:"createdByEmail"`
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

// ListKnowledgeBaseFiles returns all files for the given file type ("knowledge-base" or "blueprint").
func (c *Client) ListKnowledgeBaseFiles(fileType string) ([]KnowledgeBaseFile, int, error) {
	res, statusCode, err := c.executeHttpRequest(http.MethodGet, c.GetKlaudiaFilesUrl(fileType), nil)
	if err != nil {
		return nil, statusCode, err
	}

	var listResp KnowledgeBaseListResponse
	if err := json.Unmarshal(res, &listResp); err != nil {
		return nil, statusCode, err
	}

	return listResp.Files, statusCode, nil
}

// GetKnowledgeBaseFile retrieves a single file by its ID and type.
// Since there is no single-file GET endpoint, this lists all files and filters by ID.
func (c *Client) GetKnowledgeBaseFile(id string, fileType string) (*KnowledgeBaseFile, int, error) {
	files, statusCode, err := c.ListKnowledgeBaseFiles(fileType)
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

// UploadKnowledgeBaseFile uploads a single file of the given type ("knowledge-base" or "blueprint").
func (c *Client) UploadKnowledgeBaseFile(filename string, content []byte, clusters *KnowledgeBaseScopedClusters, fileType string) (*KnowledgeBaseFile, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file content as a form file field; the filename is set in the
	// Content-Disposition header of the part and used by the API as the file name.
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file part: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	// Add cluster scoping fields if provided. The API expects a top-level
	// "clusters" array (index-aligned with the "files" array).
	if clusters != nil {
		for _, inc := range clusters.Include {
			if err := writer.WriteField("clusters[0][include][]", inc); err != nil {
				return nil, fmt.Errorf("failed to write cluster include field: %w", err)
			}
		}
		for _, exc := range clusters.Exclude {
			if err := writer.WriteField("clusters[0][exclude][]", exc); err != nil {
				return nil, fmt.Errorf("failed to write cluster exclude field: %w", err)
			}
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	res, _, err := c.executeMultipartRequest(http.MethodPost, c.GetKlaudiaFilesUrl(fileType), body, writer.FormDataContentType())
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

// DeleteKnowledgeBaseFiles deletes one or more files of the given type by their IDs.
func (c *Client) DeleteKnowledgeBaseFiles(ids []string, fileType string) (*KnowledgeBaseDeleteResponse, error) {
	reqBody := &KnowledgeBaseDeleteRequest{FileIDs: ids}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	res, _, err := c.executeHttpRequest(http.MethodDelete, c.GetKlaudiaFilesUrl(fileType), &body)
	if err != nil {
		return nil, err
	}

	// Some APIs return 204 No Content or an empty body for successful deletes.
	if len(res) == 0 {
		return &KnowledgeBaseDeleteResponse{}, nil
	}

	var deleteResp KnowledgeBaseDeleteResponse
	if err := json.Unmarshal(res, &deleteResp); err != nil {
		return nil, err
	}

	return &deleteResp, nil
}
