package cmd

import (
	"testing"
)

func TestSimple(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}
}
