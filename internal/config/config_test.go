package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	return path
}

// --- C01: 正常なJSON設定ファイル読み込み ---
func TestLoad_ValidJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"${SLACK_BOT_TOKEN}","default_channel":"general","log_level":"debug"}`)
	t.Setenv(EnvSlackBotToken, "xoxb-test-token-value")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "xoxb-test-token-value" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-test-token-value")
	}
	if cfg.DefaultChannel != "general" {
		t.Errorf("DefaultChannel = %q, want %q", cfg.DefaultChannel, "general")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

// --- C02: 環境変数参照の展開 ---
func TestLoad_EnvVarExpansion(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"${MY_CUSTOM_TOKEN}","default_channel":"${MY_CHANNEL}"}`)
	t.Setenv("MY_CUSTOM_TOKEN", "xoxb-expanded-token")
	t.Setenv("MY_CHANNEL", "dev-channel")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "xoxb-expanded-token" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-expanded-token")
	}
	if cfg.DefaultChannel != "dev-channel" {
		t.Errorf("DefaultChannel = %q, want %q", cfg.DefaultChannel, "dev-channel")
	}
}

// --- C03: 環境変数未設定時の展開 ---
func TestLoad_EnvVarNotSet(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"${NONEXISTENT_VAR}"}`)
	// NONEXISTENT_VAR は設定しない

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "" {
		t.Errorf("Token = %q, want empty string", cfg.Token)
	}
}

// --- C04: 設定ファイル不在 ---
func TestLoad_NoConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvSlackBotToken, "xoxb-from-env")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "xoxb-from-env" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-from-env")
	}
}

// --- C05: 不正なJSON ---
func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{invalid json}`)

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "config_parse_error") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "config_parse_error")
	}
}

// --- C06: ローカル設定がグローバルを上書き ---
func TestMergeConfig(t *testing.T) {
	dst := &Config{
		Token:          "global-token",
		DefaultChannel: "global-channel",
		LogLevel:       "info",
	}
	src := &Config{
		Token:          "local-token",
		DefaultChannel: "", // 空 → 上書きしない
		LogLevel:       "debug",
	}

	mergeConfig(dst, src)

	if dst.Token != "local-token" {
		t.Errorf("Token = %q, want %q", dst.Token, "local-token")
	}
	if dst.DefaultChannel != "global-channel" {
		t.Errorf("DefaultChannel = %q, want %q", dst.DefaultChannel, "global-channel")
	}
	if dst.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", dst.LogLevel, "debug")
	}
}

// --- C07: 環境変数がローカル設定を上書き ---
func TestLoad_EnvOverridesLocal(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"local-token","default_channel":"local-channel"}`)
	t.Setenv(EnvSlackBotToken, "xoxb-env-override")
	t.Setenv(EnvSlackDefaultChannel, "env-channel")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "xoxb-env-override" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-env-override")
	}
	if cfg.DefaultChannel != "env-channel" {
		t.Errorf("DefaultChannel = %q, want %q", cfg.DefaultChannel, "env-channel")
	}
}

// --- C08: トークン直書き検出（警告出力） ---
func TestLoad_HardcodedTokenWarning(t *testing.T) {
	dir := t.TempDir()
	// 直書きトークン
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"xoxb-1234567890-abcdef"}`)

	// stderr をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.Close()
	os.Stderr = oldStderr

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if cfg.Token != "xoxb-1234567890-abcdef" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-1234567890-abcdef")
	}
	if !strings.Contains(output, "WARNING") {
		t.Errorf("expected warning on stderr, got %q", output)
	}
}

// --- C09: 空の設定ファイル ---
func TestLoad_EmptyConfigFile(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{}`)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "" {
		t.Errorf("Token = %q, want empty", cfg.Token)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q (default)", cfg.LogLevel, "warn")
	}
}

// --- C10: default_channel 未設定 ---
func TestLoad_NoDefaultChannel(t *testing.T) {
	dir := t.TempDir()
	writeTestConfig(t, dir, LocalConfigFile, `{"token":"${SLACK_BOT_TOKEN}"}`)
	t.Setenv(EnvSlackBotToken, "xoxb-test")

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultChannel != "" {
		t.Errorf("DefaultChannel = %q, want empty", cfg.DefaultChannel)
	}
}

// --- Validate テスト ---
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{"valid", &Config{Token: "xoxb-test"}, false},
		{"empty token", &Config{Token: ""}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- ResolveChannel テスト ---
func TestConfig_ResolveChannel(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *Config
		channel    string
		want       string
		wantErr    bool
	}{
		{"explicit channel", &Config{DefaultChannel: "default"}, "explicit", "explicit", false},
		{"default channel", &Config{DefaultChannel: "default"}, "", "default", false},
		{"no channel", &Config{DefaultChannel: ""}, "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.cfg.ResolveChannel(tt.channel)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveChannel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ResolveChannel() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- expandEnvVars テスト ---
func TestExpandEnvVars(t *testing.T) {
	t.Setenv("TEST_VAR", "expanded_value")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple expansion", "${TEST_VAR}", "expanded_value"},
		{"no expansion", "plain-text", "plain-text"},
		{"partial expansion", "prefix-${TEST_VAR}-suffix", "prefix-expanded_value-suffix"},
		{"unset variable", "${UNSET_VAR}", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// --- LoadFromPath テスト ---
func TestLoadFromPath(t *testing.T) {
	dir := t.TempDir()
	path := writeTestConfig(t, dir, "custom.json", `{"token":"${SLACK_BOT_TOKEN}","default_channel":"custom-ch"}`)
	t.Setenv(EnvSlackBotToken, "xoxb-custom-path")

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "xoxb-custom-path" {
		t.Errorf("Token = %q, want %q", cfg.Token, "xoxb-custom-path")
	}
	if cfg.DefaultChannel != "custom-ch" {
		t.Errorf("DefaultChannel = %q, want %q", cfg.DefaultChannel, "custom-ch")
	}
}

func TestLoadFromPath_NotFound(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
