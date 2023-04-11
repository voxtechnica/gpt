package psy

import (
	"context"
	"fmt"
	"gpt/openai"
	"strconv"
	"sync"
	"time"

	"github.com/voxtechnica/tuid-go"
)

// Chat represents a complete request/response chat exchange.
type Chat struct {
	ID       string              `json:"id,omitempty"` // batch-unique ID
	Request  openai.ChatRequest  `json:"request,omitempty"`
	Response openai.ChatResponse `json:"response,omitempty"`
	Scores   []float32           `json:"scores,omitempty"`
	ErrMsg   string              `json:"error,omitempty"`
	Millis   int64               `json:"millis,omitempty"`
}

// String produces a simple text display of the Chat intended for console output.
func (c *Chat) String() string {
	s := "Chat"
	if len(c.ID) > 0 {
		s += " " + c.ID
	}
	if c.Millis > 0 {
		s += " (" + strconv.FormatInt(c.Millis, 10) + "ms)"
	}
	if len(c.Scores) > 0 {
		s += " scores: " + fmt.Sprint(c.Scores)
	}
	if len(c.ErrMsg) > 0 {
		s += " error: " + c.ErrMsg
	}
	s += "\n" + c.Request.String()
	s += c.Response.String()
	return s
}

// NewChat creates a new Chat object with a ChatRequest.
func NewChat(id, system, prompt, model string, temperature float32, maxTokens int) Chat {
	var messages []openai.Message
	if len(system) > 0 {
		messages = append(messages, openai.Message{
			Role:    openai.SYSTEM,
			Content: system,
		})
	}
	messages = append(messages, openai.Message{
		Role:    openai.USER,
		Content: prompt,
	})
	return Chat{
		ID: id,
		Request: openai.ChatRequest{
			Model:       model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			User:        id,
		},
	}
}

// CompleteChat generates a new chat completion.
func CompleteChat(ctx context.Context, client *openai.Client,
	id, system, prompt, model string, temperature float32, maxTokens int) (Chat, error) {
	startTime := time.Now()
	var chat Chat
	var err error
	// A prompt is required:
	if prompt == "" {
		return chat, fmt.Errorf("complete chat: prompt is required")
	}
	// A unique ID is required for each chat completion:
	if id == "" {
		id = tuid.NewID().String()
	}
	// Validate the model ID:
	if model == "" {
		model = "gpt-4"
	}
	if !client.ValidModel(ctx, model) {
		return chat, fmt.Errorf("complete chat: unrecognized model ID %s", model)
	}
	// Generate the chat request:
	chat = NewChat(id, system, prompt, model, temperature, maxTokens)
	// Generate the chat completion:
	chat.Response, err = client.CompleteChat(ctx, chat.Request)
	if err != nil {
		chat.ErrMsg = err.Error()
		chat.Millis = time.Since(startTime).Milliseconds()
		return chat, err
	}
	// Extract the scores:
	text, err := chat.Response.FirstMessageContent()
	if err == nil {
		chat.Scores = SelectScores(text, All)
	}
	// Calculate the time to complete:
	chat.Millis = time.Since(startTime).Milliseconds()
	return chat, nil
}

// CompleteChatBatch concurrently processes a single batch of chat completions.
func CompleteChatBatch(ctx context.Context, client *openai.Client, chats []Chat, sel Selection) map[string]Chat {
	results := make(chan Chat, len(chats))
	var wg sync.WaitGroup
	wg.Add(len(chats))
	for _, chat := range chats {
		go func(chat Chat) {
			startTime := time.Now()
			defer wg.Done()
			var err error
			chat.Response, err = client.CompleteChat(ctx, chat.Request)
			if err != nil {
				chat.ErrMsg = err.Error()
			} else {
				chat.ErrMsg = ""
				text, err := chat.Response.FirstMessageContent()
				if err == nil {
					chat.Scores = SelectScores(text, sel)
				}
			}
			chat.Millis = time.Since(startTime).Milliseconds()
			results <- chat
		}(chat)
	}
	wg.Wait()
	close(results)
	batch := make(map[string]Chat, len(chats))
	for chat := range results {
		batch[chat.ID] = chat
	}
	return batch
}

// Batch divides the provided slice of things into batches of the specified maximum size.
func Batch[T any](items []T, batchSize int) [][]T {
	batches := make([][]T, 0)
	batch := make([]T, 0)
	for _, item := range items {
		batch = append(batch, item)
		if len(batch) == batchSize {
			batches = append(batches, batch)
			batch = make([]T, 0)
		}
	}
	if len(batch) > 0 {
		batches = append(batches, batch)
	}
	return batches
}
