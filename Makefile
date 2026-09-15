SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

BINARY      := orders
BIN_DIR     := bin
PKG         := ./...
COVER_FILE  := coverage.out

GO             ?= go
GOLANGCI_LINT  ?= golangci-lint
GOLANGCI_VER   := v2.12.2

.PHONY: help
help: ## Показать все команды
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: check
check: tidy-check fmt-check lint test ## Главная команда: ровно то же, что гоняет CI

.PHONY: test
test: ## Тесты под -race с покрытием
	$(GO) test -race -covermode=atomic -coverprofile=$(COVER_FILE) $(PKG)

.PHONY: race
race: ## Пять прогонов тестов под -race: гонки любят прятаться
	$(GO) test -race -count=5 $(PKG)

.PHONY: run
run: ## Запустить сервис локально (подхватит .env, если он есть)
	@set -a; [ -f .env ] && . ./.env || true; set +a; $(GO) run ./cmd/$(BINARY)

.PHONY: build
build: ## Собрать бинарник в bin/
	$(GO) build -o $(BIN_DIR)/$(BINARY) ./cmd/$(BINARY)

.PHONY: fmt
fmt: ## Отформатировать код
	$(GO) fmt $(PKG)

.PHONY: fmt-check
fmt-check: ## Проверить форматирование, ничего не меняя
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "Не отформатировано (запустите make fmt):"; echo "$$files"; exit 1; \
	fi

.PHONY: lint
lint: ## Линтер (go vet входит в него, отдельно запускать не нужно)
	@command -v $(GOLANGCI_LINT) >/dev/null || { \
		echo "golangci-lint не найден. Установите: make tools"; exit 1; }
	@have=$$($(GOLANGCI_LINT) version --short 2>/dev/null); want=$$(echo $(GOLANGCI_VER) | tr -d 'v'); \
	if [ "$$have" != "$$want" ]; then \
		echo "Внимание: локально golangci-lint $$have, в CI $$want — результаты могут разойтись."; \
		echo "Поставить версию курса: make tools"; \
	fi
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: ## Линтер с автоисправлением
	$(GOLANGCI_LINT) run --fix

.PHONY: tidy
tidy: ## Привести go.mod/go.sum в порядок
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Проверить, что go.mod/go.sum актуальны
	@$(GO) mod tidy -diff >/dev/null || { \
		echo "go.mod/go.sum устарели — запустите make tidy"; exit 1; }

.PHONY: cover
cover: test ## Открыть отчёт о покрытии в браузере
	$(GO) tool cover -html=$(COVER_FILE)

.PHONY: tools
tools: ## Поставить golangci-lint версии курса
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VER)

.PHONY: clean
clean: ## Удалить сборку и отчёты
	rm -rf $(BIN_DIR) $(COVER_FILE) coverage.html
