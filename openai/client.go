package openai

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"slices"
	"time"
)

// Client is the OpenAI API client.
type Client struct {
	OrgID   string
	APIKey  string
	BaseURL string
	client  *http.Client
}

// NewClient instantiates a new OpenAI API client. If either orgID or apiKey
// are not provided, the environment variables OPENAI_ORG_ID and OPENAI_API_KEY
// will be used, respectively.
func NewClient(orgID, apiKey string) *Client {
	if orgID == "" {
		orgID = os.Getenv("OPENAI_ORG_ID")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &Client{
		OrgID:   orgID,
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// newRequest creates a new HTTP request with the required headers.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create %s %s request: %w", method, path, err)
	}
	req.Header.Add("Accept", "application/json")
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Add("Authorization", "Bearer "+c.APIKey)
	}
	if c.OrgID != "" {
		req.Header.Add("OpenAI-Organization", c.OrgID)
	}
	return req, nil
}

// getRequest creates a new HTTP request with the required headers.
func (c *Client) getRequest(ctx context.Context, path string) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodGet, path, nil)
}

// postRequest creates a new HTTP request with the required headers.
func (c *Client) postRequest(ctx context.Context, path string, body io.Reader) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodPost, path, body)
}

// deleteRequest creates a new HTTP request with the required headers.
func (c *Client) deleteRequest(ctx context.Context, path string) (*http.Request, error) {
	return c.newRequest(ctx, http.MethodDelete, path, nil)
}

// sendRequest sends the provided HTTP request and returns the response body.
func (c *Client) sendRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		e := fmt.Errorf("send request %s: %w", req.URL.Path, err)
		if resp != nil {
			return nil, RequestError{Code: resp.StatusCode, Err: e}
		}
		return nil, e
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, RequestError{
			Code: resp.StatusCode,
			Err:  fmt.Errorf("read response body %s: %w", req.URL.Path, err),
		}
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		var er ErrorResponse
		if e := json.Unmarshal(body, &er); e == nil && er.Error != nil {
			return body, RequestError{
				Code: resp.StatusCode,
				Err:  er.Error,
			}
		}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return body, RequestError{
			Code: resp.StatusCode,
			Err:  fmt.Errorf("%s: %s", resp.Status, req.URL.Path),
		}
	}
	return body, nil
}

// ListModelsRaw lists the currently available models, and provides basic information
// about each one such as the owner and availability. It returns the raw JSON response.
func (c *Client) ListModelsRaw(ctx context.Context) ([]byte, error) {
	req, err := c.getRequest(ctx, "/models")
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return body, nil
}

// ListModels lists the currently available models, and provides basic information
// about each one such as the owner and availability.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	// Fetch the raw JSON response:
	body, err := c.ListModelsRaw(ctx)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of models:
	var list ModelList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list models: unmarshal response: %w", err)
	}
	models := list.Data
	// Sort the models chronologically and return:
	slices.SortFunc(models, func(a, b Model) int { return cmp.Compare(a.CreatedAt, b.CreatedAt) })
	return models, nil
}

// ReadModelRaw reads the details of the specified model. It returns the raw JSON response.
func (c *Client) ReadModelRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/models/"+id)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read model %s: %w", id, err)
	}
	return body, nil
}

// ReadModel reads the details of the specified model.
func (c *Client) ReadModel(ctx context.Context, id string) (Model, error) {
	var model Model
	body, err := c.ReadModelRaw(ctx, id)
	if err != nil {
		return model, err
	}
	if err := json.Unmarshal(body, &model); err != nil {
		return model, fmt.Errorf("read model %s: unmarshal response: %w", id, err)
	}
	return model, nil
}

// ValidModel returns true if the specified model ID is valid.
func (c *Client) ValidModel(ctx context.Context, id string) bool {
	if CommonModels[id] {
		return true
	}
	model, err := c.ReadModel(ctx, id)
	if err != nil {
		return false
	}
	CommonModels[model.ID] = true
	return model.ID == id
}

// DeleteModelRaw deletes the specified model. It returns the raw JSON response.
func (c *Client) DeleteModelRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.deleteRequest(ctx, "/models/"+id)
	if err != nil {
		return nil, fmt.Errorf("delete model %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return body, fmt.Errorf("delete model %s: %w", id, err)
	}
	return body, nil
}

// DeleteModel deletes the specified model.
func (c *Client) DeleteModel(ctx context.Context, id string) error {
	_, err := c.DeleteModelRaw(ctx, id)
	return err
}

// UploadFile uploads a jsonl file for use with subsequent fine-tuning requests.
func (c *Client) UploadFile(ctx context.Context, fileName, purpose string, data []byte) (File, error) {
	var file File

	// Create the multipart writer
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// File Purpose: usually "fine-tune"
	if purpose == "" {
		purpose = "fine-tune"
	}
	err := w.WriteField("purpose", purpose)
	if err != nil {
		return file, fmt.Errorf("upload file: field purpose: %w", err)
	}

	// File Name and Data
	var fw io.Writer
	fw, err = w.CreateFormFile("file", fileName)
	if err != nil {
		return file, fmt.Errorf("upload file: field file: %w", err)
	}
	_, err = io.Copy(fw, bytes.NewReader(data))
	if err != nil {
		return file, fmt.Errorf("upload file: field file: %w", err)
	}
	w.Close()

	// Create the request
	req, err := c.postRequest(ctx, "/files", &buf)
	if err != nil {
		return file, fmt.Errorf("upload file: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Send the request
	body, err := c.sendRequest(req)
	if err != nil {
		return file, fmt.Errorf("upload file: send request: %w", err)
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return file, fmt.Errorf("upload file: unmarshal response: %w", err)
	}
	return file, nil
}

// ListFilesRaw lists the organization's files, providing basic information about each one.
// If a purpose is provided, it will filter the list to only include files with that purpose.
// It returns the raw JSON response.
func (c *Client) ListFilesRaw(ctx context.Context, purpose string) ([]byte, error) {
	var param string
	if purpose != "" {
		param = "?purpose=" + purpose
	}
	req, err := c.getRequest(ctx, "/files"+param)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return body, nil
}

// ListFiles lists the organization's files, providing basic information about each one.
// If a purpose is provided, it will filter the list to only include files with that purpose.
func (c *Client) ListFiles(ctx context.Context, purpose string) ([]File, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFilesRaw(ctx, purpose)
	if err != nil {
		return nil, err
	}
	// Unmarshal the JSON response into a list of files:
	var list FileList
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("list files: unmarshal response: %w", err)
	}
	files := list.Data
	// Sort the files by name and return:
	slices.SortFunc(files, func(a, b File) int { return cmp.Compare(a.FileName, b.FileName) })
	return files, nil
}

// ReadFileRaw reads the metatdata detail of the specified file. It returns the raw JSON response.
func (c *Client) ReadFileRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+id)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", id, err)
	}
	return body, nil
}

// ReadFile reads the metadata detail of the specified file.
func (c *Client) ReadFile(ctx context.Context, id string) (File, error) {
	var file File
	body, err := c.ReadFileRaw(ctx, id)
	if err != nil {
		return file, err
	}
	if err := json.Unmarshal(body, &file); err != nil {
		return file, fmt.Errorf("read file %s: unmarshal response: %w", id, err)
	}
	return file, nil
}

// DownloadFile reads the contents of the specified file.
func (c *Client) DownloadFile(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/files/"+id+"/content")
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("download file %s: %w", id, err)
	}
	return body, nil
}

// DeleteFileRaw deletes the specified file. It returns the raw JSON response.
func (c *Client) DeleteFileRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.deleteRequest(ctx, "/files/"+id)
	if err != nil {
		return nil, fmt.Errorf("delete file %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return body, fmt.Errorf("delete file %s: %w", id, err)
	}
	return body, nil
}

// DeleteFile deletes the specified file.
func (c *Client) DeleteFile(ctx context.Context, id string) error {
	_, err := c.DeleteFileRaw(ctx, id)
	return err
}

// CreateBatchRaw creates a new batch job. It returns the raw JSON response.
func (c *Client) CreateBatchRaw(ctx context.Context, req BatchRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("create batch job: %w", err)
	}
	httpReq, err := c.postRequest(ctx, "/batches", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create batch job: %w", err)
	}
	raw, err := c.sendRequest(httpReq)
	if err != nil {
		return raw, fmt.Errorf("create batch job: %w", err)
	}
	return raw, nil

}

// CreateBatch creates a new batch job.
func (c *Client) CreateBatch(ctx context.Context, req BatchRequest) (Batch, error) {
	var batch Batch
	body, err := c.CreateBatchRaw(ctx, req)
	if err != nil {
		return batch, err
	}
	if err := json.Unmarshal(body, &batch); err != nil {
		return batch, fmt.Errorf("create batch job: unmarshal response: %w", err)
	}
	return batch, nil
}

// ListBatchesRaw lists the currently available batch jobs, and provides basic information
// about each one, including job status. It returns the raw JSON response.
func (c *Client) ListBatchesRaw(ctx context.Context, limit int, after string) ([]byte, error) {
	if limit < 1 {
		limit = 100
	}
	path := fmt.Sprintf("/batches?limit=%d", limit)
	if after != "" {
		path += "&after=" + after
	}
	req, err := c.getRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list batch jobs: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list batch jobs: %w", err)
	}
	return body, nil
}

// ListBatches lists the currently available batch jobs, and provides basic information
// about each one, including job status.
func (c *Client) ListBatches(ctx context.Context, limit int, after string) ([]Batch, bool, string, error) {
	var list BatchList

	// Fetch the raw JSON response:
	body, err := c.ListBatchesRaw(ctx, limit, after)
	if err != nil {
		return list.Data, list.HasMore, list.LastID, err
	}
	// Unmarshal the JSON response into a list of batches:
	if err := json.Unmarshal(body, &list); err != nil {
		return list.Data, list.HasMore, list.LastID, fmt.Errorf("list batch jobs: unmarshal response: %w", err)
	}
	return list.Data, list.HasMore, list.LastID, nil
}

// ReadBatchRaw reads the metatdata detail of the specified batch job. It returns the raw JSON response.
func (c *Client) ReadBatchRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/batches/"+id)
	if err != nil {
		return nil, fmt.Errorf("read batch job %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read batch job %s: %w", id, err)
	}
	return body, nil
}

// ReadBatch reads the metadata detail of the specified batch job.
func (c *Client) ReadBatch(ctx context.Context, id string) (Batch, error) {
	var batch Batch
	body, err := c.ReadBatchRaw(ctx, id)
	if err != nil {
		return batch, err
	}
	if err := json.Unmarshal(body, &batch); err != nil {
		return batch, fmt.Errorf("read batch job %s: unmarshal response: %w", id, err)
	}
	return batch, nil
}

// ReadBatchResponses reads the results of the specified batch job.
func (c *Client) ReadBatchResponses(ctx context.Context, id string) (Batch, map[string]BatchResponseItem, error) {
	// Read the batch and verify that results are available:
	b, err := c.ReadBatch(ctx, id)
	if err != nil {
		return b, nil, err
	}
	if b.OutputFileID == "" && b.ErrorFileID == "" {
		return b, nil, fmt.Errorf("batch %s status %s has no results", b.ID, b.Status)
	}
	responses := make(map[string]BatchResponseItem, b.RequestCounts.Total)

	// Download the batch output results and read JSONL data:
	if b.OutputFileID != "" {
		outputBytes, err := c.DownloadFile(ctx, b.OutputFileID)
		if err != nil {
			return b, responses, fmt.Errorf("download batch output file %s: %w", b.OutputFileID, err)
		}
		var line int
		scanner := bufio.NewScanner(bytes.NewReader(outputBytes))
		for scanner.Scan() {
			line++
			var item BatchResponseItem
			if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
				return b, responses, fmt.Errorf("unmarshal batch output item %d: %w", line, err)
			}
			responses[item.CustomID] = item
		}
	}

	// Download the batch error results and read JSONL data:
	if b.ErrorFileID != "" {
		errorBytes, err := c.DownloadFile(ctx, b.ErrorFileID)
		if err != nil {
			return b, responses, fmt.Errorf("download batch error file %s: %w", b.ErrorFileID, err)
		}
		var line int
		scanner := bufio.NewScanner(bytes.NewReader(errorBytes))
		for scanner.Scan() {
			line++
			var item BatchResponseItem
			if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
				return b, responses, fmt.Errorf("unmarshal batch output item %d: %w", line, err)
			}
			responses[item.CustomID] = item
		}
	}

	return b, responses, nil
}

// CancelBatchRaw cancels the specified batch job. It returns the raw JSON response.
func (c *Client) CancelBatchRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.postRequest(ctx, "/batches/"+id+"/cancel", nil)
	if err != nil {
		return nil, fmt.Errorf("cancel batch job %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return body, fmt.Errorf("cancel batch job %s: %w", id, err)
	}
	return body, nil
}

// CancelBatch cancels the specified batch job.
func (c *Client) CancelBatch(ctx context.Context, id string) (Batch, error) {
	var batch Batch
	raw, err := c.CancelBatchRaw(ctx, id)
	if err != nil {
		return batch, err
	}
	if err := json.Unmarshal(raw, &batch); err != nil {
		return batch, fmt.Errorf("cancel batch job %s: unmarshal response: %w", id, err)
	}
	return batch, nil
}

// CreateFineTuneRaw creates a new fine-tuned model. It returns the raw JSON response.
func (c *Client) CreateFineTuneRaw(ctx context.Context, req FineTuneRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("create fine-tuning job: %w", err)
	}
	httpReq, err := c.postRequest(ctx, "/fine_tuning/jobs", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create fine-tuning job: %w", err)
	}
	raw, err := c.sendRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("create fine-tuning job: %w", err)
	}
	return raw, nil
}

// CreateFineTune creates a new fine-tuned model. At a minimum, we should provide
// the base model ID, the training file ID, and a suffix for the new model name.
func (c *Client) CreateFineTune(ctx context.Context, req FineTuneRequest) (FineTuneJob, error) {
	var job FineTuneJob
	body, err := c.CreateFineTuneRaw(ctx, req)
	if err != nil {
		return job, err
	}
	if err := json.Unmarshal(body, &job); err != nil {
		return job, fmt.Errorf("create fine-tuning job: unmarshal response: %w", err)
	}
	return job, nil
}

// ListFineTunesRaw lists the currently available fine-tuning jobs, and provides basic information
// about each one, including job status events. It returns the raw JSON response.
func (c *Client) ListFineTunesRaw(ctx context.Context, limit int, after string) ([]byte, error) {
	if limit < 1 {
		limit = 20
	}
	path := fmt.Sprintf("/fine_tuning/jobs?limit=%d", limit)
	if after != "" {
		path += "&after=" + after
	}
	req, err := c.getRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list fine-tuning jobs: %w", err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list fine-tuning jobs: %w", err)
	}
	return body, nil
}

// ListFineTunes lists the currently available fine-tuning jobs, and provides basic information
// about each one, including job status events.
func (c *Client) ListFineTunes(ctx context.Context, limit int, after string) ([]FineTuneJob, bool, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFineTunesRaw(ctx, limit, after)
	if err != nil {
		return nil, false, err
	}
	// Unmarshal the JSON response into a list of fine-tunes:
	var list FineTuneJobs
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, false, fmt.Errorf("list fine-tuning jobs: unmarshal response: %w", err)
	}
	return list.Data, list.HasMore, nil
}

// ReadFineTuneRaw reads the metatdata detail of the specified fine-tuning job. It returns the raw JSON response.
func (c *Client) ReadFineTuneRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.getRequest(ctx, "/fine_tuning/jobs/"+id)
	if err != nil {
		return nil, fmt.Errorf("read fine-tuning job %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("read fine-tuning job %s: %w", id, err)
	}
	return body, nil
}

// ReadFineTune reads the metadata detail of the specified fine-tuning job.
func (c *Client) ReadFineTune(ctx context.Context, id string) (FineTuneJob, error) {
	var fineTune FineTuneJob
	body, err := c.ReadFineTuneRaw(ctx, id)
	if err != nil {
		return fineTune, err
	}
	if err := json.Unmarshal(body, &fineTune); err != nil {
		return fineTune, fmt.Errorf("read fine-tuning job %s: unmarshal response: %w", id, err)
	}
	return fineTune, nil
}

// ListFineTuneEventsRaw lists the events for the specified fine-tuning job. It returns the raw JSON response.
// TODO: support pagination parameters (after, limit)
func (c *Client) ListFineTuneEventsRaw(ctx context.Context, id string, limit int, after string) ([]byte, error) {
	if limit < 1 {
		limit = 20
	}
	path := fmt.Sprintf("/fine_tuning/jobs/%s/events?limit=%d", id, limit)
	if after != "" {
		path += "&after=" + after
	}
	req, err := c.getRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("list fine-tuning job %s events: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return nil, fmt.Errorf("list fine-tuning job %s events: %w", id, err)
	}
	return body, nil
}

// ListFineTuneEvents lists the events for the specified fine-tuning job.
func (c *Client) ListFineTuneEvents(ctx context.Context, id string, limit int, after string) ([]FineTuneEvent, bool, error) {
	// Fetch the raw JSON response:
	body, err := c.ListFineTuneEventsRaw(ctx, id, limit, after)
	if err != nil {
		return nil, false, err
	}
	// Unmarshal the JSON response into a list of fine-tune events:
	var list FineTuneEvents
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, false, fmt.Errorf("list fine-tuning job %s events: unmarshal response: %w", id, err)
	}
	return list.Data, list.HasMore, nil
}

// CancelFineTuneRaw cancels the specified fine-tuning job. It returns the raw JSON response.
func (c *Client) CancelFineTuneRaw(ctx context.Context, id string) ([]byte, error) {
	req, err := c.postRequest(ctx, "/fine_tuning/jobs/"+id+"/cancel", nil)
	if err != nil {
		return nil, fmt.Errorf("cancel fine-tuning job %s: %w", id, err)
	}
	body, err := c.sendRequest(req)
	if err != nil {
		return body, fmt.Errorf("cancel fine-tuning job %s: %w", id, err)
	}
	return body, nil
}

// CancelFineTune cancels the specified fine-tuning job.
func (c *Client) CancelFineTune(ctx context.Context, id string) (FineTuneJob, error) {
	var job FineTuneJob
	raw, err := c.CancelFineTuneRaw(ctx, id)
	if err != nil {
		return job, err
	}
	if err := json.Unmarshal(raw, &job); err != nil {
		return job, fmt.Errorf("cancel fine-tuning job %s: unmarshal response: %w", id, err)
	}
	return job, nil
}

// CompleteChatRaw creates a new chat completion. It returns the raw JSON response.
func (c *Client) CompleteChatRaw(ctx context.Context, req ChatRequest) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("complete chat: %w", err)
	}
	httpReq, err := c.postRequest(ctx, "/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("complete chat: %w", err)
	}
	raw, err := c.sendRequest(httpReq)
	if err != nil {
		return raw, fmt.Errorf("complete chat: %w", err)
	}
	return raw, nil
}

// CompleteChat creates a new chat completion.
func (c *Client) CompleteChat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	var chat ChatResponse
	raw, err := c.CompleteChatRaw(ctx, req)
	if err != nil {
		return chat, err
	}
	if err := json.Unmarshal(raw, &chat); err != nil {
		return chat, fmt.Errorf("complete chat: unmarshal response: %w", err)
	}
	return chat, nil
}
