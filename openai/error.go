package openai

import (
	"errors"
	"fmt"
)

// ErrorResponse is the error response returned by the OpenAI API.
type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

// APIError is an error returned by the OpenAI API.
type APIError struct {
	Message string  `json:"message"`
	Type    string  `json:"type"` // Examples: server_error, request
	Param   *string `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Error returns the APIError message.
func (e APIError) Error() string {
	if e.Type != "" {
		return e.Type + ": " + e.Message
	}
	return e.Message
}

// RequestError provides information about generic HTTP Request errors.
type RequestError struct {
	Code int
	Err  error
}

// Error returns the RequestError message.
func (e RequestError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("status code %d", e.Code)
	}
	var a APIError
	if ok := errors.As(e.Err, &a); ok {
		return fmt.Sprintf("%s %d %s", a.Type, e.Code, a.Message)
	}
	return fmt.Sprintf("status code %d: %s", e.Code, e.Err.Error())
}

// Unwrap returns the RequestError's underlying error.
func (e RequestError) Unwrap() error {
	return e.Err
}
