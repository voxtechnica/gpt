package openai

// CommonModels is a collection of commonly-used OpenAI models.
var CommonModels = map[string]bool{
	"gpt-3.5-turbo":          true,
	"gpt-3.5-turbo-16k":      true,
	"gpt-3.5-turbo-instruct": true,
	"gpt-4":                  true,
	"gpt-4-turbo":            true,
	"gpt-4-turbo-preview":    true,
}

// Model identifies an OpenAPI model.
type Model struct {
	// ID is the model ID, e.g. "gpt-4".
	ID string `json:"id"`

	// Object is the object type, e.g. "model".
	Object string `json:"object"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1687882411.
	CreatedAt int64 `json:"created"`

	// OwnedBy is the owner of the model, e.g. "openai".
	OwnedBy string `json:"owned_by,omitempty"`
}

type ModelList struct {
	Object string  `json:"object"` // "list" is expected
	Data   []Model `json:"data"`   // list of models
}
