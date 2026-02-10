package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var flagMessage string

// newPostCmd は post サブコマンドを作成する。
func newPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a message to a Slack channel",
		Long:  "Post a message to a Slack channel. Supports Slack mrkdwn formatting.",
		Example: `  # Post to default channel
  slack-fast-mcp post --message "Hello World"

  # Post to specific channel
  slack-fast-mcp post --channel general --message "Hello World"

  # Post with JSON output
  slack-fast-mcp post --channel general --message "Hello" --json`,
		RunE: runPost,
	}

	cmd.Flags().StringVarP(&flagMessage, "message", "m", "", "message text to post (required)")
	_ = cmd.MarkFlagRequired("message")

	return cmd
}

// runPost はメッセージ投稿を実行する。
func runPost(cmd *cobra.Command, args []string) error {
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

	message := resolveAndTagMessage(cfg, flagMessage)
	result, err := client.PostMessage(ctx, channel, message)
	if err != nil {
		return err
	}

	if flagJSON {
		out := map[string]interface{}{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"ts":           result.TS,
			"message":      result.Message,
			"permalink":    result.Permalink,
		}
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(out)
	}

	// テキスト形式出力
	fmt.Fprintf(cmd.OutOrStdout(), "✅ Message posted to #%s\n", result.ChannelName)
	if result.Permalink != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "🔗 %s\n", result.Permalink)
	}

	return nil
}
