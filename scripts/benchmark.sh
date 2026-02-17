#!/bin/bash
# benchmark.sh - Startup time benchmark for slack-fast-mcp
#
# Measures cold-start time of the binary to verify the "~10ms" claim.
# MCP servers spawn a new process per request, so startup speed = user-perceived latency.
#
# Uses perl system() for accurate measurement (avoids shell subshell overhead).
#
# Usage:
#   ./scripts/benchmark.sh              # default: 50 iterations
#   ./scripts/benchmark.sh 100          # custom iteration count
#   ./scripts/benchmark.sh 50 ./path/to/binary

set -euo pipefail

ITERATIONS="${1:-50}"
BINARY="${2:-./build/slack-fast-mcp}"

# --- Dependency check ---
if [ ! -x "${BINARY}" ]; then
    echo "Error: Binary not found: ${BINARY}"
    echo "Run 'make build' first."
    exit 1
fi

if ! perl -MTime::HiRes -e '1' 2>/dev/null; then
    echo "Error: Perl Time::HiRes module required (pre-installed on macOS/most Linux)."
    exit 1
fi

# --- Collect measurements ---
echo "Benchmark: slack-fast-mcp startup time"
echo "======================================="
echo "Binary:     ${BINARY}"
echo "Version:    $(${BINARY} version 2>/dev/null | head -1)"
echo "Iterations: ${ITERATIONS}"
echo "Platform:   $(uname -ms)"
echo ""

RESULTS_FILE=$(mktemp)
trap 'rm -f "${RESULTS_FILE}"' EXIT

# Warm up (1 run to ensure binary is in disk cache)
"${BINARY}" version > /dev/null 2>&1

echo -n "Running"
# Use perl system() for accurate timing (avoids shell subshell overhead)
perl -MTime::HiRes=gettimeofday,tv_interval -e '
    my $binary = $ARGV[0];
    my $iterations = $ARGV[1];
    my $results_file = $ARGV[2];
    open(my $fh, ">", $results_file) or die "Cannot open $results_file: $!";
    open(STDOUT, ">", "/dev/null") or die "Cannot redirect stdout: $!";
    for my $i (1..$iterations) {
        my $start = [gettimeofday()];
        system($binary, "version") == 0 or next;
        my $elapsed = tv_interval($start) * 1000;
        printf $fh "%.3f\n", $elapsed;
        if ($i % 10 == 0) {
            printf STDERR " %d", $i;
        } else {
            printf STDERR ".";
        }
    }
    print STDERR " done\n";
' "${BINARY}" "${ITERATIONS}" "${RESULTS_FILE}" 2>&1
echo ""

# --- Calculate statistics ---
SORTED=$(sort -n "${RESULTS_FILE}")

COUNT=$(wc -l < "${RESULTS_FILE}" | tr -d ' ')
if [ "${COUNT}" -eq 0 ]; then
    echo "Error: No measurements collected."
    exit 1
fi

MIN=$(echo "${SORTED}" | head -1)
MAX=$(echo "${SORTED}" | tail -1)
SUM=$(paste -sd+ "${RESULTS_FILE}" | bc -l)
AVG=$(echo "scale=3; ${SUM} / ${COUNT}" | bc -l)

# Median (middle value)
MID=$((COUNT / 2))
if [ $((COUNT % 2)) -eq 0 ]; then
    V1=$(echo "${SORTED}" | sed -n "${MID}p")
    V2=$(echo "${SORTED}" | sed -n "$((MID + 1))p")
    MEDIAN=$(echo "scale=3; (${V1} + ${V2}) / 2" | bc -l)
else
    MEDIAN=$(echo "${SORTED}" | sed -n "$((MID + 1))p")
fi

# P95 (95th percentile)
P95_IDX=$(echo "scale=0; ${COUNT} * 95 / 100" | bc)
[ "${P95_IDX}" -lt 1 ] && P95_IDX=1
P95=$(echo "${SORTED}" | sed -n "${P95_IDX}p")

# P99 (99th percentile)
P99_IDX=$(echo "scale=0; ${COUNT} * 99 / 100" | bc)
[ "${P99_IDX}" -lt 1 ] && P99_IDX=1
P99=$(echo "${SORTED}" | sed -n "${P99_IDX}p")

# --- Output results ---
echo "Results"
echo "-------"
printf "  Min:    %8s ms\n" "${MIN}"
printf "  Max:    %8s ms\n" "${MAX}"
printf "  Avg:    %8s ms\n" "${AVG}"
printf "  Median: %8s ms\n" "${MEDIAN}"
printf "  P95:    %8s ms\n" "${P95}"
printf "  P99:    %8s ms\n" "${P99}"
echo ""

# --- Pass/Fail judgment ---
THRESHOLD=15  # ms
PASS_COUNT=$(echo "${SORTED}" | awk -v t="${THRESHOLD}" '$1 <= t { n++ } END { print n+0 }')
PASS_RATE=$(echo "scale=1; ${PASS_COUNT} * 100 / ${COUNT}" | bc -l)

echo "Threshold: ${THRESHOLD}ms"
printf "Pass rate: %s/%s (%.1f%%)\n" "${PASS_COUNT}" "${COUNT}" "${PASS_RATE}"
echo ""

# Verdict: median should be within 10ms claim
if [ "$(echo "${MEDIAN} <= 10" | bc -l)" -eq 1 ]; then
    echo "Verdict: PASS -- median ${MEDIAN}ms confirms ~10ms startup claim"
    exit 0
elif [ "$(echo "${MEDIAN} <= 15" | bc -l)" -eq 1 ]; then
    echo "Verdict: PASS -- median ${MEDIAN}ms is within acceptable range of ~10ms claim"
    exit 0
else
    echo "Verdict: FAIL -- median ${MEDIAN}ms exceeds ~10ms claim"
    exit 1
fi
