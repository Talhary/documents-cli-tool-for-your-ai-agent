package cmd

import (
	"fmt"
	"os"
	"runtime"

	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	workersCount int
)

var rootCmd = &cobra.Command{
	Use:   "agentdoc",
	Short: "agentdoc: High-performance concurrent CLI suite for AI agents",
	Long: `agentdoc provides AI agents with fast, concurrent tools to inspect,
search, manipulate, edit, convert, and snapshot documents, code, spreadsheets,
PDFs, and images with structured JSON responses.`,
	Version: "dev",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if workersCount <= 0 {
			workersCount = runtime.NumCPU()
		}
	},
}

// SetVersion sets the CLI version reported by --version (injected from main).
func SetVersion(v string) {
	if v != "" {
		rootCmd.Version = v
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if output.JSONMode {
			output.PrintResponse(os.Stdout, output.ErrorResponse(rootCmd.Name(), err), nil)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

// printCmdResponse prints the response to the command's configured output writer.
func printCmdResponse(cmd *cobra.Command, resp output.Response, plainFallback func() string) {
	output.PrintResponse(cmd.OutOrStdout(), resp, plainFallback)
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&output.JSONMode, "json", false, "Output results as structured JSON")
	rootCmd.PersistentFlags().BoolVar(&output.Silent, "silent", false, "Suppress non-essential progress output")
	rootCmd.PersistentFlags().IntVarP(&workersCount, "workers", "w", runtime.NumCPU(), "Number of concurrent worker goroutines")
}
