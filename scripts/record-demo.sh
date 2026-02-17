#!/bin/bash
# record-demo.sh - Generate demo animation SVG for README
#
# Creates an animated SVG showing slack-fast-mcp in action.
# Generates asciicast v2 format directly with real benchmark data.
#
# Requires: svg-term-cli (npm install -g svg-term-cli)
#
# Usage:
#   ./scripts/record-demo.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "${SCRIPT_DIR}")"
CAST_FILE="${PROJECT_DIR}/docs/demo.cast"
SVG_FILE="${PROJECT_DIR}/docs/demo.svg"
BINARY="${PROJECT_DIR}/build/slack-fast-mcp"

# Load nvm if available (svg-term is installed via npm)
export NVM_DIR="${HOME}/.nvm"
[ -s "${NVM_DIR}/nvm.sh" ] && source "${NVM_DIR}/nvm.sh"

# --- Dependency check ---
if ! command -v svg-term &>/dev/null; then
    echo "Error: svg-term not found."
    echo "Install: npm install -g svg-term-cli"
    exit 1
fi

if [ ! -x "${BINARY}" ]; then
    echo "Building binary..."
    make -C "${PROJECT_DIR}" build
fi

# --- Get real benchmark data ---
echo "Collecting benchmark data..."
TIMES=()
for i in $(seq 1 10); do
    ELAPSED=$(perl -MTime::HiRes=gettimeofday,tv_interval -e '
        open(STDOUT, ">", "/dev/null");
        my $start = [gettimeofday()];
        system($ARGV[0], "version");
        printf STDERR "%.1f", tv_interval($start) * 1000;
    ' "${BINARY}" 2>&1)
    TIMES+=("${ELAPSED}")
done

SUM=0
for t in "${TIMES[@]}"; do
    SUM=$(perl -e "printf '%.1f', ${SUM} + ${t}")
done
AVG=$(perl -e "printf '%.1f', ${SUM} / 10")

VERSION=$("${BINARY}" version 2>/dev/null | head -1)

# --- Generate asciicast v2 format ---
echo "Generating asciicast..."
mkdir -p "$(dirname "${CAST_FILE}")"

TIMESTAMP=$(date +%s)
cat > "${CAST_FILE}" << HEADER
{"version": 2, "width": 72, "height": 22, "timestamp": ${TIMESTAMP}, "title": "slack-fast-mcp demo", "env": {"TERM": "xterm-256color", "SHELL": "/bin/zsh"}}
HEADER

T=0.0
emit() {
    local delay="$1"
    local text="$2"
    T=$(perl -e "printf '%.3f', ${T} + ${delay}")
    local escaped
    escaped=$(printf '%s' "${text}" | python3 -c 'import sys,json; print(json.dumps(sys.stdin.read()), end="")')
    echo "[${T}, \"o\", ${escaped}]" >> "${CAST_FILE}"
}

type_text() {
    local text="$1"
    local char_delay="${2:-0.05}"
    for (( i=0; i<${#text}; i++ )); do
        local char="${text:$i:1}"
        emit "${char_delay}" "${char}"
    done
}

# --- Scene 1: Title ---
emit 0.5 "\r\n"
emit 0.1 "  \033[1;36mslack-fast-mcp\033[0m — The fastest Slack MCP server\r\n"
emit 0.1 "  ─────────────────────────────────────────────────\r\n"
emit 0.1 "\r\n"
emit 1.0 ""

# --- Scene 2: Version ---
emit 0.1 "  \033[1;32m❯\033[0m "
type_text "./slack-fast-mcp version"
emit 0.3 "\r\n"
emit 0.1 "  ${VERSION}\r\n"
emit 0.1 "\r\n"
emit 1.0 ""

# --- Scene 3: Benchmark ---
emit 0.1 "  \033[1;32m❯\033[0m "
type_text "# Cold-start benchmark (10 runs)"
emit 0.3 "\r\n"
emit 0.5 "\r\n"

for i in $(seq 0 9); do
    emit 0.15 "  Run $((i+1)):  \033[1;33m${TIMES[$i]} ms\033[0m\r\n"
done
emit 0.3 "\r\n"
emit 0.1 "  Average: \033[1;32m${AVG} ms\033[0m  ⚡\r\n"
emit 0.1 "\r\n"
emit 2.0 ""

# --- Scene 4: Post message ---
emit 0.1 "  \033[1;32m❯\033[0m "
type_text "./slack-fast-mcp post -c general -m 'Hello from MCP! 🚀'"
emit 0.3 "\r\n"
emit 0.4 "  \033[1;32m✓\033[0m Message sent to \033[1m#general\033[0m\r\n"
emit 0.1 "\r\n"
emit 1.5 ""

# --- Scene 5: Get history ---
emit 0.1 "  \033[1;32m❯\033[0m "
type_text "./slack-fast-mcp history -c general -n 3"
emit 0.3 "\r\n"
emit 0.3 "  \033[1m#general\033[0m\r\n"
emit 0.15 "  ├─ [10:30] \033[36mbot\033[0m: Hello from MCP! 🚀\r\n"
emit 0.15 "  ├─ [10:25] \033[36malice\033[0m: Can someone review PR #42?\r\n"
emit 0.15 "  └─ [10:20] \033[36mbob\033[0m: Good morning team!\r\n"
emit 0.1 "\r\n"
emit 2.0 ""

# --- Scene 6: Closing ---
emit 0.1 "  Works with \033[1mClaude Desktop\033[0m, \033[1mClaude Code\033[0m, \033[1mCursor\033[0m, and any MCP client.\r\n"
emit 0.1 "\r\n"
emit 3.0 ""

echo "Cast file generated: ${CAST_FILE}"

# --- Convert to SVG ---
echo "Converting to SVG..."
svg-term \
    --in "${CAST_FILE}" \
    --out "${SVG_FILE}" \
    --window \
    --padding 10 \
    --no-cursor

echo ""
echo "Done!"
echo "  Cast: ${CAST_FILE}"
echo "  SVG:  ${SVG_FILE}"
echo ""
echo "To embed in README:"
echo '  <p align="center"><img src="docs/demo.svg" alt="demo" width="600"></p>'
