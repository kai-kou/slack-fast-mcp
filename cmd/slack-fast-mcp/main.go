// Package main is the entry point for slack-fast-mcp.
// Without arguments, it starts the MCP Server (stdio transport).
// With subcommands (post, history, reply, setup, version), operates as a CLI tool.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kai-ko/slack-fast-mcp/internal/cli"
	apperr "github.com/kai-ko/slack-fast-mcp/internal/errors"
)

// version はビルド時に ldflags で注入される。
var version = "dev"

func main() {
	// バージョンを CLI パッケージに渡す
	cli.Version = version

	// Graceful shutdown 用の context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	rootCmd := cli.NewRootCmd()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// AppError の場合は詳細なエラーメッセージを表示
		if appErr, ok := err.(*apperr.AppError); ok {
			fmt.Fprintf(os.Stderr, "\n❌ Error [%s]: %s\n", appErr.Code, appErr.Message)
			if appErr.Hint != "" {
				fmt.Fprintf(os.Stderr, "💡 %s\n\n", appErr.Hint)
			}
		} else {
			fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n\n", err)
		}
		os.Exit(1)
	}
}
