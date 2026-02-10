package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
	"github.com/spf13/cobra"
)

var (
	flagLimit  int
	flagOldest string
	flagLatest string
)

// newHistoryCmd は history サブコマンドを作成する。
func newHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get message history from a Slack channel",
		Long:  "Retrieve recent messages from a Slack channel.",
		Example: `  # Get last 10 messages from default channel
  slack-fast-mcp history

  # Get last 20 messages from specific channel
  slack-fast-mcp history --channel general --limit 20

  # Get messages in JSON format
  slack-fast-mcp history --channel general --json`,
		RunE: runHistory,
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "l", 10, "number of messages to retrieve (1-100)")
	cmd.Flags().StringVar(&flagOldest, "oldest", "", "start time (Unix timestamp)")
	cmd.Flags().StringVar(&flagLatest, "latest", "", "end time (Unix timestamp)")

	return cmd
}

// runHistory はチャンネル履歴を取得する。
func runHistory(cmd *cobra.Command, args []string) error {
	cfg, client, err := loadConfigAndClient()
	if err != nil {
		return err
	}

	channel, err := resolveChannel(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()

	result, err := client.GetHistory(ctx, channel, slackclient.HistoryOptions{
		Limit:  flagLimit,
		Oldest: flagOldest,
		Latest: flagLatest,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		out := map[string]interface{}{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"messages":     result.Messages,
			"has_more":     result.HasMore,
			"count":        result.Count,
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(out)
	}

	// テーブル形式出力
	fmt.Fprintf(cmd.OutOrStdout(), "📋 #%s — %d messages\n", result.ChannelName, result.Count)
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("─", 60))

	for i, msg := range result.Messages {
		userName := msg.UserName
		if userName == "" {
			userName = msg.User
		}

		// メッセージ本文（長すぎる場合は切り詰め）
		text := msg.Text
		if len(text) > 200 {
			text = text[:200] + "..."
		}

		fmt.Fprintf(cmd.OutOrStdout(), "  @%-16s %s\n", userName, text)

		// スレッド情報
		if msg.ReplyCount > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s💬 %d replies (thread_ts: %s)\n", strings.Repeat(" ", 18), msg.ReplyCount, msg.TS)
		}

		if i < len(result.Messages)-1 {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", strings.Repeat("·", 56))
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", strings.Repeat("─", 60))
	if result.HasMore {
		fmt.Fprintf(cmd.OutOrStdout(), "  (more messages available — use --limit to load more)\n")
	}

	return nil
}
