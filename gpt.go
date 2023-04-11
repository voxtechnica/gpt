package main

import (
	"fmt"
	"gpt/cli"
	"gpt/openai"
	"os"

	"github.com/spf13/viper"
)

// main is the entry point for the application.
func main() {
	var orgID string
	var apiKey string

	// Application configuration
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err == nil {
		orgID = viper.GetString("OPENAI_ORG_ID")
		apiKey = viper.GetString("OPENAI_API_KEY")
	}

	// Initialize the API client:
	apiClient := openai.NewClient(orgID, apiKey)

	// Initialize the Command Line Interface:
	rootCmd := cli.NewRootCommand(apiClient)

	// Execute the specified command:
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
