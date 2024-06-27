package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"gpt/psy"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/voxtechnica/tuid-go"
)

// ChatCommand is the command for completing chat prompts.
type ChatCommand struct {
	apiClient     *openai.Client
	rootCmd       *cobra.Command
	baseCmd       *cobra.Command
	promptCmd     *cobra.Command
	randomCmd     *cobra.Command
	parallelCmd   *cobra.Command
	batchCmd      *cobra.Command
	resultsCmd    *cobra.Command
	raw           bool
	verbose       bool
	model         string
	temperature   float32
	maxTokens     int
	questionField string
	questionID    string
	answerField   string
	answerID      string
	scoreField    string
	scoreSelect   string
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
	c.baseCmd.PersistentFlags().StringVarP(&c.model, "model", "m", "gpt-4o", "Model ID")
	c.baseCmd.PersistentFlags().Float32VarP(&c.temperature, "temperature", "T", 0.5, "Temperature for sampling")
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
	c.promptCmd.Flags().BoolVarP(&c.raw, "raw", "r", false, "Raw OpenAI Response?")
	c.promptCmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "Verbose output?")
	c.promptCmd.Flags().StringVarP(&c.scoreSelect, "score-select", "S", "none", "Score selection: first | last | all | none")
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
	c.randomCmd.Flags().BoolVarP(&c.raw, "raw", "r", false, "Raw OpenAI Response?")
	c.randomCmd.Flags().BoolVarP(&c.verbose, "verbose", "v", false, "Verbose output?")
	c.randomCmd.Flags().StringVarP(&c.scoreSelect, "score-select", "S", "last", "Score selection: first | last | all | none")
	c.randomCmd.Flags().StringVarP(&c.questionID, "question-id", "Q", "", "Question ID (optional, name | name=value)")
	c.randomCmd.Flags().StringVarP(&c.questionField, "question-field", "q", "", "Question field name (optional)")
	c.randomCmd.Flags().StringVarP(&c.answerID, "answer-id", "A", "random", "Answer ID (optional, name=value)")
	c.randomCmd.Flags().StringVarP(&c.answerField, "answer-field", "a", "", "Answer field name (required)")
	c.randomCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.randomCmd)

	// Parallel Command
	// Example: gpt chat parallel examples/scores.csv examples/prompt.txt examples/system.txt examples/answers.csv examples/questions.csv -a answer -q question -Q qid=angry -m gpt-4 -T 0.2
	c.parallelCmd = &cobra.Command{
		Use:   "parallel <outputFile> <promptFile> <systemFile> <answerFile> [questionFile]",
		Short: "Chat complete answers in parallel",
		Long:  "Chat complete answers from a specified file with concurrent requests.",
		Args:  cobra.MinimumNArgs(4),
		RunE:  c.parallel,
	}
	c.parallelCmd.Flags().IntP("batch-size", "b", 20, "Concurrent request batch size")
	c.parallelCmd.Flags().StringVarP(&c.scoreField, "score-field", "s", "score", "Score field name")
	c.parallelCmd.Flags().StringVarP(&c.scoreSelect, "score-select", "S", "last", "Score selection: first | last | all | none")
	c.parallelCmd.Flags().StringVarP(&c.questionID, "question-id", "Q", "", "Question ID (optional, name | name=value)")
	c.parallelCmd.Flags().StringVarP(&c.questionField, "question-field", "q", "", "Question field name (optional)")
	c.parallelCmd.Flags().StringVarP(&c.answerField, "answer-field", "a", "", "Answer field name (required)")
	c.parallelCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.parallelCmd)

	// Batch Command
	// Example: gpt chat batch examples/scores.csv examples/prompt.txt examples/system.txt examples/answers.csv examples/questions.csv -a answer -q question -Q qid=angry -m gpt-4o -T 0.8
	c.batchCmd = &cobra.Command{
		Use:   "batch <outputFile> <promptFile> <systemFile> <answerFile> [questionFile]",
		Short: "Chat complete answers as an asynchronous batch",
		Long:  "Chat complete answers from a specified file as an asynchronous batch.",
		Args:  cobra.MinimumNArgs(4),
		RunE:  c.batchCreate,
	}
	c.batchCmd.Flags().IntP("wait", "w", 0, "Wait for results? Polling interval in seconds (recommend 10)")
	c.batchCmd.Flags().BoolP("input-only", "i", false, "Generate JSONL input file only?")
	c.batchCmd.Flags().StringVarP(&c.scoreField, "score-field", "s", "score", "Score field name")
	c.batchCmd.Flags().StringVarP(&c.scoreSelect, "score-select", "S", "last", "Score selection: first | last | all | none")
	c.batchCmd.Flags().StringVarP(&c.questionID, "question-id", "Q", "", "Question ID (optional, name | name=value)")
	c.batchCmd.Flags().StringVarP(&c.questionField, "question-field", "q", "", "Question field name (optional)")
	c.batchCmd.Flags().StringVarP(&c.answerField, "answer-field", "a", "", "Answer field name (required)")
	c.batchCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.batchCmd)

	// Results Command
	// Example: gpt chat results <batchID>
	c.resultsCmd = &cobra.Command{
		Use:   "results <batchID>",
		Short: "Process batch results",
		Long:  "Process the results of a completed batch operation.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.batchResults,
	}
	c.baseCmd.AddCommand(c.resultsCmd)

	return c
}

// prompt chat-completes a specified prompt.
func (c *ChatCommand) prompt(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	promptPath := args[0]
	systemPath := ""
	if len(args) > 1 {
		systemPath = args[1]
	}

	// Validate the score selection:
	sel := psy.Selection(strings.ToLower(c.scoreSelect))
	if !sel.IsValid() {
		return fmt.Errorf("invalid score selection (expect first, last, all, or none): %s", c.scoreSelect)
	}

	// Validate the model:
	if !c.apiClient.ValidModel(ctx, c.model) {
		return fmt.Errorf("model %s is not a recognized model ID", c.model)
	}

	// Read the system and prompt files:
	system, err := psy.ReadTextFile(systemPath)
	if err != nil {
		return fmt.Errorf("system file: %w", err)
	}
	prompt, err := psy.ReadTextFile(promptPath)
	if err != nil {
		return fmt.Errorf("prompt file: %w", err)
	}

	// Generate and output a chat response:
	chatID := tuid.NewID().String()
	chat := psy.NewChat(chatID, prompt, system, c.model, c.temperature, c.maxTokens)
	return c.generateChatResponse(ctx, chat, sel)
}

// random chat-completes a random prompt from the specified answer file.
func (c *ChatCommand) random(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	promptPath := args[0]
	systemPath := args[1]
	answerPath := args[2]
	questionPath := ""
	if len(args) > 3 {
		questionPath = args[3]
	}

	// Identify Chat Parameters:
	p := psy.ChatParameters{
		InputFile:     "",
		OutputFile:    "",
		SystemFile:    systemPath,
		PromptFile:    promptPath,
		QuestionFile:  questionPath,
		QuestionField: c.questionField,
		QuestionID:    c.questionID,
		AnswerFile:    answerPath,
		AnswerField:   c.answerField,
		AnswerID:      c.answerID,
		ScoreField:    c.scoreField,
		ScoreSelect:   psy.Selection(strings.ToLower(c.scoreSelect)),
		Model:         c.model,
		Temperature:   c.temperature,
		MaxTokens:     c.maxTokens,
	}

	// Generate the chat request:
	chats, _, err := c.generateChatRequests(p)
	if err != nil {
		return fmt.Errorf("generate chat request: %w", err)
	}
	if len(chats) == 0 {
		return fmt.Errorf("no chat request generated")
	}
	chat := chats[0]

	// Generate and output a chat response:
	return c.generateChatResponse(ctx, chat, p.ScoreSelect)
}

// parallel processes chat-completions for all answers in the specified file.
// Chat completions are processed concurrently in batches of the specified size.
// The results (answers plus scores) are written to the specified CSV file.
// If the question-id is just 'name' instead of 'name=value', then that field
// name is used in both the question file and the answer file to look up the
// question for each answer.
func (c *ChatCommand) parallel(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := context.Background()
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	outputPath := args[0]
	promptPath := args[1]
	systemPath := args[2]
	answerPath := args[3]
	questionPath := ""
	if len(args) > 4 {
		questionPath = args[4]
	}

	// Identify Chat Parameters:
	p := psy.ChatParameters{
		InputFile:     "",
		OutputFile:    outputPath,
		SystemFile:    systemPath,
		PromptFile:    promptPath,
		QuestionFile:  questionPath,
		QuestionField: c.questionField,
		QuestionID:    c.questionID,
		AnswerFile:    answerPath,
		AnswerField:   c.answerField,
		AnswerID:      "",
		ScoreField:    c.scoreField,
		ScoreSelect:   psy.Selection(strings.ToLower(c.scoreSelect)),
		Model:         c.model,
		Temperature:   c.temperature,
		MaxTokens:     c.maxTokens,
	}

	// Generate the chat requests:
	chats, answers, err := c.generateChatRequests(p)
	if err != nil {
		return fmt.Errorf("generate chat requests: %w", err)
	}

	// Process the chat completions concurrently, in batches:
	var count int
	results := make(map[string]psy.Chat, len(chats))
	retries := make([]psy.Chat, 0)
	batches := psy.Batch(chats, batchSize)
	fmt.Printf("Processing %d chats in %d batches of %d each...\n", len(chats), len(batches), batchSize)
	for i, batch := range batches {
		// Process the batch:
		batchStart := time.Now()
		r := psy.CompleteChatBatch(ctx, c.apiClient, batch, p.ScoreSelect)

		// Gather the results:
		for _, chat := range r {
			count++
			if chat.ErrMsg != "" {
				retries = append(retries, chat)
				fmt.Printf("%d: %s %dms %s\n", count, chat.ID, chat.Millis, chat.ErrMsg)
			} else {
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
		fmt.Printf("batch %d of %d: %d chats in %s, %s avg, %.2f%% complete, %s remaining\n", i+1,
			len(batches), len(batch), batchDuration, averageDuration, percentComplete, timeRemaining)
	}

	// Retry any failed requests:
	if len(retries) > 0 {
		fmt.Printf("retrying %d failed requests\n", len(retries))
		var retryCount int
		r := psy.CompleteChatBatch(ctx, c.apiClient, retries, p.ScoreSelect)
		for _, chat := range r {
			retryCount++
			if chat.ErrMsg != "" {
				fmt.Printf("%d: %s %dms %s\n", count, chat.ID, chat.Millis, chat.ErrMsg)
			} else {
				fmt.Printf("%d: %s %dms\n", retryCount, chat.ID, chat.Millis)
			}
			results[chat.ID] = chat
		}
	}

	// Add the completions and scores to the answers table:
	var maxScoreCount int
	var errorCount int
	for _, a := range answers.Records {
		chatID := a["chatID"]
		if chatID == "" {
			continue
		}
		chat, ok := results[chatID]
		if !ok {
			continue
		}
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
		if len(chat.Scores) > maxScoreCount {
			maxScoreCount = len(chat.Scores)
		}
		for i, score := range chat.Scores {
			field := p.ScoreField
			if maxScoreCount > 1 {
				field = fmt.Sprintf("%s%d", p.ScoreField, i+1)
			}
			a[field] = fmt.Sprintf("%f", score)
		}
	}

	// Add field names to the results table:
	answers.AddField("completion")
	if maxScoreCount == 1 {
		answers.AddField(p.ScoreField)
	} else if maxScoreCount > 1 {
		for i := 1; i <= maxScoreCount; i++ {
			field := fmt.Sprintf("%s%d", p.ScoreField, i)
			answers.AddField(field)
		}
	}

	// Write the results to the specified CSV file:
	err = answers.WriteCSV(outputPath)

	// Report the total time taken:
	fmt.Printf("completed %d chat completions (%d errors) in %s\n", len(chats), errorCount, time.Since(startTime))
	return err
}

// batchCreate processes chat-completions in an asynchronous batch.
// It's similar to the 'parallel' command, but the concurrent batch
// processing is managed by OpenAI. It creates a JSONL batch file of
// chat requests, uploads the file, and creates an OpenAI batch, which
// should be completed within 24 hours. If instructed to wait, it polls
// for progress and results. Once the batch is complete, it downloads
// the results file, processes it, and writes the results to the specified
// CSV output file.
func (c *ChatCommand) batchCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	wait, _ := cmd.Flags().GetInt("wait")
	inputOnly, _ := cmd.Flags().GetBool("input-only")
	outputPath := args[0]
	promptPath := args[1]
	systemPath := args[2]
	answerPath := args[3]
	questionPath := ""
	if len(args) > 4 {
		questionPath = args[4]
	}
	inputPath := strings.TrimSuffix(answerPath, ".csv") + ".jsonl"

	// Identify Chat Parameters:
	p := psy.ChatParameters{
		InputFile:     inputPath,
		OutputFile:    outputPath,
		SystemFile:    systemPath,
		PromptFile:    promptPath,
		QuestionFile:  questionPath,
		QuestionField: c.questionField,
		QuestionID:    c.questionID,
		AnswerFile:    answerPath,
		AnswerField:   c.answerField,
		AnswerID:      "",
		ScoreField:    c.scoreField,
		ScoreSelect:   psy.Selection(strings.ToLower(c.scoreSelect)),
		Model:         c.model,
		Temperature:   c.temperature,
		MaxTokens:     c.maxTokens,
	}

	// Generate the chat requests:
	chats, answers, err := c.generateChatRequests(p)
	if err != nil {
		return fmt.Errorf("generate chat requests: %w", err)
	}

	// Generate and upload the batch input file:
	var inputData bytes.Buffer
	for _, chat := range chats {
		item := openai.BatchRequestItem{
			CustomID: chat.ID,
			Method:   "POST",
			URL:      "/v1/chat/completions",
			Body:     chat.Request,
		}
		b, e := json.Marshal(item)
		if e != nil {
			return fmt.Errorf("marshal chat batch request item: %w", e)
		}
		inputData.Write(b)
		inputData.WriteString("\n")
	}
	inputBytes := inputData.Bytes()
	if inputOnly {
		if e := os.WriteFile(inputPath, inputBytes, 0644); e != nil {
			return fmt.Errorf("save batch input file %s: %w", inputPath, e)
		}
		fmt.Printf("saved batch input file %s\n", inputPath)
		return nil
	}
	file, err := c.apiClient.UploadFile(ctx, inputPath, "batch", inputBytes)
	if err != nil {
		return fmt.Errorf("upload batch input file %s: %w", inputPath, err)
	}
	fmt.Printf("uploaded %s input file %s: %s\n", file.Purpose, file.ID, file.FileName)

	// Create the batch operation:
	batchRequest := openai.BatchRequest{
		InputFileID:      file.ID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
		Metadata:         p.Metadata(),
	}
	batch, err := c.apiClient.CreateBatch(ctx, batchRequest)
	if err != nil {
		return fmt.Errorf("create batch: %w", err)
	}
	fmt.Printf("created batch %s: %s\n", batch.ID, batch.Status)

	// Save the answers table as an incomplete results CSV file with Chat IDs:
	err = answers.WriteCSV(outputPath)
	if err != nil {
		return fmt.Errorf("save incomplete results file %s: %w", outputPath, err)
	}
	fmt.Printf("saved incomplete results file (with chat IDs): %s\n", outputPath)

	// Poll the batch operation for completion:
	if wait > 0 {
		fmt.Println("polling for batch completion... (Ctrl+C to cancel)")
		for {
			batch, err = c.apiClient.ReadBatch(ctx, batch.ID)
			if err != nil {
				return fmt.Errorf("read batch %s: %w", batch.ID, err)
			}
			fmt.Println(batch.Progress())
			if batch.IsDone() {
				break
			}
			time.Sleep(time.Duration(wait) * time.Second)
		}
		// Process the results:
		return c.processBatchResults(batch.ID)
	}

	// If not waiting, provide instructions for monitoring progress:
	fmt.Println("Use the following command to monitor progress:")
	fmt.Printf("gpt batch monitor %s\n", batch.ID)
	fmt.Println("Once the batch is done, use the following command to process the results:")
	fmt.Printf("gpt chat results %s\n", batch.ID)
	return nil
}

// batchResults processes the results of a completed batch operation.
// It uses the batch metadata to determine how to process the results.
// It downloads the results file, processes it, and writes the results
// to the specified CSV output file.
func (c *ChatCommand) batchResults(cmd *cobra.Command, args []string) error {
	return c.processBatchResults(args[0])
}

// generateChatRequests generates chat requests from the specified questions/answers.
func (c *ChatCommand) generateChatRequests(p psy.ChatParameters) ([]psy.Chat, *psy.Table, error) {
	var chats []psy.Chat

	// Validate the model:
	if !c.apiClient.ValidModel(context.Background(), p.Model) {
		return chats, nil, fmt.Errorf("model %s is not a recognized model ID", p.Model)
	}

	// Validate the score selection:
	if !p.ScoreSelect.IsValid() {
		return chats, nil, fmt.Errorf("invalid score selection (expect first, last, all, or none): %s", p.ScoreSelect)
	}

	// Fetch the system template (optional):
	var err error
	var system string
	if p.SystemFile != "" {
		system, err = psy.ReadTextFile(p.SystemFile)
		if err != nil {
			return chats, nil, fmt.Errorf("system file: %w", err)
		}
	}

	// Fetch the prompt template:
	template, err := psy.ReadTextFile(p.PromptFile)
	if err != nil {
		return chats, nil, fmt.Errorf("prompt file: %w", err)
	}

	// Read the (optional) question(s):
	var questions map[string]string
	var question string
	var lookupQuestion bool
	if p.QuestionFile != "" {
		if strings.Contains(p.QuestionID, "=") {
			// Lookup the question by ID:
			question, err = psy.ReadCSVField(p.QuestionFile, p.QuestionID, p.QuestionField)
			if err != nil {
				return chats, nil, fmt.Errorf("read question: %w", err)
			}
		} else {
			// Read all the questions:
			lookupQuestion = true
			questions, err = psy.ReadCSVFields(p.QuestionFile, p.QuestionID, p.QuestionField)
			if err != nil {
				return chats, nil, fmt.Errorf("read questions: %w", err)
			}
		}
	}

	// Fetch the table of answers:
	answers, err := psy.ReadCSVTable(p.AnswerFile)
	if err != nil {
		return chats, nil, fmt.Errorf("answer file: %w", err)
	}
	if !answers.HasField(p.AnswerField) {
		return chats, answers, fmt.Errorf("answer field %s not found in %s", p.AnswerField, p.AnswerFile)
	}
	if lookupQuestion {
		if !answers.HasField(p.QuestionID) {
			return chats, answers, fmt.Errorf("question ID field %s not found in %s", p.QuestionID, p.AnswerFile)
		}
		// Validate the question IDs in the answer file:
		unknownQuestions := make([]string, 0)
		for _, a := range answers.Records {
			qid := a[p.QuestionID]
			if _, ok := questions[qid]; !ok {
				unknownQuestions = append(unknownQuestions, qid)
			}
		}
		if len(unknownQuestions) > 0 {
			return chats, answers, fmt.Errorf("unknown question IDs in answer file %s: %s", p.AnswerFile, strings.Join(unknownQuestions, ", "))
		}
	}

	// Select one or all records, as specified:
	var records []psy.Record
	if p.AnswerID == "random" {
		// Select a random answer:
		record := answers.Random()
		if record == nil {
			return chats, answers, fmt.Errorf("no records found in file %s", p.AnswerFile)
		}
		records = []psy.Record{record}
	} else if p.AnswerID != "" {
		// Select the specified answer by ID:
		nameValue := strings.Split(c.answerID, "=")
		if len(nameValue) != 2 {
			return chats, answers, fmt.Errorf("answer ID %s is not a name=value pair", c.answerID)
		}
		if !answers.HasField(nameValue[0]) {
			return chats, answers, fmt.Errorf("answer ID field %s not found in file %s", nameValue[0], p.AnswerFile)
		}
		record := answers.Record(nameValue[0], nameValue[1])
		if record == nil {
			return chats, answers, fmt.Errorf("answer %s not found in file %s", c.answerID, p.AnswerFile)
		}
		records = []psy.Record{record}
	} else {
		// Select all records:
		records = answers.Records
	}

	// Generate a chat request for each answer, skipping blanks. Also, add a
	// new column to the answer table, indicating its unique chat ID. This is
	// used to reconcile the answers with the chat completions.
	chats = make([]psy.Chat, 0, len(records))
	answers.AddField("chatID")
	for _, a := range records {
		answer := psy.CleanText(a[p.AnswerField])
		// Skip blank answers:
		if answer == "" {
			a["chatID"] = ""
			continue
		}
		// Generate a unique chat ID:
		chatID := tuid.NewID().String()
		a["chatID"] = chatID
		// Prepare the prompt from the template:
		var q string
		if lookupQuestion {
			qid := a[p.QuestionID]
			q = questions[qid]
		} else {
			q = question
		}
		prompt := strings.ReplaceAll(template, "{{question}}", q)
		prompt = strings.ReplaceAll(prompt, "{{answer}}", answer)
		// Generate the chat request:
		chat := psy.NewChat(chatID, system, prompt, c.model, c.temperature, c.maxTokens)
		chats = append(chats, chat)
	}

	return chats, answers, nil
}

// generateChatResponse generates and outputs a chat response from the specified chat request.
func (c *ChatCommand) generateChatResponse(ctx context.Context, chat psy.Chat, sel psy.Selection) error {
	// Raw response?
	if c.raw {
		// Echo the Request
		j, err := json.MarshalIndent(chat.Request, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON request: %w", err)
		}
		fmt.Println(string(j))
		// Output the Response
		b, err := c.apiClient.CompleteChatRaw(ctx, chat.Request)
		if len(b) > 0 {
			fmt.Print(string(b))
		}
		return err
	}

	// Complete the chat:
	chat, err := psy.CompleteChat(ctx, c.apiClient, chat, sel)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}
	if c.verbose {
		b, _ := json.MarshalIndent(chat, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Print(chat.String())
	}
	return nil
}

// processBatchResults processes the results of a completed batch operation.
func (c *ChatCommand) processBatchResults(batchID string) error {
	// Read the batch and associated response file(s):
	b, responses, err := c.apiClient.ReadBatchResponses(context.Background(), batchID)
	if err != nil {
		return err
	}

	// Verify that the incomplete results file exists:
	outputPath := b.Metadata["output_file"]
	if outputPath == "" {
		return fmt.Errorf("output_file path not found in batch %s metadata", batchID)
	}
	results, err := psy.ReadCSVTable(outputPath)
	if err != nil {
		return fmt.Errorf("read incomplete results file %s: %w", outputPath, err)
	}

	// Identify the score field and selection method:
	scoreField := b.Metadata["score_field"]
	if scoreField == "" {
		scoreField = "score"
	}
	scoreSelect := b.Metadata["score_select"]
	if scoreSelect == "" {
		scoreSelect = "last"
	}

	// Add the completion and scores to the results table:
	var maxScoreCount int
	for _, record := range results.Records {
		chatID := record["chatID"]
		if chatID == "" {
			continue
		}
		response, ok := responses[chatID]
		if !ok {
			continue
		}
		var completion string
		var scores []float32
		if response.HasError() {
			completion = response.Error.Error()
		} else {
			completion = response.Completion()
			scores = psy.SelectScores(completion, psy.Selection(scoreSelect))
		}
		record["completion"] = completion
		if len(scores) > maxScoreCount {
			maxScoreCount = len(scores)
		}
		for i, score := range scores {
			field := scoreField
			if maxScoreCount > 1 {
				field = fmt.Sprintf("%s%d", scoreField, i+1)
			}
			record[field] = fmt.Sprintf("%f", score)
		}
	}

	// Add field names to the results table:
	results.AddField("completion")
	if maxScoreCount == 1 {
		results.AddField(scoreField)
	} else if maxScoreCount > 1 {
		for i := 1; i <= maxScoreCount; i++ {
			field := fmt.Sprintf("%s%d", scoreField, i)
			results.AddField(field)
		}
	}

	// Write the results to the specified output CSV file:
	err = results.WriteCSV(outputPath)
	fmt.Printf("completed %d chats (%d failed) in %s\n", b.RequestCounts.Total, b.RequestCounts.Failed, b.Duration())
	return err
}
