package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gpt/openai"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// FileCommand is the command for managing files.
type FileCommand struct {
	apiClient   *openai.Client
	rootCmd     *cobra.Command
	baseCmd     *cobra.Command
	listCmd     *cobra.Command
	readCmd     *cobra.Command
	uploadCmd   *cobra.Command
	downloadCmd *cobra.Command
	deleteCmd   *cobra.Command
}

// NewFileCommand creates and initializes the file commands.
func NewFileCommand(apiClient *openai.Client, root *cobra.Command) *FileCommand {
	// Base Command
	c := &FileCommand{
		apiClient: apiClient,
		rootCmd:   root,
	}
	c.baseCmd = &cobra.Command{
		Use:   "file",
		Short: "Manage files",
		Long:  "Manage files",
	}
	c.rootCmd.AddCommand(c.baseCmd)

	// List Command
	c.listCmd = &cobra.Command{
		Use:   "list",
		Short: "List files",
		Long:  "List metadata of available files",
		RunE:  c.list,
	}
	c.listCmd.Flags().StringP("purpose", "p", "", "File Purpose")
	c.listCmd.Flags().BoolP("verbose", "v", false, "Verbose? (full JSON)")
	c.listCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.baseCmd.AddCommand(c.listCmd)

	// Read Command
	c.readCmd = &cobra.Command{
		Use:   "read <fileID> [fileID]...",
		Short: "Read specified file(s)",
		Long:  "Read the metadata about one or more files, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.read,
	}
	c.readCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.baseCmd.AddCommand(c.readCmd)

	// Upload Command
	c.uploadCmd = &cobra.Command{
		Use:   "upload <jsonlFile>",
		Short: "Upload a JSONL file",
		Long:  "Upload a JSONL fine-tuning file",
		Args:  cobra.ExactArgs(1),
		RunE:  c.upload,
	}
	c.uploadCmd.Flags().StringP("purpose", "p", "fine-tune", "File Purpose")
	c.baseCmd.AddCommand(c.uploadCmd)

	// Download Command
	c.downloadCmd = &cobra.Command{
		Use:   "download <fileID>",
		Short: "Download a file",
		Long:  "Download a file to the specified path/name (default: OpenAI file name).",
		Args:  cobra.ExactArgs(1),
		RunE:  c.download,
	}
	c.downloadCmd.Flags().StringP("output", "o", "", "Output File Path")
	c.baseCmd.AddCommand(c.downloadCmd)

	// Delete Command
	c.deleteCmd = &cobra.Command{
		Use:   "delete <fileID> [fileID]...",
		Short: "Delete specified file(s)",
		Long:  "Delete one or more files, specified by ID.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.delete,
	}
	c.deleteCmd.Flags().BoolP("raw", "r", false, "Raw OpenAI Response?")
	c.baseCmd.AddCommand(c.deleteCmd)

	return c
}

// list the available files.
func (c *FileCommand) list(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	purpose := cmd.Flag("purpose").Value.String()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		body, err := c.apiClient.ListFilesRaw(ctx, purpose)
		if body != nil {
			fmt.Print(string(body))
		}
		if err != nil {
			return err
		}
		return nil
	}

	// Retrieve the files:
	files, err := c.apiClient.ListFiles(ctx, purpose)
	if err != nil {
		return err
	}

	// Display either full JSON or just the IDs:
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		j, err := json.MarshalIndent(files, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON files: %w", err)
		}
		fmt.Println(string(j))
	} else {
		for _, file := range files {
			fmt.Println(file.ID, file.Purpose, file.FileName)
		}
	}
	return nil
}

// read the metadata for the specified file(s).
func (c *FileCommand) read(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Retrieve the raw JSON response:
	raw, _ := cmd.Flags().GetBool("raw")
	if raw {
		for _, fileID := range args {
			body, err := c.apiClient.ReadFileRaw(ctx, fileID)
			if body != nil {
				fmt.Print(string(body))
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Retrieve the file(s):
	for _, fileID := range args {
		file, err := c.apiClient.ReadFile(ctx, fileID)
		if err != nil {
			return err
		}
		j, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshalling JSON file: %w", err)
		}
		fmt.Println(string(j))
	}
	return nil
}

// upload a JSONL fine-tuning file.
func (c *FileCommand) upload(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	purpose := cmd.Flag("purpose").Value.String()
	path := args[0]
	fileName := filepath.Base(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("upload file %s: %w", path, err)
	}
	file, err := c.apiClient.UploadFile(ctx, fileName, purpose, data)
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON file: %w", err)
	}
	fmt.Println(string(j))
	return nil
}

// download the specified file.
func (c *FileCommand) download(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	path := cmd.Flag("output").Value.String()
	fileID := args[0]
	if path == "" {
		// Read the file metadata to get the file name:
		file, err := c.apiClient.ReadFile(ctx, fileID)
		if err != nil {
			return fmt.Errorf("download file %s: %w", fileID, err)
		}
		path = file.FileName
	}
	// Download the file:
	data, err := c.apiClient.DownloadFile(ctx, fileID)
	if err != nil {
		return err
	}
	// Write the file to disk:
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("download file %s: %w", fileID, err)
	}
	fmt.Println("Downloaded file:", path)
	return nil
}

// delete the specified file.
func (c *FileCommand) delete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	raw, _ := cmd.Flags().GetBool("raw")
	for _, fileID := range args {
		if raw {
			body, err := c.apiClient.DeleteFileRaw(ctx, fileID)
			if body != nil {
				fmt.Print(string(body))
			}
			if err != nil {
				return err
			}
		} else {
			err := c.apiClient.DeleteFile(ctx, fileID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
