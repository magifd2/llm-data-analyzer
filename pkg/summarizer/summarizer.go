package summarizer

import (
	"context"
	"fmt"
	"strings"

	"llm-data-analyzer/pkg/llm"
	"llm-data-analyzer/pkg/splitter"
	"github.com/spf13/cobra"
)

const maxIterations = 10 // Add a constant for the max number of iterations

// Summarizer handles the recursive summarization of text.
type Summarizer struct {
	client    *llm.Client
	splitter  *splitter.Splitter
	chunkSize int
	verbose   bool
	cmd       *cobra.Command
}

// NewSummarizer creates a new Summarizer.
func NewSummarizer(client *llm.Client, chunkSize int, verbose bool, cmd *cobra.Command) (*Summarizer, error) {
	s, err := splitter.NewSplitter(chunkSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create splitter: %w", err)
	}
	return &Summarizer{
		client:    client,
		splitter:  s,
		chunkSize: chunkSize,
		verbose:   verbose,
		cmd:       cmd,
	},
	nil
}

// Summarize performs recursive summarization if the text is too long.
func (s *Summarizer) Summarize(ctx context.Context, text, prompt string) (string, error) {
	currentText := text
	iteration := 1

	for {
		if iteration > maxIterations {
			return "", fmt.Errorf("summarization reached max iterations (%d), potential infinite loop", maxIterations)
		}

		if s.verbose {
			s.cmd.Printf("Summarizer iteration %d: text length %d\n", iteration, len(currentText))
		}
		iteration++

		textTokens := s.splitter.Encode(currentText)
		promptTokens := s.splitter.Encode(prompt)

		if len(textTokens)+len(promptTokens) < s.chunkSize {
			if s.verbose {
				s.cmd.Println("Text is small enough, performing final analysis.")
			}
			fullPrompt := fmt.Sprintf("%s\n\n--- Text to Summarize ---\n%s", prompt, currentText)
			return s.client.Analyze(ctx, fullPrompt)
		}

		chunks, err := s.splitter.Split(strings.NewReader(currentText))
		if err != nil {
			return "", fmt.Errorf("failed to split text for summarization: %w", err)
		}

		if len(chunks) == 1 {
			if s.verbose {
				s.cmd.Println("Cannot split further, analyzing the whole text.")
			}
			fullPrompt := fmt.Sprintf("%s\n\n--- Text to Summarize ---\n%s", prompt, currentText)
			return s.client.Analyze(ctx, fullPrompt)
		}

		if s.verbose {
			s.cmd.Printf("Splitting text into %d chunks for summarization.\n", len(chunks))
		}

		var summaries []string
		for i, chunk := range chunks {
			if s.verbose {
				s.cmd.Printf("Summarizing sub-chunk %d/%d\n", i+1, len(chunks))
			}
			fullPrompt := fmt.Sprintf("%s\n\n--- Text to Summarize ---\n%s", prompt, chunk)
			summary, err := s.client.Analyze(ctx, fullPrompt)
			if err != nil {
				return "", err
			}
			summaries = append(summaries, summary)
		}

		currentText = strings.Join(summaries, "\n\n---\n\n")
	}
}
