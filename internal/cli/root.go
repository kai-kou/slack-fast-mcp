// Package cli implements the CLI layer using cobra for slack-fast-mcp.
// Provides subcommands: serve, post, history, reply, setup, version.
package cli

import (
	"fmt"
	"os"

	"github.com/kai-kou/slack-fast-mcp/internal/config"
	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
	"github.com/spf13/cobra"
)

// Version はアプリケーションバージョン（ビルド時に ldflags で注入）。
var Version = "dev"

// グローバルフラグ
var (
	flagConfig  string
	flagToken   string
	flagChannel string
	flagVerbose bool
	flagJSON    bool
)

// NewRootCmd はルートコマンドを作成する。
// 引数なしの場合は MCP Server モード（serve）で起動する。
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "slack-fast-mcp",
		Short: "Fast Slack MCP Server & CLI",
		Long: `slack-fast-mcp - A high-performance Slack MCP Server written in Go.

Use as an MCP Server (default) or as a standalone CLI tool.

Without any subcommand, starts the MCP Server in stdio mode.`,
		// 引数なしの場合は serve（MCP Server モード）
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// グローバルフラグ
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "config file path (default: .slack-mcp.json)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Slack Bot Token (overrides config/env)")
	rootCmd.PersistentFlags().StringVar(&flagChannel, "channel", "", "channel name or ID")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output in JSON format")

	// サブコマンド登録
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newPostCmd())
	rootCmd.AddCommand(newHistoryCmd())
	rootCmd.AddCommand(newReplyCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newSetupCmd())

	return rootCmd
}

// loadConfig はグローバルフラグを考慮して設定を読み込む。
func loadConfig() (*config.Config, error) {
	var cfg *config.Config
	var err error

	if flagConfig != "" {
		// --config で明示的にパスが指定された場合
		cfg, err = config.LoadFromPath(flagConfig)
	} else {
		cwd, e := os.Getwd()
		if e != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", e)
		}
		cfg, err = config.Load(cwd)
	}
	if err != nil {
		return nil, err
	}

	// --token フラグで上書き
	if flagToken != "" {
		cfg.Token = flagToken
	}

	// --verbose でログレベル上書き
	if flagVerbose {
		cfg.LogLevel = "debug"
	}

	return cfg, nil
}

// clientFactory はSlackClientの生成関数。テスト時にモック注入するためのフック。
var clientFactory func(token string) slackclient.SlackClient

// newSlackClient はSlackClientを生成する。clientFactoryが設定されていればそれを使用する。
func newSlackClient(token string) slackclient.SlackClient {
	if clientFactory != nil {
		return clientFactory(token)
	}
	return slackclient.NewClient(token)
}

// loadConfigAndClient は設定読み込み + バリデーション + Slackクライアント作成をまとめて行う。
func loadConfigAndClient() (*config.Config, slackclient.SlackClient, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, err
	}

	client := newSlackClient(cfg.Token)
	return cfg, client, nil
}

// resolveChannel はフラグ → 設定ファイルの優先度でチャンネルを解決する。
func resolveChannel(cfg *config.Config) (string, error) {
	return cfg.ResolveChannel(flagChannel)
}
