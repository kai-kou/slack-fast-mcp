package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	slackapi "github.com/slack-go/slack"
)

// newMockSlackServer はSlack APIのモックHTTPサーバーを作成する。
func newMockSlackServer(t *testing.T, handlers map[string]http.HandlerFunc) (*httptest.Server, *slackapi.Client) {
	t.Helper()
	mux := http.NewServeMux()
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := slackapi.New("xoxb-test-token", slackapi.OptionAPIURL(server.URL+"/"))
	return server, client
}

// jsonResponse はJSON応答を書き込むヘルパー。
func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// --- S01: PostMessage 正常系 ---
func TestClient_PostMessage_Success(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":      true,
				"channel": "C01234ABCDE",
				"ts":      "1234567890.123456",
			})
		},
		"/chat.getPermalink": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":        true,
				"permalink": "https://test.slack.com/archives/C01234ABCDE/p1234567890123456",
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.PostMessage(context.Background(), "C01234ABCDE", "Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Channel != "C01234ABCDE" {
		t.Errorf("Channel = %q, want %q", result.Channel, "C01234ABCDE")
	}
	if result.TS != "1234567890.123456" {
		t.Errorf("TS = %q, want %q", result.TS, "1234567890.123456")
	}
	if result.Message != "Hello World" {
		t.Errorf("Message = %q, want %q", result.Message, "Hello World")
	}
}

// --- S03: PostThread 正常系 ---
func TestClient_PostThread_Success(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":      true,
				"channel": "C01234ABCDE",
				"ts":      "1234567890.654321",
			})
		},
		"/chat.getPermalink": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":        true,
				"permalink": "https://test.slack.com/archives/C01234ABCDE/p1234567890654321",
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.PostThread(context.Background(), "C01234ABCDE", "1234567890.123456", "Thread reply")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ThreadTS != "1234567890.123456" {
		t.Errorf("ThreadTS = %q, want %q", result.ThreadTS, "1234567890.123456")
	}
	if result.Message != "Thread reply" {
		t.Errorf("Message = %q, want %q", result.Message, "Thread reply")
	}
}

// --- S04: GetHistory 正常系 ---
func TestClient_GetHistory_Success(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.history": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
				"messages": []map[string]any{
					{
						"user":    "U01234ABCDE",
						"text":    "Hello from history",
						"ts":      "1234567890.111111",
						"type":    "message",
						"subtype": "",
					},
				},
				"has_more": false,
			})
		},
		"/users.info": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
				"user": map[string]any{
					"id":   "U01234ABCDE",
					"name": "testuser",
				},
			})
		},
		"/chat.getPermalink": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":        true,
				"permalink": "https://test.slack.com/archives/C01234/p1234567890111111",
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.GetHistory(context.Background(), "C01234ABCDE", HistoryOptions{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 {
		t.Errorf("Count = %d, want 1", result.Count)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("Messages length = %d, want 1", len(result.Messages))
	}
	if result.Messages[0].Text != "Hello from history" {
		t.Errorf("Messages[0].Text = %q, want %q", result.Messages[0].Text, "Hello from history")
	}
	if result.Messages[0].UserName != "testuser" {
		t.Errorf("Messages[0].UserName = %q, want %q", result.Messages[0].UserName, "testuser")
	}
}

// --- S05: GetHistory limit バリデーション ---
func TestClient_GetHistory_LimitValidation(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.history": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":       true,
				"messages": []map[string]any{},
				"has_more": false,
			})
		},
	})

	client := NewClientWithAPI(api)

	// limit=0 → デフォルト10
	_, err := client.GetHistory(context.Background(), "C01234ABCDE", HistoryOptions{Limit: 0})
	if err != nil {
		t.Fatalf("unexpected error with limit=0: %v", err)
	}

	// limit=200 → 上限100
	_, err = client.GetHistory(context.Background(), "C01234ABCDE", HistoryOptions{Limit: 200})
	if err != nil {
		t.Fatalf("unexpected error with limit=200: %v", err)
	}
}

// --- S10: エラー channel_not_found ---
func TestClient_PostMessage_ChannelNotFound(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":    false,
				"error": "channel_not_found",
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.PostMessage(context.Background(), "C01234ABCDE", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "channel_not_found") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "channel_not_found")
	}
}

// --- S11: エラー not_in_channel ---
func TestClient_PostMessage_NotInChannel(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":    false,
				"error": "not_in_channel",
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.PostMessage(context.Background(), "C01234ABCDE", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not_in_channel") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "not_in_channel")
	}
}

// --- S12: エラー invalid_auth ---
func TestClient_PostMessage_InvalidAuth(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":    false,
				"error": "invalid_auth",
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.PostMessage(context.Background(), "C01234ABCDE", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_auth") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "invalid_auth")
	}
}

// --- S15: context キャンセル ---
func TestClient_PostMessage_ContextCancelled(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			// レスポンスを返さない（タイムアウトさせる）
			select {}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	client := NewClientWithAPI(api)
	_, err := client.PostMessage(ctx, "C01234ABCDE", "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- S02: PostMessage チャンネル名解決 ---
func TestClient_PostMessage_WithChannelName(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.list": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
				"channels": []map[string]any{
					{"id": "C09876ZZZZZ", "name": "general"},
				},
				"response_metadata": map[string]any{
					"next_cursor": "",
				},
			})
		},
		"/chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":      true,
				"channel": "C09876ZZZZZ",
				"ts":      "1234567890.123456",
			})
		},
		"/chat.getPermalink": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":        true,
				"permalink": "https://test.slack.com/archives/C09876ZZZZZ/p1234567890123456",
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.PostMessage(context.Background(), "general", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Channel != "C09876ZZZZZ" {
		t.Errorf("Channel = %q, want %q", result.Channel, "C09876ZZZZZ")
	}
	if result.ChannelName != "general" {
		t.Errorf("ChannelName = %q, want %q", result.ChannelName, "general")
	}
}

// --- S09: ResolveChannel 存在しないチャンネル ---
func TestClient_ResolveChannel_NotFound(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.list": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":       true,
				"channels": []map[string]any{},
				"response_metadata": map[string]any{
					"next_cursor": "",
				},
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.ResolveChannel(context.Background(), "nonexistent-channel")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "channel_not_found") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "channel_not_found")
	}
}

// --- S06/S07: ResolveChannel チャンネルID / ハッシュ付きチャンネル名 ---
func TestClient_ResolveChannel_Patterns(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.list": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
				"channels": []map[string]any{
					{"id": "C01234ABCDE", "name": "general"},
				},
				"response_metadata": map[string]any{
					"next_cursor": "",
				},
			})
		},
	})

	client := NewClientWithAPI(api)

	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{"channel ID", "C01234ABCDE", "C01234ABCDE", false},
		{"channel name", "general", "C01234ABCDE", false},
		{"hash prefix", "#general", "C01234ABCDE", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := client.ResolveChannel(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if id != tt.wantID {
				t.Errorf("ID = %q, want %q", id, tt.wantID)
			}
		})
	}
}

// --- S20: AddReaction 正常系 ---
func TestClient_AddReaction_Success(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/reactions.add": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.AddReaction(context.Background(), "C01234ABCDE", "1234567890.123456", "thumbsup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Channel != "C01234ABCDE" {
		t.Errorf("Channel = %q, want %q", result.Channel, "C01234ABCDE")
	}
	if result.Timestamp != "1234567890.123456" {
		t.Errorf("Timestamp = %q, want %q", result.Timestamp, "1234567890.123456")
	}
	if result.Reaction != "thumbsup" {
		t.Errorf("Reaction = %q, want %q", result.Reaction, "thumbsup")
	}
}

// --- S21: AddReaction already_reacted エラー ---
func TestClient_AddReaction_AlreadyReacted(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/reactions.add": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":    false,
				"error": "already_reacted",
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.AddReaction(context.Background(), "C01234ABCDE", "1234567890.123456", "thumbsup")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "already_reacted") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "already_reacted")
	}
}

// --- S22: RemoveReaction 正常系 ---
func TestClient_RemoveReaction_Success(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/reactions.remove": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.RemoveReaction(context.Background(), "C01234ABCDE", "1234567890.123456", "thumbsup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Channel != "C01234ABCDE" {
		t.Errorf("Channel = %q, want %q", result.Channel, "C01234ABCDE")
	}
	if result.Reaction != "thumbsup" {
		t.Errorf("Reaction = %q, want %q", result.Reaction, "thumbsup")
	}
}

// --- S23: RemoveReaction no_reaction エラー ---
func TestClient_RemoveReaction_NoReaction(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/reactions.remove": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok":    false,
				"error": "no_reaction",
			})
		},
	})

	client := NewClientWithAPI(api)
	_, err := client.RemoveReaction(context.Background(), "C01234ABCDE", "1234567890.123456", "thumbsup")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no_reaction") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "no_reaction")
	}
}

// --- S24: AddReaction チャンネル名解決 ---
func TestClient_AddReaction_WithChannelName(t *testing.T) {
	_, api := newMockSlackServer(t, map[string]http.HandlerFunc{
		"/conversations.list": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
				"channels": []map[string]any{
					{"id": "C09876ZZZZZ", "name": "general"},
				},
				"response_metadata": map[string]any{
					"next_cursor": "",
				},
			})
		},
		"/reactions.add": func(w http.ResponseWriter, r *http.Request) {
			jsonResponse(w, map[string]any{
				"ok": true,
			})
		},
	})

	client := NewClientWithAPI(api)
	result, err := client.AddReaction(context.Background(), "general", "1234567890.123456", "heart")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Channel != "C09876ZZZZZ" {
		t.Errorf("Channel = %q, want %q", result.Channel, "C09876ZZZZZ")
	}
	if result.ChannelName != "general" {
		t.Errorf("ChannelName = %q, want %q", result.ChannelName, "general")
	}
}

// --- classifySlackError テスト ---
func TestClassifySlackError(t *testing.T) {
	tests := []struct {
		errMsg   string
		wantCode string
	}{
		{"channel_not_found", "channel_not_found"},
		{"not_in_channel", "not_in_channel"},
		{"invalid_auth", "invalid_auth"},
		{"not_authed", "invalid_auth"},
		{"missing_scope", "missing_scope"},
		{"thread_not_found", "thread_not_found"},
		{"no_text", "no_text"},
		{"already_reacted", "already_reacted"},
		{"no_reaction", "no_reaction"},
		{"invalid_name", "invalid_reaction"},
		{"unknown_error", "network_error"},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			err := classifySlackError(fmt.Errorf("%s", tt.errMsg))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantCode) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantCode)
			}
		})
	}
}

func TestClassifySlackError_Nil(t *testing.T) {
	if classifySlackError(nil) != nil {
		t.Error("expected nil for nil input")
	}
}
