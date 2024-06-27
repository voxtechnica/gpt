package psy

import (
	"context"
	"fmt"
	"gpt/openai"
	"strconv"
	"sync"
	"time"
)

// ChatParameters represents the parameters for chat prompts and completions.
type ChatParameters struct {
	InputFile     string    `json:"inputFile,omitempty"`     // input file name
	OutputFile    string    `json:"outputFile,omitempty"`    // output file name
	SystemFile    string    `json:"systemFile,omitempty"`    // system message file
	PromptFile    string    `json:"promptFile,omitempty"`    // prompt template file
	QuestionFile  string    `json:"questionFile,omitempty"`  // question template file
	QuestionField string    `json:"questionField,omitempty"` // question field name
	QuestionID    string    `json:"questionID,omitempty"`    // question ID
	AnswerFile    string    `json:"answerFile,omitempty"`    // answer template file
	AnswerField   string    `json:"answerField,omitempty"`   // answer field name
	AnswerID      string    `json:"answerID,omitempty"`      // answer ID
	ScoreField    string    `json:"scoreField,omitempty"`    // score field name
	ScoreSelect   Selection `json:"scoreSelect,omitempty"`   // score selection
	Model         string    `json:"model,omitempty"`         // model ID
	Temperature   float32   `json:"temperature,omitempty"`   // temperature
	MaxTokens     int       `json:"maxTokens,omitempty"`     // maximum tokens
}

// Metadata returns a map of key-value pairs for the ChatParameters.
func (p ChatParameters) Metadata() map[string]string {
	m := make(map[string]string)
	if len(p.InputFile) > 0 {
		m["input_file"] = p.InputFile
	}
	if len(p.OutputFile) > 0 {
		m["output_file"] = p.OutputFile
	}
	if len(p.SystemFile) > 0 {
		m["system_file"] = p.SystemFile
	}
	if len(p.PromptFile) > 0 {
		m["prompt_file"] = p.PromptFile
	}
	if len(p.QuestionFile) > 0 {
		m["question_file"] = p.QuestionFile
	}
	if len(p.QuestionField) > 0 {
		m["question_field"] = p.QuestionField
	}
	if len(p.QuestionID) > 0 {
		m["question_id"] = p.QuestionID
	}
	if len(p.AnswerFile) > 0 {
		m["answer_file"] = p.AnswerFile
	}
	if len(p.AnswerField) > 0 {
		m["answer_field"] = p.AnswerField
	}
	if len(p.AnswerID) > 0 {
		m["answer_id"] = p.AnswerID
	}
	if len(p.ScoreField) > 0 {
		m["score_field"] = p.ScoreField
	}
	if p.ScoreSelect.IsValid() {
		m["score_select"] = p.ScoreSelect.String()
	}
	if len(p.Model) > 0 {
		m["model"] = p.Model
	}
	if p.Temperature > 0 {
		m["temperature"] = fmt.Sprintf("%f", p.Temperature)
	}
	if p.MaxTokens > 0 {
		m["max_tokens"] = strconv.Itoa(p.MaxTokens)
	}
	return m
}

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
func CompleteChat(ctx context.Context, client *openai.Client, chat Chat, sel Selection) (Chat, error) {
	startTime := time.Now()
	var err error
	// Generate the chat completion:
	chat.Response, err = client.CompleteChat(ctx, chat.Request)
	if err != nil {
		chat.ErrMsg = err.Error()
		chat.Millis = time.Since(startTime).Milliseconds()
		return chat, err
	}
	// Extract the score(s):
	text, err := chat.Response.FirstMessageContent()
	if err == nil {
		chat.Scores = SelectScores(text, sel)
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
