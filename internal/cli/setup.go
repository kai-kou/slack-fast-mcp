package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newSetupCmd は setup サブコマンドを作成する。
func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Long:  "Run the interactive setup wizard to configure slack-fast-mcp for your project.",
		RunE:  runSetup,
	}
}

// runSetup は対話形式の初期設定ウィザードを実行する。
func runSetup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "🚀 slack-fast-mcp Setup Wizard")
	fmt.Fprintln(out, strings.Repeat("─", 40))
	fmt.Fprintln(out, "")

	// Step 1: Slack App 作成確認
	fmt.Fprint(out, "Have you created a Slack App? (y/N): ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "📖 To create a Slack App:")
		fmt.Fprintln(out, "  1. Go to https://api.slack.com/apps")
		fmt.Fprintln(out, "  2. Click 'Create New App' → 'From scratch'")
		fmt.Fprintln(out, "  3. Name your app (e.g., 'slack-fast-mcp')")
		fmt.Fprintln(out, "  4. Go to 'OAuth & Permissions' → 'Bot Token Scopes'")
		fmt.Fprintln(out, "  5. Add the following scopes:")
		fmt.Fprintln(out, "     - chat:write")
		fmt.Fprintln(out, "     - channels:history")
		fmt.Fprintln(out, "     - channels:read")
		fmt.Fprintln(out, "     - users:read (optional, for username resolution)")
		fmt.Fprintln(out, "  6. Install the app to your workspace")
		fmt.Fprintln(out, "  7. Copy the 'Bot User OAuth Token' (starts with xoxb-)")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "After creating the app, run 'slack-fast-mcp setup' again.")
		return nil
	}

	// Step 2: Bot Token 入力
	fmt.Fprintln(out, "")
	var token string
	for {
		fmt.Fprint(out, "Enter your Bot User OAuth Token (xoxb-...): ")
		token, _ = reader.ReadString('\n')
		token = strings.TrimSpace(token)

		if strings.HasPrefix(token, "xoxb-") {
			break
		}
		fmt.Fprintln(out, "  ⚠️  Token must start with 'xoxb-'. Please try again.")
	}

	// Step 3: デフォルトチャンネル入力
	fmt.Fprintln(out, "")
	fmt.Fprint(out, "Enter default channel (leave empty to skip): ")
	defaultChannel, _ := reader.ReadString('\n')
	defaultChannel = strings.TrimSpace(defaultChannel)
	defaultChannel = strings.TrimPrefix(defaultChannel, "#")

	if defaultChannel == "" {
		fmt.Fprintln(out, "  ℹ️  No default channel set. You'll need to specify --channel for each command.")
	}

	// Step 4: .slack-mcp.json 生成
	fmt.Fprintln(out, "")
	configData := map[string]string{
		"token": "${SLACK_BOT_TOKEN}",
	}
	if defaultChannel != "" {
		configData["default_channel"] = defaultChannel
	}

	configJSON, _ := json.MarshalIndent(configData, "", "  ")

	configPath := filepath.Join(".", ".slack-mcp.json")
	if err := os.WriteFile(configPath, append(configJSON, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}
	fmt.Fprintf(out, "✅ Created %s\n", configPath)

	// Step 5: 環境変数の設定案内
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "📝 Set the SLACK_BOT_TOKEN environment variable:")
	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "  export SLACK_BOT_TOKEN='%s'\n", token)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "  Add this to your shell profile (~/.zshrc, ~/.bashrc) for persistence.")

	// Step 6: .gitignore 追記確認
	fmt.Fprintln(out, "")
	if !isInGitignore(".slack-mcp.json") {
		fmt.Fprint(out, "Add .slack-mcp.json to .gitignore? (Y/n): ")
		answer, _ = reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "n" && answer != "no" {
			if err := appendToGitignore(".slack-mcp.json"); err != nil {
				fmt.Fprintf(out, "  ⚠️  Could not update .gitignore: %v\n", err)
			} else {
				fmt.Fprintln(out, "  ✅ Added .slack-mcp.json to .gitignore")
			}
		}
	} else {
		fmt.Fprintln(out, "  ℹ️  .slack-mcp.json is already in .gitignore")
	}

	// Step 7: Cursor MCP 設定の案内
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "🔧 Cursor MCP Configuration")
	fmt.Fprintln(out, strings.Repeat("─", 40))
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Add to .cursor/mcp.json:")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, `  {`)
	fmt.Fprintln(out, `    "mcpServers": {`)
	fmt.Fprintln(out, `      "slack-fast-mcp": {`)
	fmt.Fprintln(out, `        "command": "/path/to/slack-fast-mcp",`)
	fmt.Fprintln(out, `        "args": [],`)
	fmt.Fprintln(out, `        "env": {`)
	fmt.Fprintln(out, `          "SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}"`)
	fmt.Fprintln(out, `        }`)
	fmt.Fprintln(out, `      }`)
	fmt.Fprintln(out, `    }`)
	fmt.Fprintln(out, `  }`)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Don't forget to invite the bot to your channel: /invite @slack-fast-mcp")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "🎉 Setup complete!")

	return nil
}

// isInGitignore は .gitignore にパターンが含まれているか確認する。
func isInGitignore(pattern string) bool {
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == pattern {
			return true
		}
	}
	return false
}

// appendToGitignore は .gitignore にパターンを追記する。
func appendToGitignore(pattern string) error {
	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// 末尾の改行を確認して必要なら追加
	data, _ := os.ReadFile(".gitignore")
	if len(data) > 0 && data[len(data)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	_, err = f.WriteString(pattern + "\n")
	return err
}
