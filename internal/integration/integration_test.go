//go:build integration

// Package integration は実Slack環境での統合テスト。
// SLACK_BOT_TOKEN と SLACK_TEST_CHANNEL 環境変数が必要。
//
// 実行方法:
//
//	SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test go test ./internal/integration/ -tags=integration -v -count=1
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kai-kou/slack-fast-mcp/internal/config"
	mcpserver "github.com/kai-kou/slack-fast-mcp/internal/mcp"
	"github.com/mark3labs/mcp-go/mcp"
)

// testEnv は統合テストの環境変数を保持する。
type testEnv struct {
	token   string
	channel string
}

// getTestEnv は統合テストに必要な環境変数を取得する。
// 未設定の場合はテストをスキップする。
func getTestEnv(t *testing.T) *testEnv {
	t.Helper()

	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		t.Skip("SLACK_BOT_TOKEN is not set; skipping integration test")
	}

	channel := os.Getenv("SLACK_TEST_CHANNEL")
	if channel == "" {
		t.Skip("SLACK_TEST_CHANNEL is not set; skipping integration test")
	}

	return &testEnv{
		token:   token,
		channel: channel,
	}
}

// newTestServer は実Slackトークンを使用したMCPサーバーを作成する。
func newTestServer(env *testEnv) *mcpTestServer {
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}
	s := mcpserver.NewServer(cfg)
	return &mcpTestServer{server: s, cfg: cfg}
}

// mcpTestServer はテスト用のMCPサーバーラッパー。
type mcpTestServer struct {
	server interface {
		// MCPServerはHandleMessage等のメソッドを持つが、
		// テストではツールハンドラーを直接呼び出す
	}
	cfg *config.Config
}

// TestIntegration_PostMessage は実Slackへのメッセージ投稿を検証する。
func TestIntegration_PostMessage(t *testing.T) {
	env := getTestEnv(t)
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}

	s := mcpserver.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[Go Integration Test] %s - slack_post_message テスト", timestamp)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "slack_post_message",
			Arguments: map[string]any{
				"channel": env.channel,
				"message": message,
			},
		},
	}

	// MCP Server経由でツールを呼び出す
	// Note: MCPServer.HandleMessage を使用して JSON-RPC レベルでテスト
	jsonRPCRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	}
	initBytes, _ := json.Marshal(jsonRPCRequest)
	initResp := s.HandleMessage(ctx, initBytes)
	if initResp == nil {
		t.Fatal("initialize response is nil")
	}

	// initialized 通知
	notifBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.HandleMessage(ctx, notifBytes)

	// tools/call
	toolCallRequest := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      request.Params.Name,
			"arguments": request.Params.Arguments,
		},
	}
	toolBytes, _ := json.Marshal(toolCallRequest)
	toolResp := s.HandleMessage(ctx, toolBytes)

	if toolResp == nil {
		t.Fatal("tools/call response is nil")
	}

	// レスポンスの解析（ツール結果テキストを抽出）
	respBytes, _ := json.Marshal(toolResp)
	toolText := extractToolText(t, respBytes)

	t.Logf("PostMessage tool result: %s", toolText)

	// ok:true の検証
	if !strings.Contains(toolText, `"ok":true`) {
		t.Errorf("PostMessage did not return ok:true. Response: %s", toolText)
	}

	// ts の存在確認
	if !strings.Contains(toolText, `"ts":`) {
		t.Errorf("PostMessage response does not contain ts. Response: %s", toolText)
	}

	// channel ID の存在確認
	if !strings.Contains(toolText, `"channel":`) {
		t.Errorf("PostMessage response does not contain channel. Response: %s", toolText)
	}
}

// TestIntegration_GetHistory は実Slackからの履歴取得を検証する。
func TestIntegration_GetHistory(t *testing.T) {
	env := getTestEnv(t)
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}

	s := mcpserver.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize
	initBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "integration-test", "version": "1.0.0"},
		},
	})
	s.HandleMessage(ctx, initBytes)

	notifBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.HandleMessage(ctx, notifBytes)

	// tools/call: slack_get_history
	toolBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "slack_get_history",
			"arguments": map[string]any{
				"channel": env.channel,
				"limit":   5,
			},
		},
	})
	toolResp := s.HandleMessage(ctx, toolBytes)

	if toolResp == nil {
		t.Fatal("tools/call response is nil")
	}

	respBytes, _ := json.Marshal(toolResp)
	toolText := extractToolText(t, respBytes)

	t.Logf("GetHistory tool result: %s", toolText)

	// ok:true の検証
	if !strings.Contains(toolText, `"ok":true`) {
		t.Errorf("GetHistory did not return ok:true. Response: %s", toolText)
	}

	// messages 配列の存在確認
	if !strings.Contains(toolText, `"messages":`) {
		t.Errorf("GetHistory response does not contain messages. Response: %s", toolText)
	}

	// count の存在確認
	if !strings.Contains(toolText, `"count":`) {
		t.Errorf("GetHistory response does not contain count. Response: %s", toolText)
	}
}

// TestIntegration_PostAndReplyThread はメッセージ投稿→スレッド返信のE2Eフローを検証する。
func TestIntegration_PostAndReplyThread(t *testing.T) {
	env := getTestEnv(t)
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}

	s := mcpserver.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Initialize
	initBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "integration-test", "version": "1.0.0"},
		},
	})
	s.HandleMessage(ctx, initBytes)

	notifBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.HandleMessage(ctx, notifBytes)

	// Step 1: メッセージ投稿
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	parentMessage := fmt.Sprintf("[Go Integration Test] %s - スレッドテスト親メッセージ", timestamp)

	postBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "slack_post_message",
			"arguments": map[string]any{
				"channel": env.channel,
				"message": parentMessage,
			},
		},
	})
	postResp := s.HandleMessage(ctx, postBytes)
	if postResp == nil {
		t.Fatal("PostMessage response is nil")
	}

	postRespBytes, _ := json.Marshal(postResp)
	postToolText := extractToolText(t, postRespBytes)
	t.Logf("PostMessage (parent) tool result: %s", postToolText)

	// ts を抽出
	parentTS := extractTSFromResponse(t, postToolText)
	if parentTS == "" {
		t.Fatal("Could not extract ts from PostMessage response")
	}
	t.Logf("Parent message ts: %s", parentTS)

	// 少し待機
	time.Sleep(2 * time.Second)

	// Step 2: スレッド返信
	replyMessage := fmt.Sprintf("[Go Integration Test] %s - スレッド返信テスト", timestamp)

	threadBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "slack_post_thread",
			"arguments": map[string]any{
				"channel":   env.channel,
				"thread_ts": parentTS,
				"message":   replyMessage,
			},
		},
	})
	threadResp := s.HandleMessage(ctx, threadBytes)
	if threadResp == nil {
		t.Fatal("PostThread response is nil")
	}

	threadRespBytes, _ := json.Marshal(threadResp)
	threadToolText := extractToolText(t, threadRespBytes)
	t.Logf("PostThread tool result: %s", threadToolText)

	// ok:true の検証
	if !strings.Contains(threadToolText, `"ok":true`) {
		t.Errorf("PostThread did not return ok:true. Response: %s", threadToolText)
	}

	// thread_ts の存在確認
	if !strings.Contains(threadToolText, `"thread_ts":`) {
		t.Errorf("PostThread response does not contain thread_ts. Response: %s", threadToolText)
	}
}

// TestIntegration_ErrorHandling は実環境でのエラーハンドリングを検証する。
func TestIntegration_ErrorHandling(t *testing.T) {
	env := getTestEnv(t)
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}

	s := mcpserver.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize
	initBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "integration-test", "version": "1.0.0"},
		},
	})
	s.HandleMessage(ctx, initBytes)

	notifBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.HandleMessage(ctx, notifBytes)

	tests := []struct {
		name          string
		toolName      string
		args          map[string]any
		wantError     string
		wantHint      bool
	}{
		{
			name:      "nonexistent channel → error",
			toolName:  "slack_post_message",
			args:      map[string]any{"channel": "nonexistent-channel-zzzzz-xxxxx", "message": "test"},
			wantError: "channel_not_found",
			wantHint:  true,
		},
		{
			name:      "empty message → no_text error",
			toolName:  "slack_post_message",
			args:      map[string]any{"channel": env.channel, "message": ""},
			wantError: "no_text",
			wantHint:  true,
		},
		{
			name:      "empty thread_ts → thread_not_found error",
			toolName:  "slack_post_thread",
			args:      map[string]any{"channel": env.channel, "thread_ts": "", "message": "reply"},
			wantError: "thread_not_found",
			wantHint:  true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolBytes, _ := json.Marshal(map[string]any{
				"jsonrpc": "2.0",
				"id":      10 + i,
				"method":  "tools/call",
				"params": map[string]any{
					"name":      tt.toolName,
					"arguments": tt.args,
				},
			})
			resp := s.HandleMessage(ctx, toolBytes)
			if resp == nil {
				t.Fatal("response is nil")
			}

			respBytes, _ := json.Marshal(resp)
			toolText := extractToolText(t, respBytes)
			t.Logf("Error test tool result: %s", toolText)

			if !strings.Contains(toolText, tt.wantError) {
				t.Errorf("Response does not contain expected error %q. Got: %s", tt.wantError, toolText)
			}

			if tt.wantHint && !strings.Contains(toolText, "Hint") {
				t.Errorf("Response does not contain 'Hint'. Got: %s", toolText)
			}
		})
	}
}

// TestIntegration_DefaultChannel はデフォルトチャンネルからの操作を検証する。
func TestIntegration_DefaultChannel(t *testing.T) {
	env := getTestEnv(t)
	cfg := &config.Config{
		Token:          env.token,
		DefaultChannel: env.channel,
	}

	s := mcpserver.NewServer(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize
	initBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "integration-test", "version": "1.0.0"},
		},
	})
	s.HandleMessage(ctx, initBytes)

	notifBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	})
	s.HandleMessage(ctx, notifBytes)

	// channel パラメータなしで履歴取得 → デフォルトチャンネルが使われるはず
	toolBytes, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "slack_get_history",
			"arguments": map[string]any{
				"limit": 3,
			},
		},
	})
	resp := s.HandleMessage(ctx, toolBytes)
	if resp == nil {
		t.Fatal("response is nil")
	}

	respBytes, _ := json.Marshal(resp)
	toolText := extractToolText(t, respBytes)
	t.Logf("DefaultChannel GetHistory tool result: %s", toolText)

	if !strings.Contains(toolText, `"ok":true`) {
		t.Errorf("GetHistory with default channel did not return ok:true. Response: %s", toolText)
	}
}

// extractToolText はJSON-RPCレスポンスからMCPツール結果のテキスト部分を抽出する。
// JSON-RPC → result → content[0] → text の順でネストされたJSONを解放する。
func extractToolText(t *testing.T, respBytes []byte) string {
	t.Helper()

	// JSON-RPC レスポンスをパース
	var rpcResp map[string]any
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		return string(respBytes) // パースできない場合は生のレスポンスを返す
	}

	// result → content を取得
	result, ok := rpcResp["result"].(map[string]any)
	if !ok {
		return string(respBytes)
	}

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		return string(respBytes)
	}

	// content[0] → text を取得
	firstContent, ok := content[0].(map[string]any)
	if !ok {
		return string(respBytes)
	}

	text, ok := firstContent["text"].(string)
	if !ok {
		return string(respBytes)
	}

	return text
}

// extractTSFromResponse はJSON-RPCレスポンス文字列からtsを抽出する。
func extractTSFromResponse(t *testing.T, respStr string) string {
	t.Helper()

	// "ts":"..." パターンを探す（thread_tsではなくtsを取得）
	// JSON内にネストされているので、簡易パースで取得
	idx := strings.Index(respStr, `"ts":"`)
	if idx == -1 {
		return ""
	}

	// "ts":"の後の値を取得
	start := idx + len(`"ts":"`)
	end := strings.Index(respStr[start:], `"`)
	if end == -1 {
		return ""
	}

	ts := respStr[start : start+end]

	// ts が有効な形式か確認 (数字.数字)
	if !strings.Contains(ts, ".") {
		return ""
	}

	return ts
}
