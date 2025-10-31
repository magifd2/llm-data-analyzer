/*
Copyright © 2025 Gemini CLI

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"llm-data-analyzer/pkg/config"
	"llm-data-analyzer/pkg/llm"
	"llm-data-analyzer/pkg/splitter"
	"llm-data-analyzer/pkg/summarizer"

	"github.com/spf13/cobra"
)

var (
	cfgFile            string
	endpointName       string
	analysisPromptFile string
	summaryPromptFile  string
	outputFile         string
	tempDir            string
	keepTempDir        bool
	verbose            bool
	isJSONL            bool

	appConfig config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "llm-data-analyzer [flags] <input_file_path>",
	Short: "A CLI tool to analyze large text files using LLMs.",
	Long: `llm-data-analyzer is a command-line tool that analyzes large text or JSONL files.

It breaks down the input file into smaller chunks that fit within the context window of a specified Large Language Model (LLM).
Each chunk is analyzed individually, and the results are then summarized to produce a final, consolidated report.`,
	Args: cobra.ExactArgs(1), // Expect exactly one argument
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config here
		var err error
		appConfig, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("Error loading config file: %w", err)
		}

		// 1. Get input file path
		inputFile := args[0]

		// 2. Select endpoint configuration
		var endpointConf config.EndpointConfig
		for _, ep := range appConfig.Endpoints {
			if ep.Name == endpointName {
				endpointConf = ep
				break
			}
		}
		if endpointConf.Name == "" {
			return fmt.Errorf("endpoint '%s' not found in config file", endpointName)
		}

		// 3. Create and manage temporary directory
		workDir := tempDir
		if workDir == "" {
			workDir, err = os.MkdirTemp("", "llm-analyzer-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory: %w", err)
			}
			if !keepTempDir {
				defer os.RemoveAll(workDir)
			}
		}
		if verbose {
			cmd.Printf("Using temporary directory: %s\n", workDir)
		}

		// 4. Open input file
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer file.Close()

		// 5. Create splitter and split the file
		s, err := splitter.NewSplitter(endpointConf.ChunkSize)
		if err != nil {
			return fmt.Errorf("failed to create splitter: %w", err)
		}

		var chunks []string
		if isJSONL {
			chunks, err = s.SplitJSONL(file)
		} else {
			chunks, err = s.Split(file)
		}
		if err != nil {
			return fmt.Errorf("failed to split file: %w", err)
		}

		if verbose {
			cmd.Printf("File split into %d chunks.\n", len(chunks))
		}

		// 6. Create LLM client
		apiKey := os.Getenv(endpointConf.APIKeyEnv)
		if apiKey == "" {
			return fmt.Errorf("API key environment variable '%s' not set", endpointConf.APIKeyEnv)
		}
		client := llm.NewClient(endpointConf.EndpointURL, apiKey, endpointConf.Model)

		// 7. Read analysis prompt
		promptBytes, err := os.ReadFile(analysisPromptFile)
		if err != nil {
			return fmt.Errorf("failed to read analysis prompt file: %w", err)
		}
		analysisPrompt := string(promptBytes)

		// 8. Analyze all chunks in parallel
		var wg sync.WaitGroup
		errorChan := make(chan error, len(chunks))

		for i, chunk := range chunks {
			wg.Add(1)
			go func(chunkIndex int, chunkText string) {
				defer wg.Done()

				fullPrompt := fmt.Sprintf("%s\n\n---\n%s", analysisPrompt, chunkText)
				if verbose {
					cmd.Printf("Analyzing chunk %d...\n", chunkIndex+1)
				}

				result, err := client.Analyze(context.Background(), fullPrompt)
				if err != nil {
					errorChan <- fmt.Errorf("failed to analyze chunk %d: %w", chunkIndex+1, err)
					return
				}

				// Save result to temporary file
				resultFile := filepath.Join(workDir, fmt.Sprintf("chunk_%d.txt", chunkIndex+1))
				if err := os.WriteFile(resultFile, []byte(result), 0644); err != nil {
					errorChan <- fmt.Errorf("failed to write result for chunk %d: %w", chunkIndex+1, err)
				}
			}(i, chunk)
		}

		wg.Wait()
		close(errorChan)

		// Check for errors during chunk analysis
		for err := range errorChan {
			return err // Return the first error encountered
		}

		if verbose {
			cmd.Println("\nAll chunks analyzed successfully.")
		}

		// 9. Combine results and generate final summary
		if verbose {
			cmd.Println("Combining results and generating final summary...")
		}

		files, err := os.ReadDir(workDir)
		if err != nil {
			return fmt.Errorf("failed to read temporary directory: %w", err)
		}

		var combinedResults strings.Builder
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(workDir, file.Name()))
			if err != nil {
				return fmt.Errorf("failed to read intermediate file %s: %w", file.Name(), err)
			}
			combinedResults.Write(content)
			combinedResults.WriteString("\n\n---\n\n")
		}

		summaryPromptBytes, err := os.ReadFile(summaryPromptFile)
		if err != nil {
			return fmt.Errorf("failed to read summary prompt file: %w", err)
		}
		summaryPrompt := string(summaryPromptBytes)

		// Create a summarizer and generate the final report
		summarizer, err := summarizer.NewSummarizer(client, endpointConf.ContextWindowSize, verbose, cmd)
		if err != nil {
			return fmt.Errorf("failed to create summarizer: %w", err)
		}

		finalResult, err := summarizer.Summarize(context.Background(), combinedResults.String(), summaryPrompt)
		if err != nil {
			return fmt.Errorf("failed to generate final summary: %w", err)
		}

		// 10. Output final result
		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(finalResult), 0644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			if verbose {
				cmd.Printf("Final summary written to %s\n", outputFile)
			}
		} else {
			cmd.Println("\n--- Final Summary ---")
			cmd.Println(finalResult)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// SetVersion sets the version for the root command.
func SetVersion(version string) {
	rootCmd.Version = version
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "config file")
	rootCmd.PersistentFlags().StringVarP(&endpointName, "endpoint-name", "e", "", "Name of the LLM endpoint to use (defined in config file)")
	rootCmd.PersistentFlags().StringVar(&analysisPromptFile, "analysis-prompt-file", "", "Path to the data analysis prompt file (required)")
	rootCmd.PersistentFlags().StringVar(&summaryPromptFile, "summary-prompt-file", "", "Path to the summary prompt file (required)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Path to the output file (default is stdout)")
	rootCmd.PersistentFlags().StringVar(&tempDir, "temp-dir", "", "Path to the temporary directory for intermediate files")
	rootCmd.PersistentFlags().BoolVar(&keepTempDir, "keep-temp-dir", false, "Keep the temporary directory after execution")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&isJSONL, "jsonl", false, "Treat the input file as JSONL")

	// Mark required flags
	rootCmd.MarkPersistentFlagRequired("analysis-prompt-file")
	rootCmd.MarkPersistentFlagRequired("summary-prompt-file")
	rootCmd.MarkPersistentFlagRequired("endpoint-name")
}