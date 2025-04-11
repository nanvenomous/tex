package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
)

// ChatWithOllama sends a single chat message and returns the response.
func ChatWithOllama(ctx context.Context, model, prompt string) error {
	httpClnt := http.Client{}
	ollamaURL, err := url.Parse("https://ollama.fiore.one")
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

func main() {
	err := ChatWithOllama(context.TODO(), "llama3.2", "hello")
	if err != nil {
		panic(err)
	}
}
