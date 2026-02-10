// Package slack provides a Slack API client with channel resolution and retry logic.
package slack

import "context"

// SlackClient はSlack API操作のインターフェース。
// テスト時にモック注入可能にするため、インターフェースとして定義する。
type SlackClient interface {
	// PostMessage はチャンネルにメッセージを投稿する。
	PostMessage(ctx context.Context, channel, message string) (*PostResult, error)

	// PostThread はスレッドに返信を投稿する。
	PostThread(ctx context.Context, channel, threadTS, message string) (*PostResult, error)

	// GetHistory はチャンネルの投稿履歴を取得する。
	GetHistory(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error)

	// ResolveChannel はチャンネル名をチャンネルIDに解決する。
	ResolveChannel(ctx context.Context, channel string) (string, error)
}

// PostResult はメッセージ投稿の結果。
type PostResult struct {
	Channel     string `json:"channel"`
	ChannelName string `json:"channel_name"`
	TS          string `json:"ts"`
	ThreadTS    string `json:"thread_ts,omitempty"`
	Message     string `json:"message"`
	Permalink   string `json:"permalink"`
}

// HistoryOptions は履歴取得のオプション。
type HistoryOptions struct {
	Limit  int    `json:"limit"`
	Oldest string `json:"oldest,omitempty"`
	Latest string `json:"latest,omitempty"`
}

// HistoryResult は履歴取得の結果。
type HistoryResult struct {
	Channel     string           `json:"channel"`
	ChannelName string           `json:"channel_name"`
	Messages    []HistoryMessage `json:"messages"`
	HasMore     bool             `json:"has_more"`
	Count       int              `json:"count"`
}

// HistoryMessage は履歴内の個別メッセージ。
type HistoryMessage struct {
	User       string `json:"user"`
	UserName   string `json:"user_name"`
	Text       string `json:"text"`
	TS         string `json:"ts"`
	ThreadTS   string `json:"thread_ts,omitempty"`
	ReplyCount int    `json:"reply_count"`
	Permalink  string `json:"permalink"`
}
