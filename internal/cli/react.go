package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var flagTimestamp string
var flagReaction string

// newReactCmd は react サブコマンドを作成する。
func newReactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react",
		Short: "Add a reaction (emoji) to a message",
		Long:  "Add an emoji reaction to a Slack message. Use emoji names without colons (e.g. 'thumbsup', not ':thumbsup:').",
		Example: `  # Add a thumbsup reaction
  slack-fast-mcp react --timestamp 1234567890.123456 --reaction thumbsup

  # Add a reaction in a specific channel
  slack-fast-mcp react --channel general --timestamp 1234567890.123456 --reaction heart

  # Add a reaction with JSON output
  slack-fast-mcp react --timestamp 1234567890.123456 --reaction eyes --json`,
		RunE: runReact,
	}

	cmd.Flags().StringVarP(&flagTimestamp, "timestamp", "t", "", "message timestamp to react to (required)")
	cmd.Flags().StringVarP(&flagReaction, "reaction", "r", "", "emoji name without colons (required)")
	_ = cmd.MarkFlagRequired("timestamp")
	_ = cmd.MarkFlagRequired("reaction")

	return cmd
}

// runReact はメッセージにリアクションを追加する。
func runReact(cmd *cobra.Command, args []string) error {
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

	result, err := client.AddReaction(ctx, channel, flagTimestamp, reaction)
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
	fmt.Fprintf(cmd.OutOrStdout(), "✅ Reaction :%s: added to message in #%s\n", result.Reaction, result.ChannelName)

	return nil
}
