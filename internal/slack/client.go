package slack

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	apperr "github.com/kai-kou/slack-fast-mcp/internal/errors"
	slackapi "github.com/slack-go/slack"
)

// maxRetries はレート制限時の最大リトライ回数。
const maxRetries = 3

// Client はSlack APIクライアントの実装。
type Client struct {
	api          *slackapi.Client
	channelCache map[string]string // チャンネル名 → ID キャッシュ
}

// NewClient は新しいSlackクライアントを作成する。
func NewClient(token string) *Client {
	return &Client{
		api:          slackapi.New(token),
		channelCache: make(map[string]string),
	}
}

// NewClientWithAPI は既存のslack.Clientを使用してクライアントを作成する（テスト用）。
func NewClientWithAPI(api *slackapi.Client) *Client {
	return &Client{
		api:          api,
		channelCache: make(map[string]string),
	}
}

// PostMessage はチャンネルにメッセージを投稿する。
func (c *Client) PostMessage(ctx context.Context, channel, message string) (*PostResult, error) {
	channelID, err := c.resolveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	var respChannel, respTS string
	err = c.withRetry(ctx, func() error {
		var e error
		respChannel, respTS, e = c.api.PostMessageContext(ctx, channelID,
			slackapi.MsgOptionText(message, false),
		)
		return e
	})
	if err != nil {
		return nil, classifySlackError(err)
	}

	permalink, _ := c.api.GetPermalinkContext(ctx, &slackapi.PermalinkParameters{
		Channel: respChannel,
		Ts:      respTS,
	})

	return &PostResult{
		Channel:     respChannel,
		ChannelName: c.getChannelName(channel, channelID),
		TS:          respTS,
		Message:     message,
		Permalink:   permalink,
	}, nil
}

// PostThread はスレッドに返信を投稿する。
func (c *Client) PostThread(ctx context.Context, channel, threadTS, message string) (*PostResult, error) {
	channelID, err := c.resolveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	var respChannel, respTS string
	err = c.withRetry(ctx, func() error {
		var e error
		respChannel, respTS, e = c.api.PostMessageContext(ctx, channelID,
			slackapi.MsgOptionText(message, false),
			slackapi.MsgOptionTS(threadTS),
		)
		return e
	})
	if err != nil {
		return nil, classifySlackError(err)
	}

	permalink, _ := c.api.GetPermalinkContext(ctx, &slackapi.PermalinkParameters{
		Channel: respChannel,
		Ts:      respTS,
	})

	return &PostResult{
		Channel:     respChannel,
		ChannelName: c.getChannelName(channel, channelID),
		TS:          respTS,
		ThreadTS:    threadTS,
		Message:     message,
		Permalink:   permalink,
	}, nil
}

// GetHistory はチャンネルの投稿履歴を取得する。
func (c *Client) GetHistory(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error) {
	channelID, err := c.resolveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	// limit のバリデーション
	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	params := &slackapi.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     limit,
	}
	if opts.Oldest != "" {
		params.Oldest = opts.Oldest
	}
	if opts.Latest != "" {
		params.Latest = opts.Latest
	}

	var resp *slackapi.GetConversationHistoryResponse
	err = c.withRetry(ctx, func() error {
		var e error
		resp, e = c.api.GetConversationHistoryContext(ctx, params)
		return e
	})
	if err != nil {
		return nil, classifySlackError(err)
	}
	if resp != nil && !resp.Ok && resp.Error != "" {
		return nil, classifySlackErrorString(resp.Error)
	}

	messages := make([]HistoryMessage, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		hm := HistoryMessage{
			User:       msg.User,
			Text:       msg.Text,
			TS:         msg.Timestamp,
			ThreadTS:   msg.ThreadTimestamp,
			ReplyCount: msg.ReplyCount,
		}

		// ユーザー名解決（ベストエフォート）
		if msg.User != "" {
			if user, err := c.api.GetUserInfoContext(ctx, msg.User); err == nil {
				hm.UserName = user.Name
			}
		}

		// パーマリンク取得（ベストエフォート）
		if permalink, err := c.api.GetPermalinkContext(ctx, &slackapi.PermalinkParameters{
			Channel: channelID,
			Ts:      msg.Timestamp,
		}); err == nil {
			hm.Permalink = permalink
		}

		messages = append(messages, hm)
	}

	return &HistoryResult{
		Channel:     channelID,
		ChannelName: c.getChannelName(channel, channelID),
		Messages:    messages,
		HasMore:     resp.HasMore,
		Count:       len(messages),
	}, nil
}

// AddReaction はメッセージにリアクション（絵文字）を追加する。
func (c *Client) AddReaction(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error) {
	channelID, err := c.resolveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	itemRef := slackapi.NewRefToMessage(channelID, timestamp)
	err = c.withRetry(ctx, func() error {
		return c.api.AddReactionContext(ctx, reaction, itemRef)
	})
	if err != nil {
		return nil, classifySlackError(err)
	}

	return &ReactionResult{
		Channel:     channelID,
		ChannelName: c.getChannelName(channel, channelID),
		Timestamp:   timestamp,
		Reaction:    reaction,
	}, nil
}

// RemoveReaction はメッセージからリアクション（絵文字）を削除する。
func (c *Client) RemoveReaction(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error) {
	channelID, err := c.resolveChannel(ctx, channel)
	if err != nil {
		return nil, err
	}

	itemRef := slackapi.NewRefToMessage(channelID, timestamp)
	err = c.withRetry(ctx, func() error {
		return c.api.RemoveReactionContext(ctx, reaction, itemRef)
	})
	if err != nil {
		return nil, classifySlackError(err)
	}

	return &ReactionResult{
		Channel:     channelID,
		ChannelName: c.getChannelName(channel, channelID),
		Timestamp:   timestamp,
		Reaction:    reaction,
	}, nil
}

// ResolveChannel はチャンネル名をチャンネルIDに解決する（公開メソッド）。
func (c *Client) ResolveChannel(ctx context.Context, channel string) (string, error) {
	return c.resolveChannel(ctx, channel)
}

// withRetry はレート制限時に指数バックオフでリトライする。
func (c *Client) withRetry(ctx context.Context, fn func() error) error {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// レート制限エラーの場合はリトライ
		if rateLimitErr, ok := err.(*slackapi.RateLimitedError); ok {
			if attempt >= maxRetries {
				return apperr.New(apperr.CodeRateLimited,
					"レート制限に到達しました", err)
			}

			// Retry-After ヘッダに従う
			waitDuration := rateLimitErr.RetryAfter
			if waitDuration == 0 {
				// フォールバック: 指数バックオフ (1s, 2s, 4s)
				waitDuration = time.Duration(math.Pow(2, float64(attempt))) * time.Second
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
				continue
			}
		}

		// レート制限以外のエラーはリトライしない
		return err
	}
	return nil
}

// classifySlackError はSlack APIエラーをAppErrorに変換する。
func classifySlackError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	return classifySlackErrorString(errStr)
}

// classifySlackErrorString はSlack APIエラー文字列をAppErrorに変換する。
func classifySlackErrorString(errStr string) error {
	switch {
	case strings.Contains(errStr, "channel_not_found"):
		return apperr.New(apperr.CodeChannelNotFound, "指定されたチャンネルが見つかりません", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "not_in_channel"):
		return apperr.New(apperr.CodeNotInChannel, "Botがチャンネルに参加していません", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "invalid_auth"), strings.Contains(errStr, "not_authed"):
		return apperr.New(apperr.CodeInvalidAuth, "トークンが無効です", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "missing_scope"):
		return apperr.New(apperr.CodeMissingScope, "必要なOAuthスコープが不足しています", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "thread_not_found"):
		return apperr.New(apperr.CodeThreadNotFound, "スレッド元メッセージが見つかりません", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "no_text"):
		return apperr.New(apperr.CodeNoText, "メッセージが空です", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "already_reacted"):
		return apperr.New(apperr.CodeAlreadyReacted, "既にこの絵文字でリアクション済みです", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "no_reaction"):
		return apperr.New(apperr.CodeNoReaction, "この絵文字のリアクションが存在しません", fmt.Errorf("%s", errStr))
	case strings.Contains(errStr, "invalid_name"):
		return apperr.New(apperr.CodeInvalidReaction, "絵文字名が無効です", fmt.Errorf("%s", errStr))
	default:
		return apperr.New(apperr.CodeNetworkError, "Slack APIへの接続に失敗しました", fmt.Errorf("%s", errStr))
	}
}

// getChannelName はキャッシュからチャンネル名を取得する。
// 元の入力がチャンネル名であればそれを返し、IDの場合は空文字を返す。
func (c *Client) getChannelName(input, resolvedID string) string {
	if IsChannelID(input) {
		// 逆引きキャッシュ確認
		for name, id := range c.channelCache {
			if id == resolvedID {
				return name
			}
		}
		return ""
	}
	return strings.TrimPrefix(input, "#")
}
