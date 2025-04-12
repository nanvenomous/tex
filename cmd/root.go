/*
Copyright © 2025 nanvenomous mrgarelli@gmail.com
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// type config struct {
// 	TemporaryDirectory string `yaml:""`
// 	Model              string `yaml:""`
// 	BaseURL            string `yaml:""`
// }

const (
	tmpDir = ""
	model  = "llama3.2"
	// model = "deepseek-coder-v2:16b"
	// model   = "deepseek-r1:32b"
	baseURL = "https://ollama.fiore.one"
	// baseURL = "http://127.0.0.1:11434"
)

var (
	flagFiles    []string
	editorEnvVar = os.Getenv("EDITOR")
	requestFile  = "request_*.md"
	responseFile = "response_*.md"
)

// rootCmd calls the tex ai agent
var rootCmd = &cobra.Command{
	Use:   "tex",
	Short: "A brief description of your application",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []cobra.Completion{}, cobra.ShellCompDirectiveNoFileComp
	},
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
			err = withTempFile(requestFile, func(tmpFl *os.File) error {
				_, err = tmpFl.Write([]byte(prompt))
				if err != nil {
					return err
				}
				err = editor(tmpFl.Name())
				if err != nil {
					return err
				}

				reqFlByt, err := os.ReadFile(tmpFl.Name())
				if err != nil {
					return err
				}
				prompt = string(reqFlByt)
				return nil
			})
		}

		return withTempFile(responseFile, func(tmpFl *os.File) error {
			go chatWithOllama(cmd.Context(), model, prompt, tmpFl)

			return editor(tmpFl.Name())
		})
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
}
