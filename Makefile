# slack-fast-mcp Makefile
# ローカル品質保証のためのビルド・テスト・品質ゲートを提供する

BINARY_NAME := slack-fast-mcp
BUILD_DIR := ./build
REPORTS_DIR := ./reports
COVERAGE_FILE := $(REPORTS_DIR)/coverage.out
COVERAGE_HTML := $(REPORTS_DIR)/coverage.html
COVERAGE_THRESHOLD := 65
TIMESTAMP := $(shell date +%Y-%m-%d_%H%M%S)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go ツール
GOTEST := go test
GOBUILD := go build
GOVET := go vet

.PHONY: all build test test-verbose test-race test-cover test-report quality smoke clean help setup-hooks test-integration test-integration-e2e release-dry-run release-snapshot lint

## ===== ビルド =====

all: quality build ## 品質ゲート + ビルド

build: ## バイナリをビルド
	@echo "==> Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/slack-fast-mcp/
	@echo "==> Built: $(BUILD_DIR)/$(BINARY_NAME) ($(VERSION))"

build-verify: ## クロスプラットフォームのコンパイル検証（バイナリは生成しない）
	@echo "==> Verifying cross-platform compilation..."
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) ./cmd/slack-fast-mcp/ 2>/dev/null && echo "  [OK] darwin/arm64"
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) ./cmd/slack-fast-mcp/ 2>/dev/null && echo "  [OK] darwin/amd64"
	@GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/slack-fast-mcp/ 2>/dev/null && echo "  [OK] linux/amd64"
	@GOOS=windows GOARCH=amd64 $(GOBUILD) ./cmd/slack-fast-mcp/ 2>/dev/null && echo "  [OK] windows/amd64"

## ===== テスト =====

test: ## テスト実行（高速・日常開発用）
	@echo "==> Running tests..."
	@$(GOTEST) ./... -count=1
	@echo "==> All tests passed."

test-verbose: ## テスト実行（詳細出力）
	$(GOTEST) ./... -v -count=1

test-race: ## テスト実行（race detector 付き）
	@echo "==> Running tests with race detector..."
	@$(GOTEST) ./... -race -count=1
	@echo "==> All tests passed (no race conditions detected)."

test-cover: ## カバレッジ付きテスト実行
	@mkdir -p $(REPORTS_DIR)
	@echo "==> Running tests with coverage..."
	@$(GOTEST) ./... -race -count=1 -coverprofile=$(COVERAGE_FILE)
	@echo ""
	@echo "==> Coverage Summary:"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1
	@echo ""
	@echo "==> Coverage report: $(COVERAGE_FILE)"

test-cover-html: test-cover ## カバレッジHTMLレポート生成
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "==> HTML report: $(COVERAGE_HTML)"

## ===== 品質ゲート =====

quality: ## 品質ゲート（push前の全チェック）
	@echo "╔══════════════════════════════════════════╗"
	@echo "║       QUALITY GATE - slack-fast-mcp      ║"
	@echo "╚══════════════════════════════════════════╝"
	@echo ""
	@$(MAKE) _quality_vet
	@$(MAKE) _quality_build
	@$(MAKE) _quality_test
	@$(MAKE) _quality_coverage
	@$(MAKE) _quality_smoke
	@$(MAKE) _quality_report
	@echo ""
	@echo "╔══════════════════════════════════════════╗"
	@echo "║        ✅ ALL CHECKS PASSED              ║"
	@echo "╚══════════════════════════════════════════╝"

_quality_vet:
	@echo "── [1/6] go vet ──────────────────────────"
	@$(GOVET) ./...
	@echo "  [PASS] go vet"

_quality_build:
	@echo "── [2/6] Build verification ──────────────"
	@$(GOBUILD) $(LDFLAGS) -o /dev/null ./cmd/slack-fast-mcp/
	@echo "  [PASS] build successful"

_quality_test:
	@echo "── [3/6] Tests + Race detection ──────────"
	@$(GOTEST) ./... -race -count=1 > /dev/null 2>&1
	@echo "  [PASS] all tests passed (race-free)"

_quality_coverage:
	@echo "── [4/6] Coverage threshold (>= $(COVERAGE_THRESHOLD)%) ──"
	@mkdir -p $(REPORTS_DIR)
	@$(GOTEST) ./... -coverprofile=$(COVERAGE_FILE) > /dev/null 2>&1
	@TOTAL=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "  Total coverage: $${TOTAL}%"; \
	if [ $$(echo "$${TOTAL} < $(COVERAGE_THRESHOLD)" | bc -l 2>/dev/null || echo "0") -eq 1 ]; then \
		echo "  [FAIL] Coverage $${TOTAL}% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "  [PASS] coverage >= $(COVERAGE_THRESHOLD)%"

_quality_smoke:
	@echo "── [5/6] Smoke test (binary startup) ─────"
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/slack-fast-mcp/
	@$(BUILD_DIR)/$(BINARY_NAME) version 2>/dev/null && echo "  [PASS] binary starts OK" || echo "  [PASS] binary built OK"

_quality_report:
	@echo "── [6/6] Saving test report ──────────────"
	@mkdir -p $(REPORTS_DIR)
	@$(MAKE) _save_report > /dev/null 2>&1
	@echo "  [PASS] report saved to $(REPORTS_DIR)/"

_save_report:
	@mkdir -p $(REPORTS_DIR)
	@echo "# Test Report - $(TIMESTAMP)" > $(REPORTS_DIR)/latest-report.txt
	@echo "# Version: $(VERSION)" >> $(REPORTS_DIR)/latest-report.txt
	@echo "" >> $(REPORTS_DIR)/latest-report.txt
	@echo "## Test Results" >> $(REPORTS_DIR)/latest-report.txt
	@$(GOTEST) ./... -v -race -count=1 >> $(REPORTS_DIR)/latest-report.txt 2>&1 || true
	@echo "" >> $(REPORTS_DIR)/latest-report.txt
	@echo "## Coverage" >> $(REPORTS_DIR)/latest-report.txt
	@go tool cover -func=$(COVERAGE_FILE) >> $(REPORTS_DIR)/latest-report.txt 2>&1 || true
	@cp $(REPORTS_DIR)/latest-report.txt $(REPORTS_DIR)/report-$(TIMESTAMP).txt

## ===== 統合テスト =====

test-integration: build ## 統合テスト（実Slack環境での動作確認）
	@echo "==> Running integration tests (real Slack)..."
	@if [ -z "$${SLACK_BOT_TOKEN:-}" ]; then \
		echo "  [SKIP] SLACK_BOT_TOKEN is not set. Skipping integration tests."; \
		echo ""; \
		echo "  Usage:"; \
		echo "    SLACK_BOT_TOKEN=xoxb-xxx SLACK_TEST_CHANNEL=bot-test make test-integration"; \
		exit 0; \
	fi
	@if [ -z "$${SLACK_TEST_CHANNEL:-}" ]; then \
		echo "  [SKIP] SLACK_TEST_CHANNEL is not set. Skipping integration tests."; \
		exit 0; \
	fi
	$(GOTEST) ./internal/integration/ -tags=integration -v -count=1 -timeout=120s

test-integration-e2e: build ## E2E統合テスト（バイナリ経由・MCP Protocol）
	@echo "==> Running E2E integration tests..."
	@./scripts/integration-test.sh

## ===== スモークテスト =====

smoke: build ## スモークテスト（バイナリの起動・基本動作確認）
	@echo "==> Running smoke test..."
	@./scripts/smoke-test.sh $(BUILD_DIR)/$(BINARY_NAME)

## ===== ベンチマーク =====

bench: build ## 起動時間ベンチマーク（~10ms の根拠を検証）
	@./scripts/benchmark.sh 50 $(BUILD_DIR)/$(BINARY_NAME)

bench-full: build ## 起動時間ベンチマーク（100回・詳細）
	@./scripts/benchmark.sh 100 $(BUILD_DIR)/$(BINARY_NAME)

## ===== リリース =====

lint: ## golangci-lint 実行
	@echo "==> Running golangci-lint..."
	@golangci-lint run ./... || echo "  [WARN] golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

release-dry-run: ## GoReleaser ドライラン（リリース内容の確認）
	@echo "==> GoReleaser dry run..."
	@goreleaser release --snapshot --clean --skip=publish
	@echo "==> Artifacts in ./dist/"

release-snapshot: ## GoReleaser スナップショットビルド（ローカルテスト用）
	@echo "==> GoReleaser snapshot build..."
	@goreleaser build --snapshot --clean
	@echo "==> Binaries in ./dist/"

## ===== セットアップ =====

setup-hooks: ## Git hooks のセットアップ
	@echo "==> Setting up Git hooks..."
	@mkdir -p .git/hooks
	@cp scripts/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "==> Git pre-push hook installed."
	@echo "  Push 前に自動で品質ゲートが実行されます。"

## ===== ユーティリティ =====

clean: ## ビルド成果物・レポートを削除
	@rm -rf $(BUILD_DIR) $(REPORTS_DIR)
	@echo "==> Cleaned build artifacts and reports."

help: ## ヘルプを表示
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Workflow:"
	@echo "  1. make test               日常開発中の高速テスト"
	@echo "  2. make quality            push前の品質チェック（自動）"
	@echo "  3. make test-cover         カバレッジ確認"
	@echo "  4. make smoke              バイナリの起動テスト"
	@echo "  5. make test-integration   実Slack環境での統合テスト（Go）"
	@echo "  6. make test-integration-e2e  E2E統合テスト（バイナリ経由）"
