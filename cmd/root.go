/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

var (
	// baseURL = "https://ollama.fiore.one"
	baseURL = "http://127.0.0.1:11434"
)

// ChatWithOllama sends a single chat message and returns the response.
func ChatWithOllama(ctx context.Context, model, prompt string) error {
	httpClnt := http.Client{}
	ollamaURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	ollamaClient := api.NewClient(ollamaURL, &httpClnt)

	err = ollamaClient.Chat(ctx, &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}, func(chtRes api.ChatResponse) error {
		fmt.Println(chtRes.Message.Content)
		return nil
	})
	return err
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tex",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ChatWithOllama(context.TODO(), "llama3.2", "hello")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
