package openai

// FineTuneRecord provides a list of messages for fine-tuning a model.
// Each role (system, user, assistant) should be represented in the message list.
// Each record should be a persisted as a JSON object on a single line in a JSONL file.
// Training and validation files should be in JSONL format.
type FineTuneRecord struct {
	Messages []Message `json:"messages"`
}

// FineTuneRequest is a request to fine-tune a model.
type FineTuneRequest struct {
	// TrainingFileID is the ID of an uploaded file containing the training data.
	TrainingFileID string `json:"training_file"`

	// ValidationFileID is the ID of an uploaded file containing the validation data.
	ValidationFileID string `json:"validation_file,omitempty"`

	// Model is the base model ID to fine-tune. Example: "gpt-3.5-turbo".
	Model string `json:"model,omitempty"`

	// Suffix is a string of up to 40 characters that will be added to your fine-tuned model name.
	// This can be useful for distinguishing between different fine-tuned models.
	Suffix string `json:"suffix,omitempty"`

	// HyperParameters provides optional hyperparameters for fine-tuning.
	HyperParameters HyperParameters `json:"hyperparameters,omitempty"`
}

// FineTuneJob provides information about an OpenAPI fine-tuning job.
type FineTuneJob struct {
	// ID is the fine-tuning job ID, e.g. "ft-AF1WoRqd3aJAHsqc9NY7iL8F".
	ID string `json:"id"`

	// Object is the object type, e.g. "fine_tuning.job".
	Object string `json:"object"`

	// Model is the base model ID, e.g. "gpt-3.5-turbo".
	Model string `json:"model"`

	// Suffix is a string of up to 40 characters that will be added to your fine-tuned model name.
	// This can be useful for distinguishing between different fine-tuned models.
	Suffix string `json:"user_provided_suffix,omitempty"`

	// FineTunedModel is the ID of the fine-tuned model.
	FineTunedModel string `json:"fine_tuned_model,omitempty"`

	// TrainingFile is the ID of the file containing the training data.
	TrainingFile string `json:"training_file"`

	// ValidationFile  is the ID of the file containing the validation data.
	ValidationFile string `json:"validation_file,omitempty"`

	// ResultFiles is a list of file IDs containing the fine-tuning job results.
	ResultFiles []string `json:"result_files,omitempty"`

	// HyperParameters provides hyperparameters used for fine-tuning.
	HyperParameters HyperParameters `json:"hyperparameters"`

	// OrganizationID is the ID of the organization that owns the fine-tune.
	OrganizationID string `json:"organization_id,omitempty"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at"`

	// FinishedAt is an update timestamp in epoch seconds, e.g. 1669599635.
	FinishedAt int64 `json:"finished_at,omitempty"`

	// TrainedTokens is the total number of billable tokens process by this job.
	TrainedTokens int `json:"trained_tokens"`

	// Status is the current status of the fine-tuning job.
	// Examples: validating_files, queued, running, succeeded, failed, cancelled
	Status string `json:"status"`

	// Error provides information about an error that occurred during fine-tuning.
	Error FineTuneError `json:"error,omitempty"`
}

// Name returns the fine-tune model name, or model ID if the name is not set.
func (f FineTuneJob) Name() string {
	if f.FineTunedModel != "" {
		return f.FineTunedModel
	}
	return f.ID
}

// FineTuneJobs provides a list of fine-tuning jobs.
type FineTuneJobs struct {
	Object  string        `json:"object"`   // "list" is expected
	Data    []FineTuneJob `json:"data"`     // list of fine-tuning jobs
	HasMore bool          `json:"has_more"` // true if there are more jobs
}

// FineTuneError provides information about an error that occurred during fine-tuning.
type FineTuneError struct {
	// Code is a machine-readable error code.
	Code string `json:"code,omitempty"`

	// Param identifies an invalid parameter (e.g. training_file or validation_file).
	Param string `json:"param,omitempty"`

	// Message is a human-readable error message.
	Message string `json:"message,omitempty"`
}

// FineTuneMetric provides progress/performance metrics for a fine-tuning job.
type FineTuneMetric struct {
	Step               int     `json:"step,omitempty"`
	TrainingLoss       float64 `json:"train_loss,omitempty"`
	ValidationLoss     float64 `json:"valid_loss,omitempty"`
	TrainingAccuracy   float64 `json:"train_mean_token_accuracy,omitempty"`
	ValidationAccuracy float64 `json:"valid_mean_token_accuracy,omitempty"`
}

// FineTuneEvent provides information about an OpenAPI fine-tuning event.
type FineTuneEvent struct {
	// ID is the event ID, e.g. "ftevent-AF1WoRqd3aJAHsqc9NY7iL8F".
	ID string `json:"id"`

	// Object is the object type, e.g. "fine_tuning.job.event".
	Object string `json:"object"`

	// CreatedAt is a creation timestamp in epoch seconds, e.g. 1669599635.
	CreatedAt int64 `json:"created_at"`

	// Level is the event level, e.g. "info".
	Level string `json:"level"`

	// Message is the event message, e.g. "Job succeeded.".
	Message string `json:"message"`

	// Metrics provides performance metrics for the fine-tuning job.
	Metrics FineTuneMetric `json:"data,omitempty"`

	// EventType is the event type, e.g. "message" or "metrics".
	EventType string `json:"type,omitempty"`
}

// FineTuneEvents provides a list of fine-tuning events.
type FineTuneEvents struct {
	Object  string          `json:"object"`   // "list" is expected
	Data    []FineTuneEvent `json:"data"`     // list of events
	HasMore bool            `json:"has_more"` // true if there are more events
}

// HyperParameters provides hyperparameters for fine-tuning.
type HyperParameters struct {
	// EpochCount is the number of epochs to train for. The default is 4.
	// An epoch refers to one full cycle through the training dataset.
	EpochCount int `json:"n_epochs,omitempty"`

	// BatchSize is the number of training examples to process in parallel.
	// By default, the batch size will be dynamically configured to be ~0.2%
	// of the number of examples in the training set, capped at 256.
	BatchSize int `json:"batch_size,omitempty"`

	// LearningRate is the learning rate multiplier for the fine-tuning.
	// A smaller learning rate may be useful to avoid overfitting.
	LearningRate float64 `json:"learning_rate_multiplier,omitempty"`
}
