// Package mcp implements the MCP Server with Slack tool handlers.
package mcp

import (
	"github.com/kai-kou/slack-fast-mcp/internal/config"
	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
	"github.com/mark3labs/mcp-go/server"
)

// Version はアプリケーションバージョン（ビルド時に ldflags で注入）。
var Version = "dev"

// NewServer は新しいMCP Serverを作成し、全ツールを登録する。
func NewServer(cfg *config.Config) *server.MCPServer {
	client := slackclient.NewClient(cfg.Token)
	return NewServerWithClient(cfg, client)
}

// NewServerWithClient は指定されたSlackClientを使用してMCP Serverを作成する（テスト用）。
func NewServerWithClient(cfg *config.Config, client slackclient.SlackClient) *server.MCPServer {
	s := server.NewMCPServer(
		"slack-fast-mcp",
		Version,
		server.WithToolCapabilities(false),
	)

	// ツール登録
	s.AddTool(postMessageTool(), postMessageHandler(client, cfg))
	s.AddTool(getHistoryTool(), getHistoryHandler(client, cfg))
	s.AddTool(postThreadTool(), postThreadHandler(client, cfg))
	s.AddTool(addReactionTool(), addReactionHandler(client, cfg))
	s.AddTool(removeReactionTool(), removeReactionHandler(client, cfg))

	return s
}
