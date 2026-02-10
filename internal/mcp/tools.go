package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kai-ko/slack-fast-mcp/internal/config"
	apperr "github.com/kai-ko/slack-fast-mcp/internal/errors"
	slackclient "github.com/kai-ko/slack-fast-mcp/internal/slack"
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
	)
}

func postMessageHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		message := request.GetString("message", "")

		if message == "" {
			appErr := apperr.New(apperr.CodeNoText, "メッセージが空です", nil)
			return mcp.NewToolResultError(appErr.FormatForMCP()), nil
		}

		channel, err := cfg.ResolveChannel(channelParam)
		if err != nil {
			return handleAppError(err)
		}

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
	)
}

func postThreadHandler(client slackclient.SlackClient, cfg *config.Config) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channelParam := request.GetString("channel", "")
		threadTS := request.GetString("thread_ts", "")
		message := request.GetString("message", "")

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

// --- ヘルパー ---

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
