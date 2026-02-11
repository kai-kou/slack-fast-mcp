package slack

import "context"

// MockClient はテスト用のSlackClientモック実装。
// 各メソッドのFunc フィールドにテスト用の関数を設定して使用する。
type MockClient struct {
	PostMessageFunc    func(ctx context.Context, channel, message string) (*PostResult, error)
	PostThreadFunc     func(ctx context.Context, channel, threadTS, message string) (*PostResult, error)
	GetHistoryFunc     func(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error)
	AddReactionFunc    func(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error)
	RemoveReactionFunc func(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error)
	ResolveChannelFunc func(ctx context.Context, channel string) (string, error)
}

// PostMessage calls the mock function.
func (m *MockClient) PostMessage(ctx context.Context, channel, message string) (*PostResult, error) {
	if m.PostMessageFunc != nil {
		return m.PostMessageFunc(ctx, channel, message)
	}
	return &PostResult{}, nil
}

// PostThread calls the mock function.
func (m *MockClient) PostThread(ctx context.Context, channel, threadTS, message string) (*PostResult, error) {
	if m.PostThreadFunc != nil {
		return m.PostThreadFunc(ctx, channel, threadTS, message)
	}
	return &PostResult{}, nil
}

// GetHistory calls the mock function.
func (m *MockClient) GetHistory(ctx context.Context, channel string, opts HistoryOptions) (*HistoryResult, error) {
	if m.GetHistoryFunc != nil {
		return m.GetHistoryFunc(ctx, channel, opts)
	}
	return &HistoryResult{}, nil
}

// AddReaction calls the mock function.
func (m *MockClient) AddReaction(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error) {
	if m.AddReactionFunc != nil {
		return m.AddReactionFunc(ctx, channel, timestamp, reaction)
	}
	return &ReactionResult{}, nil
}

// RemoveReaction calls the mock function.
func (m *MockClient) RemoveReaction(ctx context.Context, channel, timestamp, reaction string) (*ReactionResult, error) {
	if m.RemoveReactionFunc != nil {
		return m.RemoveReactionFunc(ctx, channel, timestamp, reaction)
	}
	return &ReactionResult{}, nil
}

// ResolveChannel calls the mock function.
func (m *MockClient) ResolveChannel(ctx context.Context, channel string) (string, error) {
	if m.ResolveChannelFunc != nil {
		return m.ResolveChannelFunc(ctx, channel)
	}
	return channel, nil
}
