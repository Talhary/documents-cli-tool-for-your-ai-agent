package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"docs-cli/pkg/mcp"
	"docs-cli/pkg/output"
	"github.com/spf13/cobra"
)

var (
	schemaFormat string
	schemaTool   string
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Export tool schemas formatted for OpenAI, Anthropic, or MCP",
	Long: `Exports tool definitions and argument schemas in formats ready to paste directly
into OpenAI Function Calling, Anthropic Claude Tool Use, or MCP client configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tools := mcp.GetToolDefinitions()

		if schemaTool != "" {
			var filtered []mcp.Tool
			for _, t := range tools {
				if strings.EqualFold(t.Name, schemaTool) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				return fmt.Errorf("tool %q not found in schema registry", schemaTool)
			}
			tools = filtered
		}

		var result any
		switch strings.ToLower(schemaFormat) {
		case "openai":
			type openAIFunc struct {
				Type     string      `json:"type"`
				Function map[string]any `json:"function"`
			}
			var list []openAIFunc
			for _, t := range tools {
				list = append(list, openAIFunc{
					Type: "function",
					Function: map[string]any{
						"name":        t.Name,
						"description": t.Description,
						"parameters":  t.InputSchema,
					},
				})
			}
			result = list

		case "anthropic":
			type anthropicTool struct {
				Name        string          `json:"name"`
				Description string          `json:"description"`
				InputSchema mcp.InputSchema `json:"input_schema"`
			}
			var list []anthropicTool
			for _, t := range tools {
				list = append(list, anthropicTool{
					Name:        t.Name,
					Description: t.Description,
					InputSchema: t.InputSchema,
				})
			}
			result = list

		case "mcp":
			fallthrough
		default:
			result = tools
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

		if output.JSONMode {
			printCmdResponse(cmd, output.SuccessResponse("schema", result, nil), nil)
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	},
}

func init() {
	schemaCmd.Flags().StringVarP(&schemaFormat, "format", "f", "mcp", "Schema format: 'mcp', 'openai', or 'anthropic'")
	schemaCmd.Flags().StringVarP(&schemaTool, "tool", "t", "", "Filter to a specific tool name (optional)")

	rootCmd.AddCommand(schemaCmd)
}
