```.go
/*
Copyright © 2025 nanvenomous mrgarelli@gmail.com
*/
package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

const (
	texDir  = "context"
	baseURL = "http://127.0.0.1:11434"
	// model      = "deepseek-r1:32b"
	// model      = "llama3.2"
	model = "deepseek-coder-v2:16b"
)

var (
	ollamaClient *api.Client
	// baseURL = "https://ollama.fiore.one"
	flagFiles    []string
	editorEnvVar = os.Getenv("EDITOR")
	requestFile  = filepath.Join(texDir, "request.md")
	responseFile = filepath.Join(texDir, "response.md")
)

func editor(flPth string) error {
	cmd := exec.Command(editorEnvVar, flPth)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func editRequestFile(prompt string) (string, error) {
	reqFl, err := os.Create(requestFile)
	if err != nil {
		return prompt, err
	}
	defer reqFl.Close()
	_, err = reqFl.Write([]byte(prompt))
	if err != nil {
		return prompt, err
	}
	err = editor(requestFile)
	if err != nil {
		return prompt, err
	}

	reqFlByt, err := os.ReadFile(requestFile)
	if err != nil {
		return prompt, err
	}

	return string(reqFlByt), nil
}

// rootCmd calls the tex ai agent
var rootCmd = &cobra.Command{
	Use:   "tex",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		var prompt string

		if len(flagFiles) > 0 {
			for _, fl := range flagFiles {
				ext := filepath.Ext(fl)
				flByt, err := os.ReadFile(fl)
				if err != nil {
					fmt.Printf("Failed to parse file %s: %v \n", fl, err)
					continue
				}
				prompt += "```" + ext + "\n"
				prompt += string(flByt) + "\n"
				prompt += "```" + "\n"
			}
		}

		if len(args) == 1 {
			prompt = prompt + "\n" + args[0]
		} else if len(args) > 1 {
			prompt = prompt + "\n" + strings.Join(args, " ")
		} else {
			var err error
			prompt, err = editRequestFile(prompt)
			if err != nil {
				return err
			}
		}

		return chatWithOllama(cmd.Context(), model, prompt)
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
	os.Mkdir(texDir, os.ModePerm)

	rootCmd.Flags().StringSliceVarP(&flagFiles, "file", "f", []string{}, "add a slice of files to the context")

	httpClnt := http.Client{}
	ollamaURL, err := url.Parse(baseURL)
	cobra.CheckErr(err)
	ollamaClient = api.NewClient(ollamaURL, &httpClnt)
}

```

how could I set the default completion of the root command to no completion at all?
