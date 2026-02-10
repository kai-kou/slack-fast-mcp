package cli

import (
	"os"

	mcpserver "github.com/kai-kou/slack-fast-mcp/internal/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// newServeCmd は serve サブコマンドを作成する。
func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start MCP Server in stdio mode",
		Long:  "Start the MCP Server using stdio transport. This is the default mode when no subcommand is specified.",
		RunE:  runServe,
	}
}

// runServe は MCP Server モードで起動する。
func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	// バージョンをMCPサーバーに渡す
	mcpserver.Version = Version

	// MCP Server 起動
	s := mcpserver.NewServer(cfg)
	stdioServer := server.NewStdioServer(s)
	return stdioServer.Listen(cmd.Context(), os.Stdin, os.Stdout)
}
