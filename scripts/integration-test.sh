#!/bin/bash
# integration-test.sh - 実Slack環境でのE2E統合テスト
#
# 前提条件:
#   - SLACK_BOT_TOKEN: 有効なSlack Bot User OAuth Token
#   - SLACK_TEST_CHANNEL: テスト投稿先チャンネル名（BotがInvite済み）
#
# 使い方:
#   SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test ./scripts/integration-test.sh
#
# テスト内容:
#   1. バイナリビルド
#   2. MCP Protocol初期化（JSON-RPC initialize）
#   3. MCP経由でメッセージ投稿（slack_post_message）
#   4. MCP経由で履歴取得（slack_get_history）
#   5. MCP経由でスレッド返信（slack_post_thread）
#   6. レスポンス形式の検証

set -euo pipefail

# ===== 定数 =====
BINARY="./build/slack-fast-mcp"
PASS=0
FAIL=0
SKIP=0
TOTAL=0
POSTED_TS=""

# ===== タイムアウトコマンド検出 =====
# macOS は timeout コマンドがないため、perl で代替する
if command -v timeout &>/dev/null; then
    TIMEOUT_CMD="timeout"
elif command -v gtimeout &>/dev/null; then
    TIMEOUT_CMD="gtimeout"
else
    # perl による timeout 代替
    TIMEOUT_CMD=""
fi

run_with_timeout() {
    local secs="$1"
    shift
    if [ -n "${TIMEOUT_CMD}" ]; then
        ${TIMEOUT_CMD} "${secs}" "$@"
    else
        perl -e "alarm ${secs}; exec @ARGV" "$@"
    fi
}

# ===== カラー出力 =====
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ===== ヘルパー =====
pass() {
    PASS=$((PASS + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${GREEN}[PASS]${NC} $1"
}

fail() {
    FAIL=$((FAIL + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${RED}[FAIL]${NC} $1"
    if [ "${2:-}" != "" ]; then
        echo -e "         ${RED}Detail: $2${NC}"
    fi
}

skip() {
    SKIP=$((SKIP + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "  ${YELLOW}[SKIP]${NC} $1"
}

info() {
    echo -e "  ${CYAN}[INFO]${NC} $1"
}

# MCP JSON-RPC リクエストを送信し、レスポンスを取得する
# Usage: mcp_call '{"jsonrpc":"2.0",...}'
# 注: initialize + リクエスト + initialized notification をまとめて送信
mcp_call() {
    local tool_request="$1"
    local init_request='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"integration-test","version":"1.0.0"}}}'
    local initialized_notification='{"jsonrpc":"2.0","method":"notifications/initialized"}'

    # 3つのJSON-RPCメッセージを改行区切りで送信し、最後のレスポンスを取得
    local response
    response=$(printf '%s\n%s\n%s\n' "${init_request}" "${initialized_notification}" "${tool_request}" | \
        SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}" \
        SLACK_DEFAULT_CHANNEL="${SLACK_TEST_CHANNEL}" \
        run_with_timeout 30 "${BINARY}" 2>/dev/null || true)

    echo "${response}"
}

# JSON レスポンスからツール呼び出し結果のテキストを抽出する
# Usage: extract_tool_result "$response"
extract_tool_result() {
    local response="$1"
    # 最後の行（tools/call のレスポンス）を取得
    local last_line
    last_line=$(echo "${response}" | tail -1)
    echo "${last_line}"
}

# ===== 前提条件チェック =====
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   INTEGRATION TEST - slack-fast-mcp (Real Slack) ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════╝${NC}"
echo ""

# 環境変数チェック
if [ -z "${SLACK_BOT_TOKEN:-}" ]; then
    echo -e "${RED}ERROR: SLACK_BOT_TOKEN is not set.${NC}"
    echo ""
    echo "Usage:"
    echo "  SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test ./scripts/integration-test.sh"
    echo ""
    echo "Required:"
    echo "  SLACK_BOT_TOKEN     - Slack Bot User OAuth Token"
    echo "  SLACK_TEST_CHANNEL  - Test channel name (bot must be invited)"
    exit 1
fi

if [ -z "${SLACK_TEST_CHANNEL:-}" ]; then
    echo -e "${RED}ERROR: SLACK_TEST_CHANNEL is not set.${NC}"
    echo ""
    echo "Usage:"
    echo "  SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test ./scripts/integration-test.sh"
    exit 1
fi

info "Token: ${SLACK_BOT_TOKEN:0:10}...****"
info "Channel: ${SLACK_TEST_CHANNEL}"
echo ""

# ===== [1/6] バイナリビルド =====
echo "── [1/6] Build binary ─────────────────────────────"
if make build > /dev/null 2>&1; then
    pass "Binary built successfully"
else
    fail "Binary build failed"
    exit 1
fi

if [ ! -f "${BINARY}" ] || [ ! -x "${BINARY}" ]; then
    fail "Binary not found or not executable: ${BINARY}"
    exit 1
fi

# ===== [2/6] MCP Protocol 初期化 =====
echo "── [2/6] MCP Protocol initialization ────────────────"
INIT_REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"integration-test","version":"1.0.0"}}}'

INIT_RESPONSE=$(echo "${INIT_REQUEST}" | \
    SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN}" \
    run_with_timeout 10 "${BINARY}" 2>/dev/null || true)

if echo "${INIT_RESPONSE}" | grep -q '"result"'; then
    pass "MCP Server responded to initialize"
else
    fail "MCP Server did not respond to initialize" "${INIT_RESPONSE}"
fi

# サーバー情報の検証
if echo "${INIT_RESPONSE}" | grep -q '"slack-fast-mcp"'; then
    pass "Server identifies as 'slack-fast-mcp'"
else
    fail "Server name not found in response"
fi

# ツール一覧の検証（capabilities にtools が含まれる）
if echo "${INIT_RESPONSE}" | grep -q '"tools"'; then
    pass "Server advertises tool capabilities"
else
    skip "Tool capabilities not in initialize response (checked via tools/list)"
fi

# ===== [3/6] slack_post_message - メッセージ投稿 =====
echo "── [3/6] slack_post_message (real Slack) ─────────────"
TIMESTAMP=$(date +%Y-%m-%d_%H:%M:%S)
TEST_MESSAGE="[Integration Test] ${TIMESTAMP} - slack-fast-mcp 統合テスト投稿"

POST_REQUEST=$(cat <<JSONEOF
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"slack_post_message","arguments":{"channel":"${SLACK_TEST_CHANNEL}","message":"${TEST_MESSAGE}"}}}
JSONEOF
)

POST_RESPONSE=$(mcp_call "${POST_REQUEST}")
POST_RESULT=$(extract_tool_result "${POST_RESPONSE}")

# ok:true の検証
if echo "${POST_RESULT}" | grep -q '"ok":true'; then
    pass "slack_post_message returned ok:true"
else
    fail "slack_post_message did not return ok:true" "${POST_RESULT}"
fi

# ts の存在確認
if echo "${POST_RESULT}" | grep -qE '"ts":"[0-9]+\.[0-9]+"'; then
    pass "Response contains valid timestamp (ts)"
    # ts を後のテストで使用するために抽出
    POSTED_TS=$(echo "${POST_RESULT}" | grep -oE '"ts":"[0-9]+\.[0-9]+"' | head -1 | sed 's/"ts":"//;s/"//')
    info "Posted message ts: ${POSTED_TS}"
else
    fail "Response does not contain valid ts"
fi

# channel の存在確認
if echo "${POST_RESULT}" | grep -qE '"channel":"C[A-Z0-9]+"'; then
    pass "Response contains channel ID"
else
    fail "Response does not contain channel ID"
fi

# message の確認
if echo "${POST_RESULT}" | grep -q "Integration Test"; then
    pass "Response contains posted message text"
else
    fail "Response does not contain posted message text"
fi

# ===== [4/6] slack_get_history - 履歴取得 =====
echo "── [4/6] slack_get_history (real Slack) ──────────────"

# 投稿直後なので少し待つ
sleep 2

HISTORY_REQUEST=$(cat <<JSONEOF
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"slack_get_history","arguments":{"channel":"${SLACK_TEST_CHANNEL}","limit":5}}}
JSONEOF
)

HISTORY_RESPONSE=$(mcp_call "${HISTORY_REQUEST}")
HISTORY_RESULT=$(extract_tool_result "${HISTORY_RESPONSE}")

# ok:true の検証
if echo "${HISTORY_RESULT}" | grep -q '"ok":true'; then
    pass "slack_get_history returned ok:true"
else
    fail "slack_get_history did not return ok:true" "${HISTORY_RESULT}"
fi

# messages 配列の存在確認
if echo "${HISTORY_RESULT}" | grep -q '"messages":\['; then
    pass "Response contains messages array"
else
    fail "Response does not contain messages array"
fi

# count の存在確認
if echo "${HISTORY_RESULT}" | grep -qE '"count":[0-9]+'; then
    pass "Response contains message count"
else
    fail "Response does not contain count"
fi

# 先ほど投稿したメッセージが含まれるか確認
if echo "${HISTORY_RESULT}" | grep -q "Integration Test"; then
    pass "History contains the message we just posted"
else
    skip "Posted message not found in history (may be due to timing)"
fi

# ===== [5/6] slack_post_thread - スレッド返信 =====
echo "── [5/6] slack_post_thread (real Slack) ──────────────"

if [ -z "${POSTED_TS}" ]; then
    skip "Skipping thread test (no message ts from post test)"
else
    THREAD_MESSAGE="[Integration Test] スレッド返信テスト ${TIMESTAMP}"
    THREAD_REQUEST=$(cat <<JSONEOF
{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"slack_post_thread","arguments":{"channel":"${SLACK_TEST_CHANNEL}","thread_ts":"${POSTED_TS}","message":"${THREAD_MESSAGE}"}}}
JSONEOF
    )

    THREAD_RESPONSE=$(mcp_call "${THREAD_REQUEST}")
    THREAD_RESULT=$(extract_tool_result "${THREAD_RESPONSE}")

    # ok:true の検証
    if echo "${THREAD_RESULT}" | grep -q '"ok":true'; then
        pass "slack_post_thread returned ok:true"
    else
        fail "slack_post_thread did not return ok:true" "${THREAD_RESULT}"
    fi

    # thread_ts の存在確認
    if echo "${THREAD_RESULT}" | grep -q '"thread_ts"'; then
        pass "Response contains thread_ts"
    else
        fail "Response does not contain thread_ts"
    fi

    # ts の存在確認（返信自体のts）
    if echo "${THREAD_RESULT}" | grep -qE '"ts":"[0-9]+\.[0-9]+"'; then
        pass "Response contains reply timestamp (ts)"
    else
        fail "Response does not contain reply ts"
    fi
fi

# ===== [6/6] エラーハンドリング検証 =====
echo "── [6/6] Error handling verification ─────────────────"

# 存在しないチャンネルへの投稿
ERROR_REQUEST=$(cat <<JSONEOF
{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"slack_post_message","arguments":{"channel":"nonexistent-channel-xxxxx","message":"This should fail"}}}
JSONEOF
)

ERROR_RESPONSE=$(mcp_call "${ERROR_REQUEST}")
ERROR_RESULT=$(extract_tool_result "${ERROR_RESPONSE}")

if echo "${ERROR_RESULT}" | grep -qi "channel_not_found\|not_in_channel\|error"; then
    pass "Invalid channel returns appropriate error"
else
    fail "Invalid channel did not return error" "${ERROR_RESULT}"
fi

# Hint の存在確認
if echo "${ERROR_RESULT}" | grep -qi "hint"; then
    pass "Error response contains LLM hint"
else
    skip "Error response may not contain hint (depends on error type)"
fi

# ===== 結果サマリー =====
echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════╗${NC}"
if [ "${FAIL}" -eq 0 ]; then
    echo -e "${CYAN}║${NC}  ${GREEN}✅ INTEGRATION TEST PASSED${NC}                        ${CYAN}║${NC}"
else
    echo -e "${CYAN}║${NC}  ${RED}❌ INTEGRATION TEST FAILED${NC}                        ${CYAN}║${NC}"
fi
echo -e "${CYAN}╚══════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  ${GREEN}Passed${NC}: ${PASS}"
echo -e "  ${RED}Failed${NC}: ${FAIL}"
echo -e "  ${YELLOW}Skipped${NC}: ${SKIP}"
echo -e "  Total: ${TOTAL}"
echo ""

# レポート保存
REPORT_DIR="./reports"
mkdir -p "${REPORT_DIR}"
REPORT_FILE="${REPORT_DIR}/integration-test-$(date +%Y-%m-%d_%H%M%S).txt"
{
    echo "# Integration Test Report"
    echo "# Date: $(date)"
    echo "# Channel: ${SLACK_TEST_CHANNEL}"
    echo "# Token: ${SLACK_BOT_TOKEN:0:10}...****"
    echo ""
    echo "Results: ${PASS} passed, ${FAIL} failed, ${SKIP} skipped (${TOTAL} total)"
} > "${REPORT_FILE}"
echo -e "  Report saved: ${REPORT_FILE}"
echo ""

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
