package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"docs-cli/pkg/mcp"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	batchStopOnError bool
)

// BatchStep represents a single operation in a manifest.
type BatchStep struct {
	ID        string          `json:"id"`
	Tool      string          `json:"tool"`
	Arguments json.RawMessage `json:"arguments"`
}

// BatchStepResult represents the outcome of a batch step.
type BatchStepResult struct {
	ID         string `json:"id"`
	Tool       string `json:"tool"`
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
	Result     any    `json:"result,omitempty"`
	Error      string `json:"error,omitempty"`
}

// BatchRunResult aggregates outcomes of all batch steps.
type BatchRunResult struct {
	TotalSteps int               `json:"total_steps"`
	Succeeded  int               `json:"succeeded"`
	Failed     int               `json:"failed"`
	DurationMs int64             `json:"duration_ms"`
	Steps      []BatchStepResult `json:"steps"`
}

var batchCmd = &cobra.Command{
	Use:   "batch [manifest.json]",
	Short: "Execute a sequence of tool operations from a JSON manifest file",
	Long: `Executes a batch manifest of operations sequentially using the native tool registry.
Ideal for multi-step AI agent workflows, data extraction pipelines, and automated document jobs.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestPath := args[0]
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("failed reading manifest %s: %w", manifestPath, err)
		}

		// Strip UTF-8 BOM if present
		data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
		// Decode UTF-16 LE BOM if created by PowerShell redirect
		if len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe {
			u16 := make([]uint16, (len(data)-2)/2)
			for i := range u16 {
				u16[i] = uint16(data[2+i*2]) | (uint16(data[2+i*2+1]) << 8)
			}
			data = []byte(string(utf16.Decode(u16)))
		}

		var steps []BatchStep
		if err := json.Unmarshal(data, &steps); err != nil {
			return fmt.Errorf("invalid manifest JSON in %s: %w", manifestPath, err)
		}

		startTime := time.Now()
		batchRes := BatchRunResult{
			TotalSteps: len(steps),
		}

		for idx, step := range steps {
			stepID := step.ID
			if stepID == "" {
				stepID = fmt.Sprintf("step_%d", idx+1)
			}

			stepStart := time.Now()
			rawOut, isErr := mcp.Dispatch(step.Tool, step.Arguments)
			stepDuration := time.Since(stepStart).Milliseconds()

			stepRes := BatchStepResult{
				ID:         stepID,
				Tool:       step.Tool,
				Success:    !isErr,
				DurationMs: stepDuration,
			}

			if isErr {
				stepRes.Error = rawOut
				batchRes.Failed++
			} else {
				// Try to parse result as JSON for rich output
				var parsed any
				if err := json.Unmarshal([]byte(rawOut), &parsed); err == nil {
					stepRes.Result = parsed
				} else {
					stepRes.Result = rawOut
				}
				batchRes.Succeeded++
			}

			batchRes.Steps = append(batchRes.Steps, stepRes)

			if isErr && batchStopOnError {
				break
			}
		}

		batchRes.DurationMs = time.Since(startTime).Milliseconds()

		stats := &output.Stats{
			DurationMs:     batchRes.DurationMs,
			FilesProcessed: batchRes.TotalSteps,
			MatchesFound:   batchRes.Succeeded,
		}

		printCmdResponse(cmd, output.SuccessResponse("batch.run", batchRes, stats), func() string {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("Executed batch manifest %s (%d/%d succeeded in %dms):\n\n",
				manifestPath, batchRes.Succeeded, batchRes.TotalSteps, batchRes.DurationMs))
			for _, s := range batchRes.Steps {
				status := "PASS"
				if !s.Success {
					status = fmt.Sprintf("FAIL (%s)", s.Error)
				}
				b.WriteString(fmt.Sprintf("[%s] %s (%s) - %dms\n", status, s.ID, s.Tool, s.DurationMs))
			}
			return b.String()
		})

		return nil
	},
}

func init() {
	batchCmd.Flags().BoolVar(&batchStopOnError, "stop-on-error", false, "Halt execution on the first step failure")

	rootCmd.AddCommand(batchCmd)
}
