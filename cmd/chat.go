package cmd

import (
	"context"
	"os"

	"github.com/ollama/ollama/api"
)

// chatWithOllama sends a single chat message and returns the response.
func chatWithOllama(ctx context.Context, model, prompt string, chatResFile *os.File) error {
	// tmpFl, err := os.Create(responseFile)
	// tmpFl, err := createTmpFile(responseFile)
	// if err != nil {
	// 	return err
	// }
	// defer tmpFl.Close()
	// fmt.Println(tmpFl.Name())

	err := ollamaClient.Chat(ctx, &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}, func(chtRes api.ChatResponse) error {
		_, err := chatResFile.Write([]byte(chtRes.Message.Content))
		return err
	})
	return err
}
