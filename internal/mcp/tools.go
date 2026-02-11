package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kai-kou/slack-fast-mcp/internal/config"
	apperr "github.com/kai-kou/slack-fast-mcp/internal/errors"
	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// --- slack_post_message ---

func postMessageTool() mcp.Tool {
	return mcp.NewTool("slack_post_message",
		mcp.WithDescription("Post a message to a Slack channel. "+
			"Supports Slack mrkdwn formatting (bold, italic, links, code blocks). "+
			"If channel is omitted, posts to the configured default channel. "+
			"The bot must be invited to the target channel first."),
		mcp.WithString("channel",
			mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
				"If omitted, uses the configured default channel."),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("Message text to post. Supports Slack mrkdwn: "+
				"*bold*, _italic_, `code`, ```code block```, <url|text>."),
		),
		mcp.WithString("display_name",
			mcp.Description("Display name of the sender (e.g. AI agent persona name). "+
				"If provided, appends #display_name hashtag to the message. "+
				"Useful for identifying which AI persona posted the message."),
		),
	)
}

func postMessageHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		message := request.GetString("message", "")
		displayNameParam := request.GetString("display_name", "")

		if message == "" {
			appErr := apperr.New(apperr.CodeNoText, "メッセージが空です", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

		// display_name の解決（パラメータ > Config デフォルト）
		displayName := cfg.ResolveDisplayName(displayNameParam)
		message = appendDisplayNameTag(message, displayName)

		result, err := client.PostMessage(ctx, channel, message)
		if err != nil {
			return handleAppError(err)
		}

		return toolResultJSON(map[string]any{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"ts":           result.TS,
			"message":      result.Message,
			"permalink":    result.Permalink,
		})
	}
}

// --- slack_get_history ---

func getHistoryTool() mcp.Tool {
	return mcp.NewTool("slack_get_history",
		mcp.WithDescription("Get message history from a Slack channel. "+
			"Returns recent messages with user names, timestamps, and permalinks. "+
			"If channel is omitted, uses the configured default channel. "+
			"The bot must be invited to the target channel first."),
		mcp.WithString("channel",
			mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
				"If omitted, uses the configured default channel."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Number of messages to retrieve (1-100). Defaults to 10."),
		),
		mcp.WithString("oldest",
			mcp.Description("Start of time range (Unix timestamp). Only messages after this time are included."),
		),
		mcp.WithString("latest",
			mcp.Description("End of time range (Unix timestamp). Only messages before this time are included."),
		),
	)
}

func getHistoryHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

		limit := request.GetInt("limit", 10)
		oldest := request.GetString("oldest", "")
		latest := request.GetString("latest", "")

		opts := slackclient.HistoryOptions{
			Limit:  limit,
			Oldest: oldest,
			Latest: latest,
		}

		result, err := client.GetHistory(ctx, channel, opts)
		if err != nil {
			return handleAppError(err)
		}

		return toolResultJSON(map[string]any{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"messages":     result.Messages,
			"has_more":     result.HasMore,
			"count":        result.Count,
		})
	}
}

// --- slack_post_thread ---

func postThreadTool() mcp.Tool {
	return mcp.NewTool("slack_post_thread",
		mcp.WithDescription("Post a reply to an existing message thread in a Slack channel. "+
			"Supports Slack mrkdwn formatting. "+
			"If channel is omitted, uses the configured default channel. "+
			"The bot must be invited to the target channel first."),
		mcp.WithString("channel",
			mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
				"If omitted, uses the configured default channel."),
		),
		mcp.WithString("thread_ts",
			mcp.Required(),
			mcp.Description("Timestamp of the parent message to reply to (e.g. '1234567890.123456'). "+
				"Get this from the 'ts' field of slack_get_history or slack_post_message results."),
		),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("Reply message text to post. Supports Slack mrkdwn: "+
				"*bold*, _italic_, `code`, ```code block```, <url|text>."),
		),
		mcp.WithString("display_name",
			mcp.Description("Display name of the sender (e.g. AI agent persona name). "+
				"If provided, appends #display_name hashtag to the message. "+
				"Useful for identifying which AI persona posted the message."),
		),
	)
}

func postThreadHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		threadTS := request.GetString("thread_ts", "")
		message := request.GetString("message", "")
		displayNameParam := request.GetString("display_name", "")

		if message == "" {
			appErr := apperr.New(apperr.CodeNoText, "メッセージが空です", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		if threadTS == "" {
			appErr := apperr.New(apperr.CodeThreadNotFound, "thread_ts が指定されていません", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

		// display_name の解決（パラメータ > Config デフォルト）
		displayName := cfg.ResolveDisplayName(displayNameParam)
		message = appendDisplayNameTag(message, displayName)

		result, err := client.PostThread(ctx, channel, threadTS, message)
		if err != nil {
			return handleAppError(err)
		}

		return toolResultJSON(map[string]any{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"ts":           result.TS,
			"thread_ts":    result.ThreadTS,
			"message":      result.Message,
			"permalink":    result.Permalink,
		})
	}
}

// --- slack_add_reaction ---

func addReactionTool() mcp.Tool {
	return mcp.NewTool("slack_add_reaction",
		mcp.WithDescription("Add a reaction (emoji) to a message in a Slack channel. "+
			"Use emoji names without colons (e.g. 'thumbsup', not ':thumbsup:'). "+
			"If channel is omitted, uses the configured default channel. "+
			"The bot must be invited to the target channel first. "+
			"Requires the 'reactions:write' OAuth scope."),
		mcp.WithString("channel",
			mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
				"If omitted, uses the configured default channel."),
		),
		mcp.WithString("timestamp",
			mcp.Required(),
			mcp.Description("Timestamp of the message to react to (e.g. '1234567890.123456'). "+
				"Get this from the 'ts' field of slack_get_history or slack_post_message results."),
		),
		mcp.WithString("reaction",
			mcp.Required(),
			mcp.Description("Emoji name to react with, without colons (e.g. 'thumbsup', 'heart', 'eyes', '+1'). "+
				"See https://www.webfx.com/tools/emoji-cheat-sheet/ for available emoji names."),
		),
	)
}

func addReactionHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		timestamp := request.GetString("timestamp", "")
		reaction := request.GetString("reaction", "")

		if timestamp == "" {
			appErr := apperr.New(apperr.CodeThreadNotFound, "timestamp が指定されていません", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		if reaction == "" {
			appErr := apperr.New(apperr.CodeInvalidReaction, "reaction（絵文字名）が指定されていません", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		// コロン付きの絵文字名を正規化（:thumbsup: → thumbsup）
		reaction = normalizeEmojiName(reaction)

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

		result, err := client.AddReaction(ctx, channel, timestamp, reaction)
		if err != nil {
			return handleAppError(err)
		}

		return toolResultJSON(map[string]any{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"timestamp":    result.Timestamp,
			"reaction":     result.Reaction,
		})
	}
}

// --- slack_remove_reaction ---

func removeReactionTool() mcp.Tool {
	return mcp.NewTool("slack_remove_reaction",
		mcp.WithDescription("Remove a reaction (emoji) from a message in a Slack channel. "+
			"Use emoji names without colons (e.g. 'thumbsup', not ':thumbsup:'). "+
			"Can only remove reactions that were added by the bot. "+
			"If channel is omitted, uses the configured default channel. "+
			"The bot must be invited to the target channel first. "+
			"Requires the 'reactions:write' OAuth scope."),
		mcp.WithString("channel",
			mcp.Description("Channel name (e.g. 'general') or channel ID (e.g. 'C01234ABCDE'). "+
				"If omitted, uses the configured default channel."),
		),
		mcp.WithString("timestamp",
			mcp.Required(),
			mcp.Description("Timestamp of the message to remove reaction from (e.g. '1234567890.123456'). "+
				"Get this from the 'ts' field of slack_get_history or slack_post_message results."),
		),
		mcp.WithString("reaction",
			mcp.Required(),
			mcp.Description("Emoji name to remove, without colons (e.g. 'thumbsup', 'heart', 'eyes', '+1')."),
		),
	)
}

func removeReactionHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		timestamp := request.GetString("timestamp", "")
		reaction := request.GetString("reaction", "")

		if timestamp == "" {
			appErr := apperr.New(apperr.CodeThreadNotFound, "timestamp が指定されていません", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		if reaction == "" {
			appErr := apperr.New(apperr.CodeInvalidReaction, "reaction（絵文字名）が指定されていません", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		// コロン付きの絵文字名を正規化（:thumbsup: → thumbsup）
		reaction = normalizeEmojiName(reaction)

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

		result, err := client.RemoveReaction(ctx, channel, timestamp, reaction)
		if err != nil {
			return handleAppError(err)
		}

		return toolResultJSON(map[string]any{
			"ok":           true,
			"channel":      result.Channel,
			"channel_name": result.ChannelName,
			"timestamp":    result.Timestamp,
			"reaction":     result.Reaction,
		})
	}
}

// --- ヘルパー ---

// normalizeEmojiName はコロン付きの絵文字名からコロンを除去する。
// 例: ":thumbsup:" → "thumbsup", "thumbsup" → "thumbsup"
func normalizeEmojiName(name string) string {
	return strings.Trim(name, ":")
}

// appendDisplayNameTag は display_name が指定されている場合、メッセージ末尾にハッシュタグを追加する。
// 既にメッセージ末尾にハッシュタグ行がある場合は同じ行に追加し、ない場合は改行して追加する。
func appendDisplayNameTag(message, displayName string) string {
	if displayName == "" {
		return message
	}

	tag := "#" + displayName
	// メッセージ末尾がハッシュタグ行で終わっている場合は同じ行に追加
	lines := strings.Split(message, "\n")
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	if strings.HasPrefix(lastLine, "#") {
		lines[len(lines)-1] = strings.TrimRight(lines[len(lines)-1], " ") + " " + tag
		return strings.Join(lines, "\n")
	}

	return message + "\n" + tag
}

// handleAppError はエラーをMCPツールエラーに変換する。
func handleAppError(err error) (*mcp.CallToolResult, error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		return mcp.NewToolResultError(appErr.FormatForMCP()), nil
	}
	return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
}

// toolResultJSON はJSONテキストのToolResultを生成する。
// UTF-8そのまま出力（日本語をエスケープしない）。
func toolResultJSON(data any) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error: failed to marshal response: %v", err)), nil
	}
	return mcp.NewToolResultText(string(b)), nil
}
