package cmd

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

var (
	ollamaClient *api.Client
)

// authRoundTripper wraps an http.RoundTripper and adds an Authorization header to each request.
type authRoundTripper struct {
	rt http.RoundTripper
}

// RoundTrip implements the RoundTripper interface. It clones the request, adds the Authorization header, then forwards it.
func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", os.Getenv("LIBERO_API_KEY"))

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

func init() {
	rt := &authRoundTripper{rt: http.DefaultTransport}
	httpClnt := http.Client{Transport: rt}
	ollamaURL, err := url.Parse(baseURL)
	cobra.CheckErr(err)
	ollamaClient = api.NewClient(ollamaURL, &httpClnt)
}
