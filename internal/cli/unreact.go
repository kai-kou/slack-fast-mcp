package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// newUnreactCmd は unreact サブコマンドを作成する。
func newUnreactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unreact",
		Short: "Remove a reaction (emoji) from a message",
		Long:  "Remove an emoji reaction from a Slack message. Can only remove reactions added by the bot.",
		Example: `  # Remove a thumbsup reaction
  slack-fast-mcp unreact --timestamp 1234567890.123456 --reaction thumbsup

  # Remove a reaction in a specific channel
  slack-fast-mcp unreact --channel general --timestamp 1234567890.123456 --reaction heart

  # Remove a reaction with JSON output
  slack-fast-mcp unreact --timestamp 1234567890.123456 --reaction eyes --json`,
		RunE: runUnreact,
	}

	cmd.Flags().StringVarP(&flagTimestamp, "timestamp", "t", "", "message timestamp to remove reaction from (required)")
	cmd.Flags().StringVarP(&flagReaction, "reaction", "r", "", "emoji name without colons (required)")
	_ = cmd.MarkFlagRequired("timestamp")
	_ = cmd.MarkFlagRequired("reaction")

	return cmd
}

// runUnreact はメッセージからリアクションを削除する。
func runUnreact(cmd *cobra.Command, args []string) error {
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

	// コロン付きの絵文字名を正規化
	reaction := strings.Trim(flagReaction, ":")

	result, err := client.RemoveReaction(ctx, channel, flagTimestamp, reaction)
	if err != nil {
		return err
	}

	if flagJSON {
		out := map[string]interface{}{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"timestamp":    result.Timestamp,
			"reaction":     result.Reaction,
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(out)
	}

	// テキスト形式出力
	fmt.Fprintf(cmd.OutOrStdout(), "✅ Reaction :%s: removed from message in #%s\n", result.Reaction, result.ChannelName)

	return nil
}
