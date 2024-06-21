package openai

import (
	"errors"
	"fmt"
	"strings"
)

// ChatRequest represents a request structure for the chat completion API.
// This implementation is focused on producing text completions for a conversation.
// Note that the API also supports function calling and JSON responses, which
// require additional fields, not provided here.
type ChatRequest struct {
	// Model ID to use for completion. Example: "gpt-3.5-turbo" (required field)
	Model string `json:"model"`

	// Messages is a list of messages in the conversation (required field)
	Messages []Message `json:"messages"`

	// Temperature is the sampling temperature. Higher values result in more
	// random completions. Values range between 0 and 2. Higher values like
	// 0.8 will make the output more random, while lower values like 0.2 will
	// make it more focused and deterministic. The default is 1.0.
	Temperature float32 `json:"temperature,omitempty"`

	// TopP is the top-p sampling parameter. If set to a value between 0 and 1,
	// the returned text will be sampled from the smallest possible set of
	// tokens whose cumulative probability exceeds the value of top_p. For
	// example, if top_p is set to 0.1, the API will only consider the top 10%
	// probability tokens each step. This can be used to ensure that the
	// returned text doesn't contain undesirable tokens. The default is 1.0.
	TopP float32 `json:"top_p,omitempty"`

	// N is the number of results to return. The default is 1.
	N int `json:"n,omitempty"`

	// MaxTokens is the maximum number of tokens to generate.
	// The default is "infinity" (limited only by the context window size).
	MaxTokens int `json:"max_tokens,omitempty"`

	// PresencePenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on whether they appear in the text so far.
	// The default is 0.0.
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty is a floating point value between -2.0 and 2.0 that
	// penalizes new tokens based on their existing frequency in the text so
	// far. The default is 0.0.
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// User is a unique identifier representing your end-user, which can help
	// OpenAI to monitor and detect abuse. The default is an empty string.
	User string `json:"user,omitempty"`
}

// String produces a simple text display of the ChatRequest intended for console output.
func (c *ChatRequest) String() string {
	s := "--------------------\n" + c.Model
	if c.Temperature > 0 {
		s += fmt.Sprintf(" temp=%.2f", c.Temperature)
	}
	if c.MaxTokens > 0 {
		s += fmt.Sprintf(" max=%d", c.MaxTokens)
	}
	if c.User != "" {
		s += fmt.Sprintf(" user=%s", c.User)
	}
	s += "\n"
	for _, m := range c.Messages {
		s += m.String()
	}
	return s
}

// ChatResponse provides a predicted text completion in response to a provided
// prompt and other parameters.
type ChatResponse struct {
	ID        string          `json:"id"`      // eg. "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve"
	Object    string          `json:"object"`  // eg. "chat.completion"
	CreatedAt int64           `json:"created"` // epoch seconds, eg. 1677966478
	Model     string          `json:"model"`   // eg. "gpt-3.5-turbo"
	Usage     Usage           `json:"usage"`
	Choices   []MessageChoice `json:"choices"`
}

// String provides a simple text display of the ChatResponse intended for console output.
func (c *ChatResponse) String() string {
	var s string
	for _, m := range c.Choices {
		s += m.Message.String()
	}
	var finish string
	if len(c.Choices) > 0 {
		finish = "finish=" + c.Choices[0].FinishReason
	}
	s += fmt.Sprintf("--------------------\n%s %s %s\n", c.Model, c.Usage, finish)
	return s
}

// FirstMessageContent returns the content of the first message in the response.
func (c *ChatResponse) FirstMessageContent() (string, error) {
	if len(c.Choices) == 0 {
		return "", errors.New("chat: no choices found")
	}
	if len(c.Choices[0].Message.Content) == 0 {
		return "", errors.New("chat: no content found")
	}
	return c.Choices[0].Message.Content, nil
}

// MessageChoice represents a choice in a chat completion.
type MessageChoice struct {
	Message      Message `json:"message"`
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"` // e.g. "stop"
}

// Message represents a message in a chat conversation.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// String provides a simple text display of the Message intended for console output.
func (m *Message) String() string {
	return fmt.Sprintf("--------------------\n%s:\n%s\n", m.Role, strings.TrimSpace(m.Content))
}

// Usage provides the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// String returns a string representation of Usage.
func (u Usage) String() string {
	return fmt.Sprintf("prompt=%d completion=%d total=%d", u.PromptTokens, u.CompletionTokens, u.TotalTokens)
}
