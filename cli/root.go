package cli

import (
	"fmt"
	"gpt/openai"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// RootCommand is the root command for the application.
type RootCommand struct {
	apiClient *openai.Client
	rootCmd   *cobra.Command
	aboutCmd  *cobra.Command
	docCmd    *cobra.Command
	batchCmd  *BatchCommand
	chatCmd   *ChatCommand
	fileCmd   *FileCommand
	modelCmd  *ModelCommand
	tuneCmd   *TuneCommand
}

// NewRootCommand creates and initializes the root command and all its subcommands.
func NewRootCommand(apiClient *openai.Client) *RootCommand {
	// Root Command (application name)
	c := &RootCommand{
		apiClient: apiClient,
	}
	c.rootCmd = &cobra.Command{
		Use:     "gpt",
		Short:   "gpt: OpenAI GPT Command Line Tool",
		Long:    "gpt is a command line tool for working with OpenAI GPT models",
		Version: "0.2.1",
	}

	// About Command
	c.aboutCmd = &cobra.Command{
		Use:   "about",
		Short: "Print application information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gpt is a command line tool for working with OpenAI GPT models")
			fmt.Println("Version:", cmd.Root().Version)
			fmt.Println("Organization ID:", c.apiClient.OrgID)
			fmt.Println("API Key:", c.apiClient.APIKey)
		},
	}
	c.rootCmd.AddCommand(c.aboutCmd)

	// Documentation Command:
	c.docCmd = &cobra.Command{
		Use:   "docs",
		Short: "Generate gpt markdown documentation",
		Long:  "Generate gpt markdown documentation in the ./docs directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If the docs directory doesn't exist, create it
			if _, err := os.Stat("./docs"); os.IsNotExist(err) {
				err = os.Mkdir("./docs", 0755)
				if err != nil {
					return err
				}
			}
			// Generate the documentation
			return doc.GenMarkdownTree(c.rootCmd, "./docs")
		},
	}
	c.rootCmd.AddCommand(c.docCmd)

	// Other Commands
	c.batchCmd = NewBatchCommand(apiClient, c.rootCmd)
	c.chatCmd = NewChatCommand(apiClient, c.rootCmd)
	c.fileCmd = NewFileCommand(apiClient, c.rootCmd)
	c.modelCmd = NewModelCommand(apiClient, c.rootCmd)
	c.tuneCmd = NewTuneCommand(apiClient, c.rootCmd)

	return c
}

// Execute executes the root command.
func (c *RootCommand) Execute() error {
	return c.rootCmd.Execute()
}
