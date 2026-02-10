package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	slackclient "github.com/kai-kou/slack-fast-mcp/internal/slack"
)

// TestVersionCmd は version サブコマンドのテスト。
func TestVersionCmd(t *testing.T) {
	Version = "1.0.0-test"

	t.Run("text output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"version"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("version command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1.0.0-test") {
			t.Errorf("expected version in output, got: %s", output)
		}
		if !strings.Contains(output, "Go:") {
			t.Errorf("expected Go version in output, got: %s", output)
		}
		if !strings.Contains(output, "Platform:") {
			t.Errorf("expected Platform in output, got: %s", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"version", "--json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("version --json command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"version":"1.0.0-test"`) {
			t.Errorf("expected JSON version in output, got: %s", output)
		}
		if !strings.Contains(output, `"go_version"`) {
			t.Errorf("expected go_version in JSON output, got: %s", output)
		}
	})
}

// TestHelpCmd は help 出力のテスト。
func TestHelpCmd(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()

	// 必要なサブコマンドが表示されているか
	expectedSubcommands := []string{"serve", "post", "history", "reply", "setup", "version"}
	for _, sub := range expectedSubcommands {
		if !strings.Contains(output, sub) {
			t.Errorf("expected subcommand '%s' in help output, got: %s", sub, output)
		}
	}

	// グローバルフラグが表示されているか
	expectedFlags := []string{"--config", "--token", "--channel", "--verbose", "--json"}
	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected flag '%s' in help output, got: %s", flag, output)
		}
	}
}

// TestPostCmdRequiresMessage は post コマンドの --message 必須チェック。
func TestPostCmdRequiresMessage(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"post"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when message is not provided")
	}

	if !strings.Contains(err.Error(), "message") {
		t.Errorf("expected error about missing message, got: %v", err)
	}
}

// TestReplyCmdRequiresFlags は reply コマンドの必須フラグチェック。
func TestReplyCmdRequiresFlags(t *testing.T) {
	t.Run("missing thread-ts", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"reply", "--message", "test"})

		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error when thread-ts is not provided")
		}
	})

	t.Run("missing message", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"reply", "--thread-ts", "123.456"})

		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error when message is not provided")
		}
	})
}

// TestPostCmdHelp は post サブコマンドのヘルプ出力テスト。
func TestPostCmdHelp(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"post", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("post help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--message") {
		t.Errorf("expected --message flag in post help, got: %s", output)
	}
	if !strings.Contains(output, "Examples:") {
		t.Errorf("expected examples in post help, got: %s", output)
	}
}

// TestHistoryCmdHelp は history サブコマンドのヘルプ出力テスト。
func TestHistoryCmdHelp(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"history", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("history help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "--limit") {
		t.Errorf("expected --limit flag in history help, got: %s", output)
	}
	if !strings.Contains(output, "--oldest") {
		t.Errorf("expected --oldest flag in history help, got: %s", output)
	}
	if !strings.Contains(output, "--latest") {
		t.Errorf("expected --latest flag in history help, got: %s", output)
	}
}

// TestSetupCmdHelp は setup サブコマンドのヘルプ出力テスト。
func TestSetupCmdHelp(t *testing.T) {
	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"setup", "--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("setup help failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "wizard") {
		t.Errorf("expected 'wizard' in setup help, got: %s", output)
	}
}

// TestNoTokenError はトークン未設定時のエラーテスト。
func TestNoTokenError(t *testing.T) {
	// 環境変数をクリアして post コマンドを実行
	t.Setenv("SLACK_BOT_TOKEN", "")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "")

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"post", "--message", "test", "--channel", "general"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when token is not configured")
	}

	// エラーメッセージにトークン関連の情報が含まれるか
	errStr := err.Error()
	if !strings.Contains(errStr, "token") {
		t.Errorf("expected token-related error, got: %s", errStr)
	}
}

// --- モッククライアントを使ったサブコマンドテスト ---

// setupMockClient はテスト用のモッククライアントを設定し、クリーンアップ関数を返す。
func setupMockClient(t *testing.T, mock *slackclient.MockClient) {
	t.Helper()
	oldFactory := clientFactory
	clientFactory = func(token string) slackclient.SlackClient {
		return mock
	}
	t.Cleanup(func() {
		clientFactory = oldFactory
		// グローバルフラグもリセット
		flagConfig = ""
		flagToken = ""
		flagChannel = ""
		flagVerbose = false
		flagJSON = false
		flagMessage = ""
		flagThreadTS = ""
		flagLimit = 10
		flagOldest = ""
		flagLatest = ""
	})
}

// TestPostCmdWithMock はモックを使った post コマンドテスト。
func TestPostCmdWithMock(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "")

	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			return &slackclient.PostResult{
				Channel:     "C12345",
				ChannelName: "general",
				TS:          "1234567890.123456",
				Message:     message,
				Permalink:   "https://slack.com/archives/C12345/p1234567890123456",
			}, nil
		},
	}
	setupMockClient(t, mock)

	t.Run("text output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"post", "--channel", "general", "--message", "Hello World"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("post command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Message posted") {
			t.Errorf("expected success message, got: %s", output)
		}
		if !strings.Contains(output, "#general") {
			t.Errorf("expected channel name in output, got: %s", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"post", "--channel", "general", "--message", "Hello", "--json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("post --json failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"ok"`) {
			t.Errorf("expected JSON ok field, got: %s", output)
		}
		if !strings.Contains(output, `"channel"`) {
			t.Errorf("expected JSON channel field, got: %s", output)
		}
		if !strings.Contains(output, `"permalink"`) {
			t.Errorf("expected JSON permalink field, got: %s", output)
		}
	})
}

// TestHistoryCmdWithMock はモックを使った history コマンドテスト。
func TestHistoryCmdWithMock(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "")

	mock := &slackclient.MockClient{
		GetHistoryFunc: func(ctx context.Context, channel string, opts slackclient.HistoryOptions) (*slackclient.HistoryResult, error) {
			return &slackclient.HistoryResult{
				Channel:     "C12345",
				ChannelName: "general",
				Messages: []slackclient.HistoryMessage{
					{
						User:     "U12345",
						UserName: "testuser",
						Text:     "Hello from test",
						TS:       "1234567890.111111",
					},
					{
						User:       "U67890",
						UserName:   "anotheruser",
						Text:       "Thread message",
						TS:         "1234567890.222222",
						ReplyCount: 3,
					},
				},
				HasMore: false,
				Count:   2,
			}, nil
		},
	}
	setupMockClient(t, mock)

	t.Run("text output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"history", "--channel", "general"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("history command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "#general") {
			t.Errorf("expected channel name, got: %s", output)
		}
		if !strings.Contains(output, "2 messages") {
			t.Errorf("expected message count, got: %s", output)
		}
		if !strings.Contains(output, "testuser") {
			t.Errorf("expected username, got: %s", output)
		}
		if !strings.Contains(output, "Hello from test") {
			t.Errorf("expected message text, got: %s", output)
		}
		if !strings.Contains(output, "3 replies") {
			t.Errorf("expected reply count, got: %s", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"history", "--channel", "general", "--json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("history --json failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"messages"`) {
			t.Errorf("expected messages in JSON, got: %s", output)
		}
		if !strings.Contains(output, `"count"`) {
			t.Errorf("expected count in JSON, got: %s", output)
		}
	})
}

// TestReplyCmdWithMock はモックを使った reply コマンドテスト。
func TestReplyCmdWithMock(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "")

	mock := &slackclient.MockClient{
		PostThreadFunc: func(ctx context.Context, channel, threadTS, message string) (*slackclient.PostResult, error) {
			return &slackclient.PostResult{
				Channel:     "C12345",
				ChannelName: "general",
				TS:          "1234567890.999999",
				ThreadTS:    threadTS,
				Message:     message,
				Permalink:   "https://slack.com/archives/C12345/p1234567890999999",
			}, nil
		},
	}
	setupMockClient(t, mock)

	t.Run("text output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"reply", "--channel", "general", "--thread-ts", "1234567890.123456", "--message", "Reply here"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("reply command failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Reply posted") {
			t.Errorf("expected success message, got: %s", output)
		}
		if !strings.Contains(output, "#general") {
			t.Errorf("expected channel name, got: %s", output)
		}
	})

	t.Run("json output", func(t *testing.T) {
		rootCmd := NewRootCmd()
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs([]string{"reply", "--channel", "general", "--thread-ts", "1234567890.123456", "--message", "Reply", "--json"})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("reply --json failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"thread_ts"`) {
			t.Errorf("expected thread_ts in JSON, got: %s", output)
		}
	})
}

// TestDefaultChannelFromConfig はデフォルトチャンネルの動作テスト。
func TestDefaultChannelFromConfig(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "default-channel")

	mock := &slackclient.MockClient{
		PostMessageFunc: func(ctx context.Context, channel, message string) (*slackclient.PostResult, error) {
			if channel != "default-channel" {
				t.Errorf("expected channel 'default-channel', got: %s", channel)
			}
			return &slackclient.PostResult{
				Channel:     "C12345",
				ChannelName: "default-channel",
				TS:          "1234567890.123456",
				Message:     message,
			}, nil
		},
	}
	setupMockClient(t, mock)

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	// --channel を指定しない → デフォルトチャンネルが使われる
	rootCmd.SetArgs([]string{"post", "--message", "test"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("post with default channel failed: %v", err)
	}
}

// TestIsInGitignore は .gitignore チェックのテスト。
func TestIsInGitignore(t *testing.T) {
	// テスト用の一時ディレクトリに移動
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	t.Run("no gitignore file", func(t *testing.T) {
		if isInGitignore(".slack-mcp.json") {
			t.Error("expected false when .gitignore doesn't exist")
		}
	})

	t.Run("pattern not in gitignore", func(t *testing.T) {
		os.WriteFile(".gitignore", []byte("node_modules\n.env\n"), 0644)
		if isInGitignore(".slack-mcp.json") {
			t.Error("expected false when pattern not in .gitignore")
		}
	})

	t.Run("pattern in gitignore", func(t *testing.T) {
		os.WriteFile(".gitignore", []byte("node_modules\n.slack-mcp.json\n.env\n"), 0644)
		if !isInGitignore(".slack-mcp.json") {
			t.Error("expected true when pattern in .gitignore")
		}
	})
}

// TestAppendToGitignore は .gitignore 追記のテスト。
func TestAppendToGitignore(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	t.Run("creates gitignore if not exists", func(t *testing.T) {
		os.Remove(".gitignore")
		if err := appendToGitignore(".slack-mcp.json"); err != nil {
			t.Fatalf("appendToGitignore failed: %v", err)
		}
		data, _ := os.ReadFile(".gitignore")
		if !strings.Contains(string(data), ".slack-mcp.json") {
			t.Errorf("expected pattern in .gitignore, got: %s", data)
		}
	})

	t.Run("appends to existing gitignore", func(t *testing.T) {
		os.WriteFile(".gitignore", []byte("node_modules\n"), 0644)
		if err := appendToGitignore(".slack-mcp.json"); err != nil {
			t.Fatalf("appendToGitignore failed: %v", err)
		}
		data, _ := os.ReadFile(".gitignore")
		content := string(data)
		if !strings.Contains(content, "node_modules") {
			t.Error("existing content should be preserved")
		}
		if !strings.Contains(content, ".slack-mcp.json") {
			t.Error("new pattern should be added")
		}
	})

	t.Run("adds newline before append if missing", func(t *testing.T) {
		os.WriteFile(".gitignore", []byte("node_modules"), 0644) // no trailing newline
		if err := appendToGitignore(".slack-mcp.json"); err != nil {
			t.Fatalf("appendToGitignore failed: %v", err)
		}
		data, _ := os.ReadFile(".gitignore")
		content := string(data)
		if strings.Contains(content, "node_modules.slack-mcp.json") {
			t.Error("should have newline between entries")
		}
	})
}

// TestNoChannelError はチャンネル未指定かつデフォルト未設定のエラーテスト。
func TestNoChannelError(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_DEFAULT_CHANNEL", "")

	mock := &slackclient.MockClient{}
	setupMockClient(t, mock)

	rootCmd := NewRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"post", "--message", "test"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when no channel is specified")
	}
	if !strings.Contains(err.Error(), "no_default_channel") {
		t.Errorf("expected no_default_channel error, got: %v", err)
	}
}
