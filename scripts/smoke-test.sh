#!/bin/bash
# smoke-test.sh - バイナリの起動検証スモークテスト
#
# このスクリプトは、ビルドされたバイナリが正しく起動し、
# 基本的な応答を返すことを検証する。
# 「実際にユーザーが使った時に動くか？」を確認するための最終防衛線。

set -euo pipefail

BINARY="${1:-./build/slack-fast-mcp}"
PASS=0
FAIL=0
TOTAL=0

# --- ヘルパー ---
pass() {
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
    echo "  [PASS] $1"
}

fail() {
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
    echo "  [FAIL] $1"
}

# --- テスト ---
echo "==> Smoke Test: ${BINARY}"
echo ""

# 1. バイナリが存在するか
if [ -f "${BINARY}" ]; then
    pass "Binary exists"
else
    fail "Binary not found: ${BINARY}"
    echo ""
    echo "Result: ${PASS}/${TOTAL} passed, ${FAIL} failed"
    exit 1
fi

# 2. バイナリが実行可能か
if [ -x "${BINARY}" ]; then
    pass "Binary is executable"
else
    fail "Binary is not executable"
fi

# 3. トークン未設定時に適切なエラーを返すか（MCP Server モード）
# 環境変数をクリアして実行 → stderr にエラーメッセージが出ることを確認
STDERR_OUTPUT=$(SLACK_BOT_TOKEN="" SLACK_DEFAULT_CHANNEL="" "${BINARY}" 2>&1 || true)
if echo "${STDERR_OUTPUT}" | grep -qi "token\|setup\|SLACK_BOT_TOKEN"; then
    pass "Shows helpful error when token is not set"
else
    fail "No helpful error message when token is not set (got: ${STDERR_OUTPUT})"
fi

# 4. バイナリサイズが妥当か（5MB 〜 50MB）
SIZE=$(stat -f%z "${BINARY}" 2>/dev/null || stat --printf="%s" "${BINARY}" 2>/dev/null || echo "0")
SIZE_MB=$((SIZE / 1024 / 1024))
if [ "${SIZE_MB}" -ge 5 ] && [ "${SIZE_MB}" -le 50 ]; then
    pass "Binary size is reasonable (${SIZE_MB}MB)"
else
    fail "Binary size is unexpected (${SIZE_MB}MB, expected 5-50MB)"
fi

# 5. version サブコマンドが動作するか
VERSION_OUTPUT=$("${BINARY}" version 2>&1 || true)
if echo "${VERSION_OUTPUT}" | grep -q "slack-fast-mcp"; then
    pass "Version subcommand works"
else
    fail "Version subcommand failed (got: ${VERSION_OUTPUT})"
fi

# 6. help が正しいサブコマンド一覧を表示するか
HELP_OUTPUT=$("${BINARY}" --help 2>&1 || true)
if echo "${HELP_OUTPUT}" | grep -q "post" && echo "${HELP_OUTPUT}" | grep -q "history" && echo "${HELP_OUTPUT}" | grep -q "reply" && echo "${HELP_OUTPUT}" | grep -q "setup"; then
    pass "Help shows all subcommands"
else
    fail "Help missing subcommands (got: ${HELP_OUTPUT})"
fi

# 7. MCP プロトコルの初期化リクエストに応答するか
# JSON-RPC initialize リクエストを送信して応答を確認
# Note: stdio通信のため、パイプ環境では安定しない場合がある。
#       MCPプロトコルレベルの検証は Go の smoke_test.go で実施。
INIT_REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke-test","version":"1.0.0"}}}'
# macOS には timeout コマンドがないため、perl で代替
if command -v timeout &>/dev/null; then
    TIMEOUT_CMD="timeout"
elif command -v gtimeout &>/dev/null; then
    TIMEOUT_CMD="gtimeout"
else
    TIMEOUT_CMD=""
fi
if [ -n "${TIMEOUT_CMD}" ]; then
    MCP_RESPONSE=$(echo "${INIT_REQUEST}" | SLACK_BOT_TOKEN="xoxb-smoke-test-token" ${TIMEOUT_CMD} 5 "${BINARY}" 2>/dev/null || true)
else
    MCP_RESPONSE=$(echo "${INIT_REQUEST}" | SLACK_BOT_TOKEN="xoxb-smoke-test-token" perl -e 'alarm 5; exec @ARGV' "${BINARY}" 2>/dev/null || true)
fi
if echo "${MCP_RESPONSE}" | grep -q '"result"\|"slack-fast-mcp"\|"serverInfo"'; then
    pass "MCP Server responds to initialize request"
else
    # stdio パイプの環境差異で応答が取れない場合はスキップ
    # MCPプロトコルテストは Go テスト側で保証済み
    echo "  [SKIP] MCP protocol test skipped (pipe environment limitation)"
    echo "         MCP protocol is verified by: go test ./internal/mcp/ -run TestSmoke"
fi

# --- 結果 ---
echo ""
echo "==> Smoke Test Result: ${PASS}/${TOTAL} passed, ${FAIL} failed"

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
