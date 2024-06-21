package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"time"

	"github.com/spf13/cobra"
)

// TuneCommand is the command for managing fine-tuning jobs.
type TuneCommand struct {
	apiClient *openai.Client
	rootCmd   *cobra.Command
	baseCmd   *cobra.Command
	listCmd   *cobra.Command
	readCmd   *cobra.Command
	eventsCmd *cobra.Command
	createCmd *cobra.Command
	cancelCmd *cobra.Command
	raw       bool
}

// NewTuneCommand creates and initializes the tune commands.
func NewTuneCommand(apiClient *openai.Client, root *cobra.Command) *TuneCommand {
	// Base Command
	c := &TuneCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "tune",
		Short: "Manage fine-tuning jobs",
		Long:  "Manage fine-tuning jobs.",
	}
	c.baseCmd.PersistentFlags().BoolVarP(&c.raw, "raw", "r", false, "Raw OpenAI Response?")
	c.rootCmd.AddCommand(c.baseCmd)

	// List Command
	c.listCmd = &cobra.Command{
		Use:   "list",
		Short: "List fine-tuning jobs",
		Long:  "List metadata of available fine-tuning jobs.",
		RunE:  c.list,
	}
	c.listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	c.baseCmd.AddCommand(c.listCmd)

	// Read Command
	c.readCmd = &cobra.Command{
		Use:   "read <jobID> [jobID]...",
		Short: "Read specified fine-tuning job(s)",
		Long:  "Read the metadata about one or more fine-tuning jobs, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.read,
	}
	c.baseCmd.AddCommand(c.readCmd)

	// Events Command
	c.eventsCmd = &cobra.Command{
		Use:   "events <jobID>",
		Short: "List events for a fine-tuning job",
		Long:  "List the event history for a specified fine-tuning job.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.events,
	}
	c.eventsCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	c.baseCmd.AddCommand(c.eventsCmd)

	// Create Command
	c.createCmd = &cobra.Command{
		Use:   "create <trainingFileID> [validationFileID]",
		Short: "Create a fine-tuning job",
		Long:  "Create a fine-tuning job using the provided file ID(s).",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.create,
	}
	c.createCmd.Flags().StringP("base", "b", "gpt-3.5-turbo", "Base model (default: gpt-3.5-turbo)")
	c.createCmd.Flags().StringP("suffix", "s", "", "Name suffix for the fine-tuned model")
	c.baseCmd.AddCommand(c.createCmd)

	// Cancel Command
	c.cancelCmd = &cobra.Command{
		Use:   "cancel <jobID> [jobID]...",
		Short: "Cancel specified fine-tuning job(s)",
		Long:  "Cancel one or more fine-tuning jobs, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.cancel,
	}
	c.baseCmd.AddCommand(c.cancelCmd)

	return c
}

// list the fine-tuned models.
func (c *TuneCommand) list(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	if c.raw {
		body, e := c.apiClient.ListFineTunesRaw(ctx)
		if body != nil {
			fmt.Print(string(body))
		}
		if e != nil {
			return e
		}
		return nil
	}

	// Retrieve the fine-tuned models.
	tunes, err := c.apiClient.ListFineTunes(ctx)
	if err != nil {
		return err
	}

	// Print the fine-tuned models.
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	if verbose {
		j, err := json.MarshalIndent(tunes, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling FineTune JSON: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, tune := range tunes {
			fmt.Println(tune.ID, tune.Status, tune.FineTunedModel)
		}
	}
	return nil
}

// read the metadata for the specified fine-tuned model(s).
func (c *TuneCommand) read(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	if c.raw {
		for _, id := range args {
			body, err := c.apiClient.ReadFineTuneRaw(ctx, id)
			if body != nil {
				fmt.Print(string(body))
			}
			if err != nil {
				return err
			}
		}
	} else {
		for _, id := range args {
			tune, err := c.apiClient.ReadFineTune(ctx, id)
			if err != nil {
				return err
			}
			j, err := json.MarshalIndent(tune, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling FineTune JSON: %w", err)
			}
			fmt.Println(string(j))
		}
	}
	return nil
}

// events lists the events for a specified fine-tuned model.
func (c *TuneCommand) events(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw OpenAI response?
	if c.raw {
		body, e := c.apiClient.ListFineTuneEventsRaw(ctx, args[0])
		if body != nil {
			fmt.Print(string(body))
		}
		if e != nil {
			return e
		}
		return nil
	}

	// Retrieve the events.
	events, err := c.apiClient.ListFineTuneEvents(ctx, args[0])
	if err != nil {
		return err
	}

	// Print the events.
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}
	if verbose {
		j, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling Events JSON: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, event := range events {
			t := time.Unix(event.CreatedAt, 0)
			fmt.Println(t, event.Level, event.Message)
		}
	}
	return nil
}

// create a fine-tuned model.
func (c *TuneCommand) create(cmd *cobra.Command, args []string) error {
	// Gather request parameters
	ctx := context.Background()
	base := cmd.Flag("base").Value.String()
	suffix := cmd.Flag("suffix").Value.String()
	trainingFileID := args[0]
	validationFileID := ""
	if len(args) > 1 {
		validationFileID = args[1]
	}

	// Validate the base model.
	if !c.apiClient.ValidModel(ctx, base) {
		return fmt.Errorf("invalid base model: %s", base)
	}

	// Validate the training file ID.
	_, err := c.apiClient.ReadFile(ctx, trainingFileID)
	if err != nil {
		return fmt.Errorf("invalid training file ID %s: %w", trainingFileID, err)
	}

	// Validate the validation file ID.
	if validationFileID != "" {
		_, err := c.apiClient.ReadFile(ctx, validationFileID)
		if err != nil {
			return fmt.Errorf("invalid validation file ID %s: %w", validationFileID, err)
		}
	}

	// Create the fine-tuned model, returning the raw response if requested.
	req := openai.FineTuneRequest{
		TrainingFileID:   trainingFileID,
		ValidationFileID: validationFileID,
		Model:            base,
		Suffix:           suffix,
	}
	if c.raw {
		body, err := c.apiClient.CreateFineTuneRaw(ctx, req)
		if body != nil {
			fmt.Print(string(body))
		}
		if err != nil {
			return err
		}
	} else {
		tune, err := c.apiClient.CreateFineTune(ctx, req)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(tune, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling FineTune JSON: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// cancel a fine-tuned model job in progress.
func (c *TuneCommand) cancel(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	for _, id := range args {
		if c.raw {
			body, err := c.apiClient.CancelFineTuneRaw(ctx, id)
			if body != nil {
				fmt.Print(string(body))
			}
			if err != nil {
				return err
			}
		} else {
			tune, err := c.apiClient.CancelFineTune(ctx, id)
			if err != nil {
				return err
			}
			j, err := json.MarshalIndent(tune, "", "  ")
			if err != nil {
				return fmt.Errorf("error marshalling FineTune JSON: %w", err)
			}
			fmt.Println(string(j))
		}
	}
	return nil
}
