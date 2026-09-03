package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Stats tracks telemetry useful for AI agents to assess tool cost and scale.
type Stats struct {
	DurationMs     int64 `json:"duration_ms"`
	FilesProcessed int   `json:"files_processed,omitempty"`
	MatchesFound   int   `json:"matches_found,omitempty"`
	LinesAffected  int   `json:"lines_affected,omitempty"`
	BytesRead      int64 `json:"bytes_read,omitempty"`
	BytesWritten   int64 `json:"bytes_written,omitempty"`
}

// Response represents the standardized machine-readable envelope.
type Response struct {
	Success bool   `json:"success"`
	Command string `json:"command,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Stats   *Stats `json:"stats,omitempty"`
}

// Global flags
var (
	JSONMode bool
	Silent   bool
)

// PrintResponse writes the response according to configured mode.
func PrintResponse(w io.Writer, resp Response, plainFallback func() string) {
	if JSONMode {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		return
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		return
	}

	if plainFallback != nil {
		content := plainFallback()
		if content != "" {
			fmt.Fprintln(w, content)
		}
	} else if resp.Message != "" {
		fmt.Fprintln(w, resp.Message)
	}
}

// SuccessResponse creates a success response.
func SuccessResponse(command string, data any, stats *Stats) Response {
	return Response{
		Success: true,
		Command: command,
		Data:    data,
		Stats:   stats,
	}
}

// ErrorResponse creates an error response.
func ErrorResponse(command string, err error) Response {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return Response{
		Success: false,
		Command: command,
		Error:   errMsg,
	}
}
