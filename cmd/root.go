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
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

var (
	// baseURL = "https://ollama.fiore.one"
	ollamaClient *api.Client
	baseURL      = "http://127.0.0.1:11434"
	model        = "llama3.2"
	flagFiles    []string
)

// ChatWithOllama sends a single chat message and returns the response.
func ChatWithOllama(ctx context.Context, model, prompt string) error {
	// tmpFl, err := os.CreateTemp("", "tex_*.md")
	tmpFl, err := os.Create("/tmp/tex_response.md")
	if err != nil {
		return err
	}
	defer tmpFl.Close()
	fmt.Println(tmpFl.Name())

	err = ollamaClient.Chat(ctx, &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}, func(chtRes api.ChatResponse) error {
		// fmt.Println(chtRes.Message.Content)
		_, err := tmpFl.Write([]byte(chtRes.Message.Content))
		return err
	})
	return err
}

// rootCmd calls the tex ai agent
var rootCmd = &cobra.Command{
	Use:   "tex",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		var prompt string

		if len(flagFiles) > 0 {
			for _, fl := range flagFiles {
				flByt, err := os.ReadFile(fl)
				if err != nil {
					fmt.Printf("Failed to parse file %s: %v \n", fl, err)
					continue
				}
				prompt += string(flByt) + "\n"
			}
		}

		if len(args) == 1 {
			prompt = prompt + "\n" + args[0]
		} else if len(args) > 1 {
			prompt = prompt + "\n" + strings.Join(args, " ")
		}

		return ChatWithOllama(cmd.Context(), model, prompt)
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
	rootCmd.Flags().StringSliceVarP(&flagFiles, "file", "f", []string{}, "add a slice of files to the context")

	httpClnt := http.Client{}
	ollamaURL, err := url.Parse(baseURL)
	cobra.CheckErr(err)
	ollamaClient = api.NewClient(ollamaURL, &httpClnt)
}
