package openai

import (
	"fmt"
	"time"
)

// Batch provides information about an OpenAI batch, used for processing a large
// collection of asynchronous API requests (e.g. Chat completions).
// As a request, the fields InputFileID, Endpoint, and CompletionWindow are required.
type Batch struct {
	// ID is the batch ID, e.g. "batch-XjGxS3KTG0uNmNOK362iJua3".
	ID string `json:"id,omitempty"`

	// Object is the object type, e.g. "batch".
	Object string `json:"object,omitempty"`

	// InputFileID is the ID of the JSONL input file containing the API requests.
	// The file must have a purpose of "batch", contain a max of 50,000 JSON objects,
	// and be no larger than 100 MB in size. This field is required.
	InputFileID string `json:"input_file_id"`

	// Endpoint is the API endpoint for the batched requests.
	// Example: "/v1/chat/completions". This field is required.
	Endpoint string `json:"endpoint"`

	// CompletionWindow is the time frame within which the batch should be processed.
	// Example: "24h". This field is required.
	CompletionWindow string `json:"completion_window"`

	// Metadata is a map of up to 16 key-value pairs to include with the batch.
	// Keys may be up to 64 characters and values may be up to 512 characters.
	Metadata map[string]string `json:"metadata,omitempty"`

	// Status is the status of the batch: "validating", "failed", "in_progress",
	// "finalizing", "completed", "expired", "cancelling", "cancelled".
	Status string `json:"status,omitempty"`

	// OutputFileID is the ID of the JSONL output file containing the API responses.
	OutputFileID string `json:"output_file_id,omitempty"`

	// ErrorFileID is the ID of the output file containing any errors that occurred.
	ErrorFileID string `json:"error_file_id,omitempty"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at,omitempty"`

	// InProgressAt is a timestamp in epoch seconds when the batch started processing.
	InProgressAt int64 `json:"in_progress_at,omitempty"`

	// ExpiresAt is a timestamp in epoch seconds when the batch will expire.
	ExpiresAt int64 `json:"expires_at,omitempty"`

	// FinalizingAt is a timestamp in epoch seconds when the batch is finalizing.
	FinalizingAt int64 `json:"finalizing_at,omitempty"`

	// CompletedAt is a timestamp in epoch seconds when the batch completed.
	CompletedAt int64 `json:"completed_at,omitempty"`

	// FailedAt is a timestamp in epoch seconds when the batch failed.
	FailedAt int64 `json:"failed_at,omitempty"`

	// ExpiredAt is a timestamp in epoch seconds when the batch expired.
	ExpiredAt int64 `json:"expired_at,omitempty"`

	// CancellingAt is a timestamp in epoch seconds when the batch is cancelling.
	CancellingAt int64 `json:"cancelling_at,omitempty"`

	// CancelledAt is a timestamp in epoch seconds when the batch was cancelled.
	CancelledAt int64 `json:"cancelled_at,omitempty"`

	// RequestCounts provides information about the number of requests in the batch.
	RequestCounts RequestCounts `json:"request_counts,omitempty"`

	// Errors is a list of errors that occurred during processing.
	Errors BatchErrorList `json:"errors,omitempty"`
}

// Duration provides the time elapsed since the batch was created.
func (b *Batch) Duration() time.Duration {
	if b.CreatedAt == 0 {
		return 0
	}
	var maxAt int64
	if b.CompletedAt > maxAt {
		maxAt = b.CompletedAt
	}
	if b.FailedAt > maxAt {
		maxAt = b.FailedAt
	}
	if b.ExpiredAt > maxAt {
		maxAt = b.ExpiredAt
	}
	if b.CancelledAt > maxAt {
		maxAt = b.CancelledAt
	}
	createdAt := time.Unix(b.CreatedAt, 0)
	if maxAt > 0 {
		return time.Unix(maxAt, 0).Sub(createdAt)
	}
	return time.Since(createdAt)
}

// Progress provides information about the progress of a batch.
func (b *Batch) Progress() string {
	return fmt.Sprintf("%s %s, %d total, %d completed, %d failed, %s elapsed", b.ID, b.Status,
		b.RequestCounts.Total, b.RequestCounts.Completed, b.RequestCounts.Failed, b.Duration())
}

// IsDone returns true if the batch has completed, failed, expired, or been cancelled.
func (b *Batch) IsDone() bool {
	switch b.Status {
	case "completed", "failed", "expired", "cancelled":
		return true
	}
	return false
}

// RequestCounts provides information about the number of requests in a batch.
type RequestCounts struct {
	// Total is the total number of requests in the batch.
	Total int `json:"total,omitempty"`

	// Completed is the number of requests that have been completed.
	Completed int `json:"completed,omitempty"`

	// Failed is the number of requests that have failed.
	Failed int `json:"failed,omitempty"`
}

// BatchError provides information about an error that occurred during batch processing.
type BatchError struct {
	// Code is the error code identifying the error type.
	Code string `json:"code,omitempty"`

	// Message is a human-readable description of the error.
	Message string `json:"message,omitempty"`

	// Param is the name of the parameter that caused the error, if applicable.
	Param string `json:"param,omitempty"`

	// Line is the line number in the input file where the error occurred, if applicable.
	Line int `json:"line,omitempty"`
}

// HasError returns true if the BatchError is not empty.
func (e BatchError) HasError() bool {
	return len(e.Code) > 0 || len(e.Message) > 0 || len(e.Param) > 0 || e.Line > 0
}

// Error provides a string representation of the BatchError.
func (e *BatchError) Error() string {
	s := "error " + e.Code
	if len(e.Param) > 0 {
		s += " param " + e.Param
	}
	if e.Line > 0 {
		s += " line " + fmt.Sprint(e.Line)
	}
	return s + ": " + e.Message
}

// BatchErrorList is a list of errors that occurred during batch processing.
type BatchErrorList struct {
	Object string       `json:"object,omitempty"` // "list" is expected
	Data   []BatchError `json:"data,omitempty"`   // list of batch errors
}

// BatchList is a list of batches that belong to the user's organization.
type BatchList struct {
	Object  string  `json:"object"`   // "list" is expected
	Data    []Batch `json:"data"`     // list of batches
	FirstID string  `json:"first_id"` // first batch ID in the collection
	LastID  string  `json:"last_id"`  // use with the "after" query parameter
	HasMore bool    `json:"has_more"` // true if there are more batches to retrieve
}

// BatchRequest contains the just the fields required to create a batch.
type BatchRequest struct {
	// InputFileID is the ID of the JSONL input file containing the API requests.
	// The file must have a purpose of "batch", contain a max of 50,000 JSON objects,
	// and be no larger than 100 MB in size.
	InputFileID string `json:"input_file_id"`

	// Endpoint is the API endpoint for the batched requests.
	// Example: "/v1/chat/completions".
	Endpoint string `json:"endpoint"`

	// CompletionWindow is the time frame within which the batch should be processed.
	// Example: "24h".
	CompletionWindow string `json:"completion_window"`

	// Metadata is a map of up to 16 key-value pairs to include with the batch.
	// Keys may be up to 64 characters and values may be up to 512 characters.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BatchRequestItem contains information about an individual API request in a batch.
// The batch input file will contain lines of these request JSON objects.
// This implementation assumes the request will be a Chat Completion request.
// Note that the API supports other types of requests as well (e.g. embeddings).
type BatchRequestItem struct {
	// CustomID is a developer-provided per-request id that will be used to match outputs to inputs.
	// It must be unique for each request in the batch.
	CustomID string `json:"custom_id"`

	// Method is the HTTP method for the request, e.g. "POST".
	Method string `json:"method"`

	// URL is the relative URL for the request, e.g. "/v1/chat/completions".
	URL string `json:"url"`

	// Body is the HTTP request body to be submitted.
	Body ChatRequest `json:"body"`
}

// BatchResponseItem contains information about an individual API response in a batch.
// The batch output and error files will contain lines of these response JSON objects.
// This implementation assumes the response will be a Chat Completion response.
// Note that the API supports other types of responses as well (e.g. embeddings).
type BatchResponseItem struct {
	// ID is the OpenAI response ID, e.g. "batch_req_6p9XYPYSTTRi0xEviKjjilqrWU2Ve".
	ID string `json:"id"`

	// CustomID is the developer-provided per-request id that was used to match outputs to inputs.
	CustomID string `json:"custom_id"`

	// Response is the API response body content.
	Response BatchItemResponse `json:"response"`

	// Error is the error message if the request failed.
	Error BatchError `json:"error,omitempty"`
}

// HasError returns true if the BatchResponseItem contains an error.
func (r BatchResponseItem) HasError() bool {
	return r.Error.HasError()
}

// Completion provides the first message content from the response.
func (r BatchResponseItem) Completion() string {
	if len(r.Response.Body.Choices) > 0 {
		return r.Response.Body.Choices[0].Message.Content
	}
	return ""
}

// BatchItemResponse contains the HTTP response output for a batch request item.
type BatchItemResponse struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int `json:"status_code"`

	// RequestID is the ID of the request that produced this response.
	RequestID string `json:"request_id"`

	// Body is the response body content.
	Body ChatResponse `json:"body"`
}
