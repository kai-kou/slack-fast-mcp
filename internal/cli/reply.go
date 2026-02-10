package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var flagThreadTS string

// newReplyCmd は reply サブコマンドを作成する。
func newReplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to a thread in a Slack channel",
		Long:  "Post a reply to an existing message thread. Supports Slack mrkdwn formatting.",
		Example: `  # Reply to a thread
  slack-fast-mcp reply --thread-ts 1234567890.123456 --message "Reply here"

  # Reply to a thread in a specific channel
  slack-fast-mcp reply --channel general --thread-ts 1234567890.123456 --message "Reply"

  # Reply with JSON output
  slack-fast-mcp reply --thread-ts 1234567890.123456 --message "Reply" --json`,
		RunE: runReply,
	}

	cmd.Flags().StringVarP(&flagThreadTS, "thread-ts", "t", "", "thread timestamp to reply to (required)")
	cmd.Flags().StringVarP(&flagMessage, "message", "m", "", "reply message text (required)")
	_ = cmd.MarkFlagRequired("thread-ts")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

// runReply はスレッドに返信を投稿する。
func runReply(cmd *cobra.Command, args []string) error {
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

	result, err := client.PostThread(ctx, channel, flagThreadTS, flagMessage)
	if err != nil {
		return err
	}

	if flagJSON {
		out := map[string]interface{}{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"ts":           result.TS,
			"thread_ts":    result.ThreadTS,
			"message":      result.Message,
			"permalink":    result.Permalink,
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(out)
	}

	// テキスト形式出力
	fmt.Fprintf(cmd.OutOrStdout(), "✅ Reply posted to thread in #%s\n", result.ChannelName)
	if result.Permalink != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "🔗 %s\n", result.Permalink)
	}

	return nil
}
