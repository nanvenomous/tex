package cmd

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

var (
	ollamaClient *api.Client
)

// roundTripper wraps an http.RoundTripper and adds common headers to each request.
type roundTripper struct {
	rt http.RoundTripper
}

// RoundTrip implements the RoundTripper interface. It clones the request, adds the common headers, then forwards it.
func (a *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	for hdr, val := range config.Headers {
		req.Header.Set(hdr, val)
	}

	return a.rt.RoundTrip(req)
}

// chatWithOllama sends a single chat message and returns the response.
func chatWithOllama(ctx context.Context, model, prompt string, chatResFile *os.File) error {
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

// setupChat does the setup for the ollama chat file
func setupChat() error {
	rt := &roundTripper{rt: http.DefaultTransport}
	httpClnt := http.Client{Transport: rt}
	ollamaURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return err
	}
	ollamaClient = api.NewClient(ollamaURL, &httpClnt)
	return nil
}
