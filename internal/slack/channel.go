package slack

import (
	"context"
	"regexp"
	"strings"

	apperr "github.com/kai-ko/slack-fast-mcp/internal/errors"
	slackapi "github.com/slack-go/slack"
)

// channelIDPattern はチャンネルIDの形式を判定する正規表現。
// C: パブリックチャンネル, G: プライベートチャンネル/グループ, D: ダイレクトメッセージ
var channelIDPattern = regexp.MustCompile(`^[CGD][A-Z0-9]{8,}$`)

// IsChannelID は文字列がチャンネルID形式かどうかを判定する。
func IsChannelID(s string) bool {
	return channelIDPattern.MatchString(s)
}

// resolveChannel はチャンネル名をチャンネルIDに変換する。
// 1. チャンネルID形式ならそのまま返す
// 2. "#" 付きなら除去してチャンネル名として検索
// 3. それ以外は conversations.list で検索
func (c *Client) resolveChannel(ctx context.Context, channel string) (string, error) {
	// チャンネルIDならそのまま返す
	if IsChannelID(channel) {
		return channel, nil
	}

	// "#" プレフィックスを除去
	name := strings.TrimPrefix(channel, "#")

	// キャッシュを確認
	if id, ok := c.channelCache[name]; ok {
		return id, nil
	}

	// conversations.list で検索
	id, err := c.findChannelByName(ctx, name)
	if err != nil {
		return "", err
	}

	// キャッシュに保存
	c.channelCache[name] = id
	return id, nil
}

// findChannelByName は conversations.list API でチャンネル名からIDを検索する。
// ページネーション: cursor ベース、最大5ページ（1000チャンネル）で打ち切り。
func (c *Client) findChannelByName(ctx context.Context, name string) (string, error) {
	params := &slackapi.GetConversationsParameters{
		Types:           []string{"public_channel", "private_channel"},
		Limit:           200,
		ExcludeArchived: true,
	}

	maxPages := 5
	for page := 0; page < maxPages; page++ {
		channels, nextCursor, err := c.api.GetConversationsContext(ctx, params)
		if err != nil {
			return "", classifySlackError(err)
		}

		for _, ch := range channels {
			if ch.Name == name {
				return ch.ID, nil
			}
		}

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	return "", apperr.New(apperr.CodeChannelNotFound,
		"指定されたチャンネルが見つかりません: "+name, nil)
}
