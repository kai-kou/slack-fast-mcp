package errors

import (
	"fmt"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "with underlying error",
			err:  New(CodeChannelNotFound, "チャンネルが見つかりません", fmt.Errorf("api error")),
			want: "channel_not_found: チャンネルが見つかりません (api error)",
		},
		{
			name: "without underlying error",
			err:  New(CodeNoText, "メッセージが空です", nil),
			want: "no_text: メッセージが空です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")
	err := New(CodeNetworkError, "接続失敗", inner)

	if err.Unwrap() != inner {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), inner)
	}
}

func TestAppError_FormatForMCP(t *testing.T) {
	err := New(CodeChannelNotFound, "チャンネルが見つかりません", nil)
	got := err.FormatForMCP()
	want := "Error [channel_not_found]: チャンネルが見つかりません\nHint: The channel was not found. Ask the user to verify the channel name or ID. Do not include the '#' prefix."
	if got != want {
		t.Errorf("FormatForMCP() = %q, want %q", got, want)
	}
}

func TestNew_HintFromMap(t *testing.T) {
	tests := []struct {
		code     string
		wantHint string
	}{
		{CodeChannelNotFound, "The channel was not found. Ask the user to verify the channel name or ID. Do not include the '#' prefix."},
		{CodeNotInChannel, "The bot is not a member of this channel. Ask the user to invite the bot by running: /invite @slack-fast-mcp"},
		{CodeInvalidAuth, "The Slack token is invalid or expired. Ask the user to regenerate the token at https://api.slack.com/apps"},
		{CodeNoText, "The message parameter is required and cannot be empty."},
		{CodeNoDefaultChannel, "No channel specified and no default_channel configured. Set default_channel in config or specify the channel parameter."},
		{CodeTokenNotConfigured, "No Slack token found. Ask the user to run 'slack-fast-mcp setup' or set the SLACK_BOT_TOKEN environment variable."},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := New(tt.code, "test", nil)
			if err.Hint != tt.wantHint {
				t.Errorf("Hint = %q, want %q", err.Hint, tt.wantHint)
			}
		})
	}
}

func TestNewWithHint(t *testing.T) {
	err := NewWithHint("custom_error", "カスタムエラー", "Custom hint for LLM", nil)
	if err.Code != "custom_error" {
		t.Errorf("Code = %q, want %q", err.Code, "custom_error")
	}
	if err.Hint != "Custom hint for LLM" {
		t.Errorf("Hint = %q, want %q", err.Hint, "Custom hint for LLM")
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"bot token", "xoxb-1234567890-abcdef", "xoxb-****"},
		{"user token", "xoxp-9876543210-ghijkl", "xoxp-****"},
		{"session token", "xoxs-111222333-mnopqr", "xoxs-****"},
		{"not a token", "some-random-string", "some-random-string"},
		{"short string", "abc", "abc"},
		{"empty string", "", ""},
		{"env var reference", "${SLACK_BOT_TOKEN}", "${SLACK_BOT_TOKEN}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskToken(tt.input)
			if got != tt.want {
				t.Errorf("MaskToken(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
