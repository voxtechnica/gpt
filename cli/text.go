package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// TextCommand is the command for completing text prompts.
type TextCommand struct {
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

// NewTextCommand creates and initializes the text commands.
func NewTextCommand(apiClient *openai.Client, root *cobra.Command) *TextCommand {
	// Base Command
	c := &TextCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "text",
		Short: "Complete a text prompt",
		Long:  "Complete a text prompt.",
	}
	c.baseCmd.PersistentFlags().StringVarP(&c.model, "model", "m", "text-davinci-003", "Model ID")
	c.baseCmd.PersistentFlags().Float32VarP(&c.temperature, "temperature", "T", 0.2, "Temperature for sampling")
	c.baseCmd.PersistentFlags().IntVarP(&c.maxTokens, "max-tokens", "t", 256, "Maximum number of tokens to generate")
	c.rootCmd.AddCommand(c.baseCmd)

	// Prompt Command
	c.promptCmd = &cobra.Command{
		Use:   "prompt <promptFile>",
		Short: "Complete a text prompt",
		Long:  "Complete a text prompt from a specified file",
		Args:  cobra.ExactArgs(1),
		RunE:  c.prompt,
	}
	c.promptCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.promptCmd.Flags().BoolP("verbose", "v", false, "Verbose output?")
	c.baseCmd.AddCommand(c.promptCmd)

	// Random Command
	c.randomCmd = &cobra.Command{
		Use:   "random <promptFile> <answerFile> [questionFile]",
		Short: "Text complete a random text answer",
		Long:  "Text complete a random text answer from a specified file.",
		Args:  cobra.MinimumNArgs(2),
		RunE:  c.random,
	}
	c.randomCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.randomCmd.Flags().BoolP("verbose", "v", false, "Verbose output?")
	c.randomCmd.Flags().StringP("question-id", "Q", "", "Question ID (optional, name=value)")
	c.randomCmd.Flags().StringP("question-field", "q", "", "Question field name (optional)")
	c.randomCmd.Flags().StringP("answer-field", "a", "", "Answer field name (required)")
	c.randomCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.randomCmd)

	// Batch Command
	c.batchCmd = &cobra.Command{
		Use:   "batch <outputFile> <promptFile> <answerFile> [questionFile]",
		Short: "Text complete a batch of answers",
		Long:  "Text  complete a batch of answers from a specified file.",
		Args:  cobra.MinimumNArgs(3),
		RunE:  c.batch,
	}
	c.batchCmd.Flags().IntP("batch-size", "b", 15, "Batch size for concurrent requests")
	c.batchCmd.Flags().StringP("question-id", "Q", "", "Question ID (optional, name=value)")
	c.batchCmd.Flags().StringP("question-field", "q", "", "Question field name (optional)")
	c.batchCmd.Flags().StringP("answer-field", "a", "", "Answer field name (required)")
	c.batchCmd.MarkFlagRequired("answer-field")
	c.baseCmd.AddCommand(c.batchCmd)

	return c
}

// prompt completes a specified prompt.
func (c *TextCommand) prompt(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	verbose, _ := cmd.Flags().GetBool("verbose")
	maxTokens, _ := cmd.Flags().GetInt("max-tokens")
	temperature, _ := cmd.Flags().GetFloat32("temperature")
	model, _ := cmd.Flags().GetString("model")
	promptFile := args[0]

	// Read the prompt file:
	f, err := os.Open(promptFile)
	if err != nil {
		return fmt.Errorf("error opening prompt file %s: %w", promptFile, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading prompt file %s: %w", promptFile, err)
	}
	prompt := string(b)

	// Generate the completion request:
	request := openai.TextRequest{
		Model:       model,
		Prompt:      prompt,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	// Output the request:
	if raw || verbose {
		b, _ := json.MarshalIndent(request, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Print(request.String())
	}

	// Raw response?
	if raw {
		response, e := c.apiClient.CompleteTextRaw(ctx, request)
		if response != nil {
			fmt.Print(string(response))
		}
		if e != nil {
			return e
		}
		return nil
	}

	// Complete the prompt:
	response, err := c.apiClient.CompleteText(ctx, request)
	if err != nil {
		return err
	}

	// Output the response:
	if verbose {
		b, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Print(response.String())
	}
	return nil
}

// random completes a random prompt from the specified answer file.
func (c *TextCommand) random(cmd *cobra.Command, args []string) error {
	fmt.Println("complete random not implemented yet. Use 'chat' instead.")
	return nil
}

// batch processes completions for all answers in the specified file.
// The results are written to the specified CSV file.
func (c *TextCommand) batch(cmd *cobra.Command, args []string) error {
	fmt.Println("complete batch not implemented yet. Use 'chat' instead.")
	return nil
}
