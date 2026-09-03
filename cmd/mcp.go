package cmd

import (
	"os"

	"docs-cli/pkg/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start native Model Context Protocol (MCP) JSON-RPC stdio server",
	Long: `Starts an MCP stdio server communicating over standard I/O via JSON-RPC 2.0.
Can be directly configured in Claude Desktop, Cursor, Antigravity, Cline, Roo Code,
or any MCP-compatible AI agent harness to expose all agentdoc tools automatically.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcp.RunServer(os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
