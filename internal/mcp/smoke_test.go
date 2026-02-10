package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/kai-ko/slack-fast-mcp/internal/config"
	slackclient "github.com/kai-ko/slack-fast-mcp/internal/slack"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestSmoke_ServerInitialization はMCP Serverが正しく初期化されることを検証する。
// ユーザーが実際に Cursor から接続した時に、ツール一覧が正しく返されるかの検証。
func TestSmoke_ServerInitialization(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{
		Token:          "xoxb-smoke-test",
		DefaultChannel: "test-channel",
	}

	s := NewServerWithClient(cfg, mock)
	if s == nil {
		t.Fatal("NewServerWithClient returned nil")
	}
}

// TestSmoke_AllToolsCallable は全3ツールがMCPプロトコル経由で呼び出し可能であることを検証する。
// これが通れば「Cursorからツールを呼び出したら動く」ことが保証される。
func TestSmoke_AllToolsCallable(t *testing.T) {
	postCalled := false
	historyCalled := false
	threadCalled := false

	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			postCalled = true
			return &slackclient.PostResult{
				Channel: "C01234", TS: "123.456", Message: message,
			}, nil
		},
		GetHistoryFunc: func(ctx context.Context, channel string, opts slackclient.HistoryOptions) (*slackclient.HistoryResult, error) {
			historyCalled = true
			return &slackclient.HistoryResult{
				Channel: "C01234", Count: 0, Messages: []slackclient.HistoryMessage{},
			}, nil
		},
		PostThreadFunc: func(ctx context.Context, channel, threadTS, message string) (*slackclient.PostResult, error) {
			threadCalled = true
			return &slackclient.PostResult{
				Channel: "C01234", TS: "123.789", ThreadTS: threadTS, Message: message,
			}, nil
		},
	}

	cfg := &config.Config{
		Token:          "xoxb-smoke-test",
		DefaultChannel: "test-channel",
	}

	// slack_post_message
	handler := postMessageHandler(mock, cfg)
	result, err := handler(context.Background(), newTestCallToolRequest("slack_post_message", map[string]any{
		"message": "smoke test message",
	}))
	if err != nil {
		t.Fatalf("slack_post_message: unexpected error: %v", err)
	}
	assertResultOK(t, "slack_post_message", result)
	if !postCalled {
		t.Error("slack_post_message: PostMessage was not called")
	}

	// slack_get_history
	historyHandler := getHistoryHandler(mock, cfg)
	result, err = historyHandler(context.Background(), newTestCallToolRequest("slack_get_history", map[string]any{}))
	if err != nil {
		t.Fatalf("slack_get_history: unexpected error: %v", err)
	}
	assertResultOK(t, "slack_get_history", result)
	if !historyCalled {
		t.Error("slack_get_history: GetHistory was not called")
	}

	// slack_post_thread
	threadHandler := postThreadHandler(mock, cfg)
	result, err = threadHandler(context.Background(), newTestCallToolRequest("slack_post_thread", map[string]any{
		"thread_ts": "123.456",
		"message":   "smoke test thread reply",
	}))
	if err != nil {
		t.Fatalf("slack_post_thread: unexpected error: %v", err)
	}
	assertResultOK(t, "slack_post_thread", result)
	if !threadCalled {
		t.Error("slack_post_thread: PostThread was not called")
	}
}

// TestSmoke_ErrorsAreUserFriendly はエラー時にLLM/ユーザーが理解できるメッセージが返されることを検証する。
func TestSmoke_ErrorsAreUserFriendly(t *testing.T) {
	mock := &slackclient.MockClient{}
	cfg := &config.Config{
		Token:          "xoxb-smoke-test",
		DefaultChannel: "", // デフォルトチャンネル未設定
	}

	tests := []struct {
		name      string
		tool      string
		args      map[string]any
		wantError string
		wantHint  bool
	}{
		{
			name:      "no channel, no default → helpful error",
			tool:      "slack_post_message",
			args:      map[string]any{"message": "test"},
			wantError: "no_default_channel",
			wantHint:  true,
		},
		{
			name:      "empty message → helpful error",
			tool:      "slack_post_message",
			args:      map[string]any{"message": ""},
			wantError: "no_text",
			wantHint:  true,
		},
		{
			name:      "no thread_ts → helpful error",
			tool:      "slack_post_thread",
			args:      map[string]any{"message": "reply", "thread_ts": ""},
			wantError: "thread_not_found",
			wantHint:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
			switch tt.tool {
			case "slack_post_message":
				handler = postMessageHandler(mock, cfg)
			case "slack_post_thread":
				handler = postThreadHandler(mock, cfg)
			default:
				t.Fatalf("unknown tool: %s", tt.tool)
			}

			result, err := handler(context.Background(), newTestCallToolRequest(tt.tool, tt.args))
			if err != nil {
				t.Fatalf("unexpected Go error: %v", err)
			}

			text := extractText(t, result)
			if !strings.Contains(text, tt.wantError) {
				t.Errorf("response = %q, want to contain error code %q", text, tt.wantError)
			}
			if tt.wantHint && !strings.Contains(text, "Hint:") {
				t.Errorf("response = %q, want to contain 'Hint:' for LLM guidance", text)
			}
		})
	}
}

// assertResultOK はツール呼び出し結果が正常（ok:true を含む）であることを検証する。
func assertResultOK(t *testing.T, toolName string, result *mcp.CallToolResult) {
	t.Helper()
	text := extractText(t, result)
	if !strings.Contains(text, `"ok":true`) {
		t.Errorf("%s: response does not contain ok:true. Got: %s", toolName, text)
	}
}
