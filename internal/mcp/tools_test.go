package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kai-kou/slack-fast-mcp/internal/config"
	apperr "github.com/kai-kou/slack-fast-mcp/internal/errors"
	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
	"github.com/mark3labs/mcp-go/mcp"
)

// newTestCallToolRequest はテスト用のCallToolRequestを作成する。
func newTestCallToolRequest(toolName string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}
}

// --- M01: slack_post_message 正常系 ---
func TestPostMessageHandler_Success(t *testing.T) {
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			return &slackclient.PostResult{
				Channel:     "C01234ABCDE",
				ChannelName: "general",
				TS:          "1234567890.123456",
				Message:     message,
				Permalink:   "https://test.slack.com/archives/C01234ABCDE/p1234567890123456",
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "Hello from test",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := extractText(t, result)
	if !strings.Contains(text, `"ok":true`) {
		t.Errorf("result = %q, want to contain ok:true", text)
	}
	if !strings.Contains(text, `"channel":"C01234ABCDE"`) {
		t.Errorf("result = %q, want to contain channel", text)
	}
}

// --- M02: slack_post_message チャンネル指定 ---
func TestPostMessageHandler_WithChannel(t *testing.T) {
	var capturedChannel string
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			capturedChannel = channel
			return &slackclient.PostResult{
				Channel: "C09876ZZZZZ",
				TS:      "1234567890.123456",
				Message: message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "default-ch"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"channel": "specific-channel",
		"message": "Hello",
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedChannel != "specific-channel" {
		t.Errorf("capturedChannel = %q, want %q", capturedChannel, "specific-channel")
	}
}

// --- M03: slack_post_message デフォルトチャンネル ---
func TestPostMessageHandler_DefaultChannel(t *testing.T) {
	var capturedChannel string
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			capturedChannel = channel
			return &slackclient.PostResult{
				Channel: "C01234ABCDE",
				TS:      "1234567890.123456",
				Message: message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "my-default"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "Hello",
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedChannel != "my-default" {
		t.Errorf("capturedChannel = %q, want %q", capturedChannel, "my-default")
	}
}

// --- M04: slack_post_message no_default_channel ---
func TestPostMessageHandler_NoDefaultChannel(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{DefaultChannel: ""}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "Hello",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "no_default_channel") {
		t.Errorf("result = %q, want to contain no_default_channel error", text)
	}
}

// --- M05: slack_post_message no_text ---
func TestPostMessageHandler_NoText(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{DefaultChannel: "general"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "no_text") {
		t.Errorf("result = %q, want to contain no_text error", text)
	}
}

// --- M06: slack_get_history 正常系 ---
func TestGetHistoryHandler_Success(t *testing.T) {
	mock := &slackclient.MockClient{
		GetHistoryFunc: func(ctx context.Context, channel string, opts slackclient.HistoryOptions) (*slackclient.HistoryResult, error) {
			return &slackclient.HistoryResult{
				Channel:     "C01234ABCDE",
				ChannelName: "general",
				Messages: []slackclient.HistoryMessage{
					{
						User:     "U01234",
						UserName: "testuser",
						Text:     "Hello World",
						TS:       "1234567890.111111",
					},
				},
				HasMore: false,
				Count:   1,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := getHistoryHandler(mock, cfg)

	req := newTestCallToolRequest("slack_get_history", map[string]any{
		"limit": float64(10),
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, `"ok":true`) {
		t.Errorf("result = %q, want to contain ok:true", text)
	}
	if !strings.Contains(text, `"count":1`) {
		t.Errorf("result = %q, want to contain count:1", text)
	}
}

// --- M07: slack_get_history limit 指定 ---
func TestGetHistoryHandler_WithLimit(t *testing.T) {
	var capturedLimit int
	mock := &slackclient.MockClient{
		GetHistoryFunc: func(ctx context.Context, channel string, opts slackclient.HistoryOptions) (*slackclient.HistoryResult, error) {
			capturedLimit = opts.Limit
			return &slackclient.HistoryResult{Count: 0}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := getHistoryHandler(mock, cfg)

	req := newTestCallToolRequest("slack_get_history", map[string]any{
		"limit": float64(25),
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedLimit != 25 {
		t.Errorf("capturedLimit = %d, want 25", capturedLimit)
	}
}

// --- M08: slack_post_thread 正常系 ---
func TestPostThreadHandler_Success(t *testing.T) {
	mock := &slackclient.MockClient{
		PostThreadFunc: func(ctx context.Context, channel, threadTS, message string) (*slackclient.PostResult, error) {
			return &slackclient.PostResult{
				Channel:  "C01234ABCDE",
				TS:       "1234567890.654321",
				ThreadTS: threadTS,
				Message:  message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := postThreadHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_thread", map[string]any{
		"thread_ts": "1234567890.123456",
		"message":   "Thread reply",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, `"ok":true`) {
		t.Errorf("result = %q, want to contain ok:true", text)
	}
	if !strings.Contains(text, `"thread_ts":"1234567890.123456"`) {
		t.Errorf("result = %q, want to contain thread_ts", text)
	}
}

// --- M09: slack_post_thread thread_not_found ---
func TestPostThreadHandler_NoThreadTS(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{DefaultChannel: "general"}
	handler := postThreadHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_thread", map[string]any{
		"thread_ts": "",
		"message":   "Reply",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "thread_not_found") {
		t.Errorf("result = %q, want to contain thread_not_found error", text)
	}
}

// --- M10: Slack API エラーのMCPエラー変換 ---
func TestPostMessageHandler_SlackAPIError(t *testing.T) {
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			return nil, apperr.New(apperr.CodeNotInChannel, "Botがチャンネルに参加していません", nil)
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "Hello",
	})

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	if !strings.Contains(text, "not_in_channel") {
		t.Errorf("result = %q, want to contain not_in_channel", text)
	}
	if !strings.Contains(text, "Hint:") {
		t.Errorf("result = %q, want to contain Hint:", text)
	}
}

// --- M12: ツール一覧確認 ---
func TestServer_ToolsRegistered(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{DefaultChannel: "general"}
	s := NewServerWithClient(cfg, mock)
	if s == nil {
		t.Fatal("server is nil")
	}
}

// --- M13: appendDisplayNameTag テスト ---
func TestAppendDisplayNameTag(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		displayName string
		want        string
	}{
		{
			name:        "empty display_name → no change",
			message:     "Hello World",
			displayName: "",
			want:        "Hello World",
		},
		{
			name:        "basic tag append",
			message:     "Hello World",
			displayName: "くろ",
			want:        "Hello World\n#くろ",
		},
		{
			name:        "message already has hashtag line → same line",
			message:     "Hello World\n#cursor #slack-fast-mcp",
			displayName: "くろ",
			want:        "Hello World\n#cursor #slack-fast-mcp #くろ",
		},
		{
			name:        "message has hashtag line with trailing space",
			message:     "Hello World\n#cursor #dev ",
			displayName: "しろ",
			want:        "Hello World\n#cursor #dev #しろ",
		},
		{
			name:        "multiline message without hashtag",
			message:     "Line 1\nLine 2\nLine 3",
			displayName: "くろ",
			want:        "Line 1\nLine 2\nLine 3\n#くろ",
		},
		{
			name:        "single line with hashtag prefix",
			message:     "#already-tagged",
			displayName: "くろ",
			want:        "#already-tagged #くろ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendDisplayNameTag(tt.message, tt.displayName)
			if got != tt.want {
				t.Errorf("appendDisplayNameTag(%q, %q) = %q, want %q", tt.message, tt.displayName, got, tt.want)
			}
		})
	}
}

// --- M14: slack_post_message with display_name ---
func TestPostMessageHandler_WithDisplayName(t *testing.T) {
	var capturedMessage string
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			capturedMessage = message
			return &slackclient.PostResult{
				Channel: "C01234ABCDE",
				TS:      "1234567890.123456",
				Message: message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general"}
	handler := postMessageHandler(mock, cfg)

	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message":      "Hello from test\n#cursor #project",
		"display_name": "くろ",
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedMessage, "#くろ") {
		t.Errorf("capturedMessage = %q, want to contain #くろ", capturedMessage)
	}
	// ハッシュタグ行に追記されるはず
	if !strings.Contains(capturedMessage, "#cursor #project #くろ") {
		t.Errorf("capturedMessage = %q, want hashtag appended to existing line", capturedMessage)
	}
}

// --- M15: slack_post_message with display_name from config ---
func TestPostMessageHandler_DisplayNameFromConfig(t *testing.T) {
	var capturedMessage string
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			capturedMessage = message
			return &slackclient.PostResult{
				Channel: "C01234ABCDE",
				TS:      "1234567890.123456",
				Message: message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general", DisplayName: "しろ"}
	handler := postMessageHandler(mock, cfg)

	// display_name パラメータ未指定 → Config のデフォルトが使われる
	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "Hello from config",
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedMessage, "#しろ") {
		t.Errorf("capturedMessage = %q, want to contain #しろ from config", capturedMessage)
	}
}

// --- M16: display_name パラメータが Config を上書き ---
func TestPostMessageHandler_DisplayNameParamOverridesConfig(t *testing.T) {
	var capturedMessage string
	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			capturedMessage = message
			return &slackclient.PostResult{
				Channel: "C01234ABCDE",
				TS:      "1234567890.123456",
				Message: message,
			}, nil
		},
	}

	cfg := &config.Config{DefaultChannel: "general", DisplayName: "しろ"}
	handler := postMessageHandler(mock, cfg)

	// display_name パラメータ指定 → Config よりパラメータ優先
	req := newTestCallToolRequest("slack_post_message", map[string]any{
		"message":      "Hello with override",
		"display_name": "くろ",
	})

	_, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(capturedMessage, "#くろ") {
		t.Errorf("capturedMessage = %q, want to contain #くろ (param override)", capturedMessage)
	}
	if strings.Contains(capturedMessage, "#しろ") {
		t.Errorf("capturedMessage = %q, should NOT contain #しろ (config default overridden)", capturedMessage)
	}
}

// --- ヘルパー ---

func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			return textContent.Text
		}
	}
	b, _ := json.Marshal(result)
	t.Fatalf("no text content found in result: %s", string(b))
	return ""
}
