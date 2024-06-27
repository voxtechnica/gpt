package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"time"

	"github.com/spf13/cobra"
)

// BatchCommand is the command for managing batch operations.
type BatchCommand struct {
	apiClient  *openai.Client
	rootCmd    *cobra.Command
	baseCmd    *cobra.Command
	createCmd  *cobra.Command
	readCmd    *cobra.Command
	monitorCmd *cobra.Command
	cancelCmd  *cobra.Command
	listCmd    *cobra.Command
	raw        bool
}

// NewBatchCommand creates and initializes the batch commands.
func NewBatchCommand(apiClient *openai.Client, root *cobra.Command) *BatchCommand {
	// Base Command
	c := &BatchCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "batch",
		Short: "Manage batch operations",
		Long:  "Manage batch operations.",
	}
	c.baseCmd.PersistentFlags().BoolVarP(&c.raw, "raw", "r", false, "Raw OpenAI Response?")
	c.rootCmd.AddCommand(c.baseCmd)

	// Create Command
	c.createCmd = &cobra.Command{
		Use:   "create <inputFileID>",
		Short: "Create a new batch operation",
		Long:  "Create a new Chat Completion batch using the specified input file ID.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}
	c.baseCmd.AddCommand(c.createCmd)

	// Read Command
	c.readCmd = &cobra.Command{
		Use:   "read <batchID> [batchID]...",
		Short: "Read specified batch operation(s)",
		Long:  "Read the metadata about one or more batch operations, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.read,
	}
	c.baseCmd.AddCommand(c.readCmd)

	// Monitor Command
	c.monitorCmd = &cobra.Command{
		Use:   "monitor <batchID>",
		Short: "Monitor specified batch operation",
		Long:  "Monitor the progress of a batch operation, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.monitor,
	}
	c.monitorCmd.Flags().IntP("wait", "w", 10, "Wait interval (seconds)")
	c.baseCmd.AddCommand(c.monitorCmd)

	// Cancel Command
	c.cancelCmd = &cobra.Command{
		Use:   "cancel <batchID> [batchID]...",
		Short: "Cancel specified batch operation(s)",
		Long:  "Cancel one or more batch operations, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.cancel,
	}
	c.baseCmd.AddCommand(c.cancelCmd)

	// List Command
	c.listCmd = &cobra.Command{
		Use:   "list",
		Short: "List batch operations",
		Long:  "List metadata of available batch operations.",
		RunE:  c.list,
	}
	c.listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	c.listCmd.Flags().IntP("limit", "l", 20, "Limit")
	c.listCmd.Flags().StringP("after", "a", "", "After (last ID received)")
	c.baseCmd.AddCommand(c.listCmd)

	return c
}

// create is the handler for the "batch create" command.
func (c *BatchCommand) create(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate the input file
	inputFileID := args[0]
	f, err := c.apiClient.ReadFile(ctx, inputFileID)
	if err != nil {
		return fmt.Errorf("invalid input file ID %s: %w", inputFileID, err)
	}
	if f.Purpose != "batch" {
		return fmt.Errorf("invalid input file ID %s purpose %s: not a batch file", inputFileID, f.Purpose)
	}

	// Create the batch operation
	request := openai.BatchRequest{
		InputFileID:      inputFileID,
		Endpoint:         "/v1/chat/completions",
		CompletionWindow: "24h",
		Metadata: map[string]string{
			"input_file": f.FileName,
		},
	}

	// Retrieve the raw OpenAI response?
	if c.raw {
		// Echo the Request
		j, e := json.MarshalIndent(request, "", "  ")
		if e != nil {
			return fmt.Errorf("error marshalling JSON request: %w", e)
		}
		fmt.Println(string(j))
		// Output the Response
		b, e := c.apiClient.CreateBatchRaw(ctx, request)
		if len(b) > 0 {
			fmt.Print(string(b))
		}
		return e
	}

	// Create the batch operation
	b, err := c.apiClient.CreateBatch(ctx, request)
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling Batch JSON: %w", err)
	}
	fmt.Println(string(j))
	return nil
}

// read is the handler for the "batch read" command.
func (c *BatchCommand) read(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	if c.raw {
		for _, batchID := range args {
			body, e := c.apiClient.ReadBatchRaw(ctx, batchID)
			if body != nil {
				fmt.Print(string(body))
			}
			if e != nil {
				return e
			}
		}
		return nil
	}

	// Retrieve the batch operations
	for _, batchID := range args {
		batch, e := c.apiClient.ReadBatch(ctx, batchID)
		if e != nil {
			return e
		}
		j, err := json.MarshalIndent(batch, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling Batch JSON: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// monitor is the handler for the "batch monitor" command.
func (c *BatchCommand) monitor(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	wait, _ := cmd.Flags().GetInt("wait")
	batchID := args[0]
	for {
		batch, err := c.apiClient.ReadBatch(ctx, batchID)
		if err != nil {
			return fmt.Errorf("read batch %s: %w", batchID, err)
		}
		fmt.Println(batch.Progress())
		if batch.IsDone() {
			break
		}
		time.Sleep(time.Duration(wait) * time.Second)
	}
	return nil
}

// cancel is the handler for the "batch cancel" command.
func (c *BatchCommand) cancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	for _, batchID := range args {
		if c.raw {
			body, e := c.apiClient.CancelBatchRaw(ctx, batchID)
			if body != nil {
				fmt.Print(string(body))
			}
			if e != nil {
				return e
			}
		} else {
			batch, e := c.apiClient.CancelBatch(ctx, batchID)
			if e != nil {
				return e
			}
			j, err := json.MarshalIndent(batch, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling Batch JSON: %w", err)
			}
			fmt.Println(string(j))
		}
	}
	return nil
}

// list is the handler for the "batch list" command.
func (c *BatchCommand) list(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	limit, _ := cmd.Flags().GetInt("limit")
	after, _ := cmd.Flags().GetString("after")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Retrieve the raw OpenAI response?
	if c.raw {
		body, e := c.apiClient.ListBatchesRaw(ctx, limit, after)
		if body != nil {
			fmt.Println(string(body))
		}
		if e != nil {
			return e
		}
		return nil
	}

	// Retrieve the batch operations
	batches, hasMore, lastID, e := c.apiClient.ListBatches(ctx, limit, after)
	if e != nil {
		return e
	}
	if len(batches) == 0 {
		fmt.Println("No batch operations found.")
		return nil
	}

	// Print the batch operations
	if verbose {
		j, err := json.MarshalIndent(batches, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling Batches JSON: %w", err)
		}
		fmt.Println(string(j))
	} else {
		fmt.Println("BatchID\tInputFileID\tCreatedAt\tStatus\tElapsed")
		for _, b := range batches {
			createdAt := time.Unix(b.CreatedAt, 0).Format(time.DateTime)
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", b.ID, b.InputFileID, createdAt, b.Status, b.Duration())
		}
	}
	if hasMore {
		fmt.Printf("More results available. Use --limit=%d --after=%s to retrieve.\n", limit, lastID)
	}
	return nil
}
