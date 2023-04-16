package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"gpt/psy"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/voxtechnica/tuid-go"
)

// ChatCommand is the command for completing chat prompts.
type ChatCommand struct {
	apiClient   *openai.Client
	rootCmd     *cobra.Command
	baseCmd     *cobra.Command
	promptCmd   *cobra.Command
	randomCmd   *cobra.Command
	batchCmd    *cobra.Command
	model       string
	temperature float32
	maxTokens   int
}

// NewChatCommand creates and initializes the chat commands.
func NewChatCommand(apiClient *openai.Client, root *cobra.Command) *ChatCommand {
	// Base Command
	c := &ChatCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "chat",
		Short: "Complete a chat prompt",
		Long:  "Complete a chat prompt.",
	}
	c.baseCmd.PersistentFlags().StringVarP(&c.model, "model", "m", "gpt-4", "Model ID")
	c.baseCmd.PersistentFlags().Float32VarP(&c.temperature, "temperature", "T", 0.2, "Temperature for sampling")
	c.baseCmd.PersistentFlags().IntVarP(&c.maxTokens, "max-tokens", "t", 0, "Maximum number of tokens to generate")
	c.rootCmd.AddCommand(c.baseCmd)

	// Prompt Command
	// Example: gpt chat prompt examples/limerick.txt examples/system.txt -m gpt-4 -T 0.4 -t 256
	c.promptCmd = &cobra.Command{
		Use:   "prompt <promptFile> [systemFile]",
		Short: "Chat complete a test prompt",
		Long:  "Chat complete a test prompt from a specified file.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.prompt,
	}
	c.promptCmd.Flags().BoolP("verbose", "v", false, "Verbose output?")
	c.baseCmd.AddCommand(c.promptCmd)

	// Random Command
	// Example: gpt chat random examples/prompt.txt examples/system.txt examples/answers.csv examples/questions.csv -a answer -q question -Q qid=angry -m gpt-4 -T 0.2
	c.randomCmd = &cobra.Command{
		Use:   "random <promptFile> <systemFile> <answerFile> [questionFile]",
		Short: "Chat complete a random answer",
		Long:  "Chat complete a random answer from a specified file.",
		Args:  cobra.MinimumNArgs(3),
		RunE:  c.random,
	}
	c.randomCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.randomCmd.Flags().BoolP("verbose", "v", false, "Verbose output?")
	c.randomCmd.Flags().StringP("score-select", "S", "last", "Score selection: first | last | all | none")
	c.randomCmd.Flags().StringP("question-id", "Q", "", "Question ID (optional, name | name=value)")
	c.randomCmd.Flags().StringP("question-field", "q", "", "Question field name (optional)")
	c.randomCmd.Flags().StringP("answer-id", "A", "", "Answer ID (optional, name=value)")
	c.randomCmd.Flags().StringP("answer-field", "a", "", "Answer field name (required)")
	c.randomCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.randomCmd)

	// Batch Command
	// Example: gpt chat batch examples/scores.csv examples/prompt.txt examples/system.txt examples/answers.csv examples/questions.csv -a answer -q question -Q qid=angry -m gpt-4 -T 0.2
	c.batchCmd = &cobra.Command{
		Use:   "batch <outputFile> <promptFile> <systemFile> <answerFile> [questionFile]",
		Short: "Chat complete a batch of answers",
		Long:  "Chat complete a batch of answers from a specified file.",
		Args:  cobra.MinimumNArgs(4),
		RunE:  c.batch,
	}
	c.batchCmd.Flags().IntP("batch-size", "b", 20, "Batch size for concurrent requests")
	c.batchCmd.Flags().StringP("score-field", "s", "score", "Score field name")
	c.batchCmd.Flags().StringP("score-select", "S", "last", "Score selection: first | last | all | none")
	c.batchCmd.Flags().StringP("question-id", "Q", "", "Question ID (optional, name | name=value)")
	c.batchCmd.Flags().StringP("question-field", "q", "", "Question field name (optional)")
	c.batchCmd.Flags().StringP("answer-field", "a", "", "Answer field name (required)")
	c.batchCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.batchCmd)

	return c
}

// prompt chat-completes a specified prompt.
func (c *ChatCommand) prompt(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	verbose, _ := cmd.Flags().GetBool("verbose")
	promptPath := args[0]
	systemPath := ""
	if len(args) > 1 {
		systemPath = args[1]
	}
	prompt, err := psy.ReadTextFile(promptPath)
	if err != nil {
		return fmt.Errorf("prompt file: %w", err)
	}
	system, err := psy.ReadTextFile(systemPath)
	if err != nil {
		return fmt.Errorf("system file: %w", err)
	}
	chat, err := psy.CompleteChat(ctx, c.apiClient, "", system, prompt, c.model, c.temperature, c.maxTokens)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}
	if verbose {
		b, _ := json.MarshalIndent(chat, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Print(chat.String())
	}
	return nil
}

// random chat-completes a random prompt from the specified answer file.
func (c *ChatCommand) random(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	verbose, _ := cmd.Flags().GetBool("verbose")
	scoreSelect, _ := cmd.Flags().GetString("score-select")
	answerID, _ := cmd.Flags().GetString("answer-id")
	answerField, _ := cmd.Flags().GetString("answer-field")
	questionField, _ := cmd.Flags().GetString("question-field")
	questionID, _ := cmd.Flags().GetString("question-id")
	questionPath := ""
	promptPath := args[0]
	systemPath := args[1]
	answerPath := args[2]
	if len(args) > 3 {
		questionPath = args[3]
	}

	// Validate the score selection:
	sel := psy.Selection(strings.ToLower(scoreSelect))
	if !sel.IsValid() {
		return fmt.Errorf("invalid score selection (expect first, last, all, or none): %s", scoreSelect)
	}

	// Read the answers:
	answers, err := psy.ReadCSVTable(answerPath)
	if err != nil {
		return fmt.Errorf("answer file: %w", err)
	}
	if len(answerField) == 0 {
		return fmt.Errorf("answer field is required")
	}
	if !answers.HasField(answerField) {
		return fmt.Errorf("answer field %s not found in file %s", answerField, answerPath)
	}

	// Select an answer:
	var record psy.Record
	if len(answerID) > 0 {
		// Select the specified answer by ID:
		nameValue := strings.Split(answerID, "=")
		if len(nameValue) != 2 {
			return fmt.Errorf("answer ID %s is not a name=value pair", answerID)
		}
		if !answers.HasField(nameValue[0]) {
			return fmt.Errorf("answer ID field %s not found in file %s", nameValue[0], answerPath)
		}
		// Read the specified record:
		record = answers.Record(nameValue[0], nameValue[1])
		if record == nil {
			return fmt.Errorf("answer %s not found in file %s", answerID, answerPath)
		}
	} else {
		// Select a random answer:
		record = answers.Random()
	}
	answer := psy.CleanText(record[answerField])
	if len(answer) == 0 {
		return fmt.Errorf("answer: selected field %s is empty in file %s", answerField, answerPath)
	}

	// Read the question(s):
	var question string
	if len(questionPath) > 0 {
		if strings.Contains(questionID, "=") {
			// Lookup the question by ID:
			question, err = psy.ReadCSVField(questionPath, questionID, questionField)
			if err != nil {
				return fmt.Errorf("read question: %w", err)
			}
		} else {
			// Lookup the question specified in the answer record:
			var questions map[string]string
			questions, err = psy.ReadCSVFields(questionPath, questionID, questionField)
			if err != nil {
				return fmt.Errorf("read questions: %w", err)
			}
			qid := record[questionID]
			q, ok := questions[qid]
			if !ok {
				return fmt.Errorf("question ID %s not found in file %s", qid, questionPath)
			}
			question = q
		}
	}

	// Prepare the prompt:
	system, err := psy.ReadTextFile(systemPath)
	if err != nil {
		return fmt.Errorf("system file: %w", err)
	}
	prompt, err := psy.ReadTextFile(promptPath)
	if err != nil {
		return fmt.Errorf("prompt file: %w", err)
	}
	prompt = strings.ReplaceAll(prompt, "{{question}}", question)
	prompt = strings.ReplaceAll(prompt, "{{answer}}", answer)

	// Validate the model:
	if !c.apiClient.ValidModel(ctx, c.model) {
		return fmt.Errorf("model %s is not a recognized model ID", c.model)
	}

	// Generate the chat request:
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
	userID := tuid.NewID().String()
	request := openai.ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
		User:        userID,
	}

	// Raw response?
	if raw {
		// Echo the Request
		jsonReq, e := json.MarshalIndent(request, "", "  ")
		if e != nil {
			return fmt.Errorf("error marshalling JSON request: %w", e)
		}
		fmt.Println(string(jsonReq))
		// Output the Response
		body, e := c.apiClient.CompleteChatRaw(ctx, request)
		if e != nil {
			return e
		}
		fmt.Print(string(body))
		return nil
	}

	// Chat complete the prompt:
	response, err := c.apiClient.CompleteChat(ctx, request)
	if err != nil {
		return err
	}

	// Extract the scores:
	var scores []float32
	text, err := response.FirstMessageContent()
	if err == nil {
		scores = psy.SelectScores(text, sel)
	}
	if err != nil {
		return err
	}
	chat := psy.Chat{
		ID:       userID,
		Request:  request,
		Response: response,
		Scores:   scores,
		Millis:   time.Since(startTime).Milliseconds(),
	}
	if verbose {
		j, err := json.MarshalIndent(chat, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON chat completion: %w", err)
		}
		fmt.Println(string(j))
	} else {
		fmt.Print(chat.String())
	}
	return nil
}

// batch processes chat-completions for all answers in the specified file.
// The results (answers plus scores) are written to the specified CSV file.
// If the question-id is just 'name' instead of 'name=value', then that field
// name is used in both the question file and the answer file to look up the
// question for each answer.
func (c *ChatCommand) batch(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	scoreField, _ := cmd.Flags().GetString("score-field")
	scoreSelect, _ := cmd.Flags().GetString("score-select")
	answerField, _ := cmd.Flags().GetString("answer-field")
	questionField, _ := cmd.Flags().GetString("question-field")
	questionID, _ := cmd.Flags().GetString("question-id")
	questionPath := ""
	outputPath := args[0]
	promptPath := args[1]
	systemPath := args[2]
	answerPath := args[3]
	if len(args) > 4 {
		questionPath = args[4]
	}

	// Validate the score selection:
	sel := psy.Selection(strings.ToLower(scoreSelect))
	if !sel.IsValid() {
		return fmt.Errorf("invalid score selection (expect first, last, all, or none): %s", scoreSelect)
	}

	// Fetch the prompt template:
	system, err := psy.ReadTextFile(systemPath)
	if err != nil {
		return fmt.Errorf("system file: %w", err)
	}
	template, err := psy.ReadTextFile(promptPath)
	if err != nil {
		return fmt.Errorf("prompt file: %w", err)
	}

	// Read the question(s):
	var questions map[string]string
	var question string
	var lookupQuestion bool
	if len(questionPath) > 0 {
		if strings.Contains(questionID, "=") {
			// Lookup the question by ID:
			question, err = psy.ReadCSVField(questionPath, questionID, questionField)
			if err != nil {
				return fmt.Errorf("read question: %w", err)
			}
		} else {
			// Read all the questions:
			lookupQuestion = true
			questions, err = psy.ReadCSVFields(questionPath, questionID, questionField)
			if err != nil {
				return fmt.Errorf("read questions: %w", err)
			}
		}
	}

	// Fetch the table of answers:
	answers, err := psy.ReadCSVTable(answerPath)
	if err != nil {
		return fmt.Errorf("answer file: %w", err)
	}
	if !answers.HasField(answerField) {
		return fmt.Errorf("answer field %s not found in %s", answerField, answerPath)
	}
	if lookupQuestion {
		if !answers.HasField(questionID) {
			return fmt.Errorf("question ID field %s not found in %s", questionID, answerPath)
		}
		// Validate the question IDs in the answer file:
		unknownQuestions := make([]string, 0)
		for _, a := range answers.Records {
			qid := a[questionID]
			if _, ok := questions[qid]; !ok {
				unknownQuestions = append(unknownQuestions, qid)
			}
		}
		if len(unknownQuestions) > 0 {
			return fmt.Errorf("unknown question IDs in answer file %s: %s", answerPath, strings.Join(unknownQuestions, ", "))
		}
	}
	answers.AddField("chatID")
	answers.AddField("completion")

	// Validate the model:
	if !c.apiClient.ValidModel(ctx, c.model) {
		return fmt.Errorf("model %s is not a recognized model ID", c.model)
	}

	// Generate a chat request for each answer, skipping blanks:
	chats := make([]psy.Chat, 0, len(answers.Records))
	for _, a := range answers.Records {
		// Skip blank answers:
		if a[answerField] == "" {
			a["chatID"] = ""
			continue
		}
		// Generate a unique chat ID:
		chatID := tuid.NewID().String()
		a["chatID"] = chatID
		// Prepare the prompt from the template:
		var q string
		if lookupQuestion {
			qid := a[questionID]
			q = questions[qid]
		} else {
			q = question
		}
		answer := psy.CleanText(a[answerField])
		prompt := strings.ReplaceAll(template, "{{question}}", q)
		prompt = strings.ReplaceAll(prompt, "{{answer}}", answer)
		// Generate the chat request:
		chat := psy.NewChat(chatID, system, prompt, c.model, c.temperature, c.maxTokens)
		chats = append(chats, chat)
	}

	// Process the chat completions in batches:
	var count int
	var maxScoreCount int
	results := make(map[string]psy.Chat, len(chats))
	retries := make([]psy.Chat, 0)
	batches := psy.Batch(chats, batchSize)
	fmt.Printf("Processing %d chats in %d batches of %d each...\n", len(chats), len(batches), batchSize)
	for i, batch := range batches {
		// Process the batch:
		batchStart := time.Now()
		r := psy.CompleteChatBatch(ctx, c.apiClient, batch, sel)

		// Gather the results:
		for _, chat := range r {
			count++
			if chat.ErrMsg != "" {
				retries = append(retries, chat)
				fmt.Printf("%d: %s %dms %s\n", count, chat.ID, chat.Millis, chat.ErrMsg)
			} else {
				if len(chat.Scores) > maxScoreCount {
					maxScoreCount = len(chat.Scores)
				}
				fmt.Printf("%d: %s %dms\n", count, chat.ID, chat.Millis)
			}
			results[chat.ID] = chat
		}

		// Report batch time taken, progress, and predicted time remaining:
		batchDuration := time.Since(batchStart)
		totalDuration := time.Since(startTime)
		averageDuration := totalDuration / time.Duration(count)
		timeRemaining := time.Duration(len(chats)-count) * averageDuration
		percentComplete := float32(count) / float32(len(chats)) * 100
		fmt.Printf("batch %d of %d: %d chats in %sms, %s avg, %.2f%% complete, %s remaining\n", i+1,
			len(batches), len(batch), batchDuration, averageDuration, percentComplete, timeRemaining)
	}

	// Retry any failed requests:
	if len(retries) > 0 {
		fmt.Printf("retrying %d failed requests\n", len(retries))
		var retryCount int
		r := psy.CompleteChatBatch(ctx, c.apiClient, retries, sel)
		for _, chat := range r {
			retryCount++
			if chat.ErrMsg != "" {
				fmt.Printf("%d: %s %dms %s\n", count, chat.ID, chat.Millis, chat.ErrMsg)
			} else {
				if len(chat.Scores) > maxScoreCount {
					maxScoreCount = len(chat.Scores)
				}
				fmt.Printf("%d: %s %dms\n", retryCount, chat.ID, chat.Millis)
			}
			results[chat.ID] = chat
		}
	}

	// Add the score field names to the answers table:
	if maxScoreCount == 1 {
		answers.AddField(scoreField)
	} else if maxScoreCount > 1 {
		for i := 1; i <= maxScoreCount; i++ {
			field := fmt.Sprintf("%s%d", scoreField, i)
			answers.AddField(field)
		}
	}

	// Add the scores to the answers table:
	var errorCount int
	for _, a := range answers.Records {
		chatID := a["chatID"]
		if chatID == "" {
			continue
		}
		chat := results[chatID]
		var completion string
		if chat.ErrMsg != "" {
			errorCount++
			completion = chat.ErrMsg
		} else {
			completion, err = chat.Response.FirstMessageContent()
			if err != nil {
				errorCount++
				completion = err.Error()
			}
		}
		a["completion"] = completion
		for i, score := range chat.Scores {
			field := scoreField
			if maxScoreCount > 1 {
				field = fmt.Sprintf("%s%d", scoreField, i+1)
			}
			a[field] = fmt.Sprintf("%f", score)
		}
	}

	// Write the results to the specified CSV file:
	err = answers.WriteCSV(outputPath)

	// Report the total time taken:
	fmt.Printf("completed %d chat completions (%d errors) in %s\n", len(chats), errorCount, time.Since(startTime))
	return err
}
