// Package main is the entry point for slack-fast-mcp.
// Without arguments, it starts the MCP Server (stdio transport).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kai-ko/slack-fast-mcp/internal/config"
	apperr "github.com/kai-ko/slack-fast-mcp/internal/errors"
	mcpserver "github.com/kai-ko/slack-fast-mcp/internal/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// version はビルド時に ldflags で注入される。
var version = "dev"

func main() {
	// バージョンをMCPサーバーに渡す
	mcpserver.Version = version

	// Graceful shutdown 用の context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// 設定読み込み
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.Load(cwd)
	if err != nil {
		return err
	}

	// トークン検証
	if err := cfg.Validate(); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			fmt.Fprintf(os.Stderr, "\n%s\n\n", appErr.FormatForMCP())
			fmt.Fprintf(os.Stderr, "Quick setup:\n")
			fmt.Fprintf(os.Stderr, "  1. Set SLACK_BOT_TOKEN environment variable\n")
			fmt.Fprintf(os.Stderr, "  2. Or create .slack-mcp.json with: {\"token\": \"${SLACK_BOT_TOKEN}\"}\n")
			fmt.Fprintf(os.Stderr, "  3. Or run: slack-fast-mcp setup\n\n")
		}
		return err
	}

	// MCP Server 起動
	s := mcpserver.NewServer(cfg)

	// stdio transport で起動
	stdioServer := server.NewStdioServer(s)
	return stdioServer.Listen(ctx, os.Stdin, os.Stdout)
}
