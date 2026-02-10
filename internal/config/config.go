// Package config handles configuration loading, merging, and environment variable expansion.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	apperr "github.com/kai-kou/slack-fast-mcp/internal/errors"
)

// Config はアプリケーション設定を表す構造体。
type Config struct {
	Token          string `json:"token"`
	DefaultChannel string `json:"default_channel"`
	DisplayName    string `json:"display_name"`
	LogLevel       string `json:"log_level"`
}

const (
	// LocalConfigFile はプロジェクトローカル設定ファイル名。
	LocalConfigFile = ".slack-mcp.json"

	// GlobalConfigDir はグローバル設定ディレクトリのサブパス。
	GlobalConfigDir = "slack-fast-mcp"

	// GlobalConfigFile はグローバル設定ファイル名。
	GlobalConfigFile = "config.json"
)

// 環境変数名
const (
	EnvSlackBotToken        = "SLACK_BOT_TOKEN"
	EnvSlackDefaultChannel  = "SLACK_DEFAULT_CHANNEL"
	EnvSlackDisplayName     = "SLACK_DISPLAY_NAME"
	EnvSlackFastMCPLogLevel = "SLACK_FAST_MCP_LOG_LEVEL"
)

// envVarPattern は ${VAR_NAME} 形式の環境変数参照を検出する正規表現。
var envVarPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

// tokenPattern はトークン直書きを検出する正規表現。
var tokenPattern = regexp.MustCompile(`^xox[bps]-`)

// Load はプロジェクトディレクトリから設定を読み込む。
// 読み込み順序: グローバル設定 → ローカル設定（上書き） → 環境変数（上書き）
func Load(projectDir string) (*Config, error) {
	cfg := &Config{
		LogLevel: "warn", // デフォルト値
	}

	// 1. グローバル設定ファイル読み込み
	globalPath, err := globalConfigPath()
	if err == nil {
		if gcfg, err := loadFromFile(globalPath); err == nil {
			mergeConfig(cfg, gcfg)
		}
		// グローバル設定がなくてもエラーにしない
	}

	// 2. プロジェクトローカル設定ファイル読み込み
	localPath := filepath.Join(projectDir, LocalConfigFile)
	if lcfg, err := loadFromFile(localPath); err == nil {
		mergeConfig(cfg, lcfg)
	} else if !os.IsNotExist(err) {
		// ファイルが存在するがパースエラーの場合
		return nil, apperr.New(apperr.CodeConfigParseError,
			fmt.Sprintf("設定ファイルの解析に失敗しました: %s", localPath), err)
	}

	// 3. トークンの環境変数展開（${VAR} 形式）
	cfg.Token = expandEnvVars(cfg.Token)
	cfg.DefaultChannel = expandEnvVars(cfg.DefaultChannel)
	cfg.DisplayName = expandEnvVars(cfg.DisplayName)

	// 4. 環境変数で上書き
	if v := os.Getenv(EnvSlackBotToken); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv(EnvSlackDefaultChannel); v != "" {
		cfg.DefaultChannel = v
	}
	if v := os.Getenv(EnvSlackDisplayName); v != "" {
		cfg.DisplayName = v
	}
	if v := os.Getenv(EnvSlackFastMCPLogLevel); v != "" {
		cfg.LogLevel = v
	}

	// 5. トークン直書き検出（警告のみ）
	if tokenPattern.MatchString(cfg.Token) {
		// ファイルに直書きされていた場合のみ警告
		// (環境変数から取得した場合は正常なので警告しない)
		if rawToken := getRawTokenFromFiles(globalPath, localPath); tokenPattern.MatchString(rawToken) {
			fmt.Fprintf(os.Stderr, "WARNING: Token appears to be hardcoded in config file.\n")
			fmt.Fprintf(os.Stderr, "Consider using environment variable reference: \"${SLACK_BOT_TOKEN}\"\n")
			fmt.Fprintf(os.Stderr, "See: https://github.com/kai-kou/slack-fast-mcp#security\n")
		}
	}

	return cfg, nil
}

// LoadFromPath は指定パスの設定ファイルのみを読み込む（CLI の --config 用）。
func LoadFromPath(path string) (*Config, error) {
	cfg, err := loadFromFile(path)
	if err != nil {
		return nil, apperr.New(apperr.CodeConfigParseError,
			fmt.Sprintf("設定ファイルの解析に失敗しました: %s", path), err)
	}

	// 環境変数展開
	cfg.Token = expandEnvVars(cfg.Token)
	cfg.DefaultChannel = expandEnvVars(cfg.DefaultChannel)
	cfg.DisplayName = expandEnvVars(cfg.DisplayName)

	// 環境変数で上書き
	if v := os.Getenv(EnvSlackBotToken); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv(EnvSlackDefaultChannel); v != "" {
		cfg.DefaultChannel = v
	}
	if v := os.Getenv(EnvSlackDisplayName); v != "" {
		cfg.DisplayName = v
	}
	if v := os.Getenv(EnvSlackFastMCPLogLevel); v != "" {
		cfg.LogLevel = v
	}

	return cfg, nil
}

// Validate は設定の必須項目を検証する。
func (c *Config) Validate() error {
	if c.Token == "" {
		return apperr.New(apperr.CodeTokenNotConfigured,
			"トークンが設定されていません", nil)
	}
	return nil
}

// ResolveChannel はチャンネルを解決する。
// パラメータ指定 > デフォルトチャンネル の優先順位。
func (c *Config) ResolveChannel(channel string) (string, error) {
	if channel != "" {
		return channel, nil
	}
	if c.DefaultChannel != "" {
		return c.DefaultChannel, nil
	}
	return "", apperr.New(apperr.CodeNoDefaultChannel,
		"チャンネル未指定かつデフォルト未設定", nil)
}

// ResolveDisplayName は表示名を解決する。
// パラメータ指定 > デフォルト表示名 の優先順位。空の場合は空文字を返す（エラーにしない）。
func (c *Config) ResolveDisplayName(displayName string) string {
	if displayName != "" {
		return displayName
	}
	return c.DisplayName
}

// loadFromFile はJSONファイルからConfigを読み込む。
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// mergeConfig は src の非空フィールドで dst を上書きする。
func mergeConfig(dst, src *Config) {
	if src.Token != "" {
		dst.Token = src.Token
	}
	if src.DefaultChannel != "" {
		dst.DefaultChannel = src.DefaultChannel
	}
	if src.DisplayName != "" {
		dst.DisplayName = src.DisplayName
	}
	if src.LogLevel != "" {
		dst.LogLevel = src.LogLevel
	}
}

// expandEnvVars は文字列中の ${VAR_NAME} を環境変数の値に展開する。
func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// ${VAR_NAME} → VAR_NAME を抽出
		varName := match[2 : len(match)-1]
		return os.Getenv(varName)
	})
}

// globalConfigPath はグローバル設定ファイルのパスを返す。
func globalConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, GlobalConfigDir, GlobalConfigFile), nil
}

// getRawTokenFromFiles は設定ファイルから生のトークン値を取得する（警告判定用）。
func getRawTokenFromFiles(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var raw struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		// 環境変数参照でない場合（${...} を含まない）
		if raw.Token != "" && !strings.Contains(raw.Token, "${") {
			return raw.Token
		}
	}
	return ""
}
