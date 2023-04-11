package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"

	"github.com/spf13/cobra"
)

// ModelCommand is the command for managing models.
type ModelCommand struct {
	apiClient *openai.Client
	rootCmd   *cobra.Command
	baseCmd   *cobra.Command
	listCmd   *cobra.Command
	readCmd   *cobra.Command
	raw       bool
}

// NewModelCommand creates and initializes the model commands.
func NewModelCommand(apiClient *openai.Client, root *cobra.Command) *ModelCommand {
	// Base Command
	c := &ModelCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "model",
		Short: "Manage models",
		Long:  "Manage models",
	}
	c.baseCmd.PersistentFlags().BoolVarP(&c.raw, "raw", "r", false, "Raw OpenAI Response?")
	c.rootCmd.AddCommand(c.baseCmd)

	// List Command
	c.listCmd = &cobra.Command{
		Use:   "list",
		Short: "List models",
		Long:  "List available models",
		RunE:  c.list,
	}
	c.listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	c.baseCmd.AddCommand(c.listCmd)

	// Read Command
	c.readCmd = &cobra.Command{
		Use:   "read <modelID> [modelID]...",
		Short: "Read specified model(s)",
		Long:  "Read the details about one or more models, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.read,
	}
	c.baseCmd.AddCommand(c.readCmd)

	return c
}

// list the available models.
func (c *ModelCommand) list(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	if c.raw {
		body, err := c.apiClient.ListModelsRaw(ctx)
		if body != nil {
			fmt.Print(string(body))
		}
		if err != nil {
			return err
		}
		return nil
	}

	// Retrieve the models:
	models, err := c.apiClient.ListModels(ctx)
	if err != nil {
		return err
	}

	// Display either full JSON or just the IDs:
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		j, err := json.MarshalIndent(models, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON models: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, model := range models {
			fmt.Println(model.ID)
		}
	}
	return nil
}

// read the metadata for the specified model(s).
func (c *ModelCommand) read(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	if c.raw {
		for _, modelID := range args {
			body, err := c.apiClient.ReadModelRaw(ctx, modelID)
			if body != nil {
				fmt.Print(string(body))
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Retrieve the model(s):
	for _, modelID := range args {
		model, err := c.apiClient.ReadModel(ctx, modelID)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(model, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON model: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}
