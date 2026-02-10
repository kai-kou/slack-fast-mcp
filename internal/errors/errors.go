// Package errors defines application-level error types for slack-fast-mcp.
package errors

import "fmt"

// AppError はアプリケーションエラーを表す構造体。
// Code: エラーコード（channel_not_found 等）
// Message: 人間向けメッセージ
// Hint: LLM向けの解決ヒント（英語）
// Err: 元のエラー
type AppError struct {
	Code    string
	Message string
	Hint    string
	Err     error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// エラーコード定数
const (
	CodeChannelNotFound    = "channel_not_found"
	CodeNotInChannel       = "not_in_channel"
	CodeInvalidAuth        = "invalid_auth"
	CodeMissingScope       = "missing_scope"
	CodeNoText             = "no_text"
	CodeNoDefaultChannel   = "no_default_channel"
	CodeThreadNotFound     = "thread_not_found"
	CodeRateLimited        = "rate_limited"
	CodeTokenNotConfigured = "token_not_configured"
	CodeConfigParseError   = "config_parse_error"
	CodeNetworkError       = "network_error"
)

// エラーHintマップ（LLM向け・英語）
var hintMap = map[string]string{
	CodeChannelNotFound:    "The channel was not found. Ask the user to verify the channel name or ID. Do not include the '#' prefix.",
	CodeNotInChannel:       "The bot is not a member of this channel. Ask the user to invite the bot by running: /invite @slack-fast-mcp",
	CodeInvalidAuth:        "The Slack token is invalid or expired. Ask the user to regenerate the token at https://api.slack.com/apps",
	CodeMissingScope:       "Required OAuth scope is missing. Ask the user to add the missing scope in Slack App settings and reinstall the app.",
	CodeNoText:             "The message parameter is required and cannot be empty.",
	CodeNoDefaultChannel:   "No channel specified and no default_channel configured. Set default_channel in config or specify the channel parameter.",
	CodeThreadNotFound:     "The thread_ts does not match any existing message. Ask the user to verify the thread timestamp.",
	CodeRateLimited:        "Slack API rate limit reached. The tool will automatically retry. If this persists, wait a moment and try again.",
	CodeTokenNotConfigured: "No Slack token found. Ask the user to run 'slack-fast-mcp setup' or set the SLACK_BOT_TOKEN environment variable.",
	CodeConfigParseError:   "Failed to parse config file. Ask the user to verify the JSON syntax in .slack-mcp.json",
	CodeNetworkError:       "Failed to connect to Slack API. Check network connectivity and try again.",
}

// New は指定されたコードでAppErrorを生成する。
// Hintは自動的にhintMapから取得される。
func New(code, message string, err error) *AppError {
	hint := hintMap[code]
	return &AppError{
		Code:    code,
		Message: message,
		Hint:    hint,
		Err:     err,
	}
}

// NewWithHint はカスタムHintを指定してAppErrorを生成する。
func NewWithHint(code, message, hint string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Hint:    hint,
		Err:     err,
	}
}

// FormatForMCP はMCPツールエラーメッセージとしてフォーマットする。
func (e *AppError) FormatForMCP() string {
	return fmt.Sprintf("Error [%s]: %s\nHint: %s", e.Code, e.Message, e.Hint)
}

// MaskToken はトークン文字列をマスキングする。
// xoxb-xxxx → xoxb-****
func MaskToken(s string) string {
	if len(s) < 5 {
		return s
	}
	prefixes := []string{"xoxb-", "xoxp-", "xoxs-"}
	for _, prefix := range prefixes {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return prefix + "****"
		}
	}
	return s
}
