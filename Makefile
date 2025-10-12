.DEFAULT_GOAL := server

.PHONY: server
server: build-server
	@./cmd/server/server $(args)

.PHONY: build-server
build-server:
	@go build -o ./cmd/server/server ./cmd/server/server.go

.PHONY: agent
agent: build-agent
	@./cmd/agent/agent $(args)

.PHONY: build-agent
build-agent:
	@go build -o ./cmd/agent/agent ./cmd/agent/agent.go

.PHONY: test
test:
	@echo "🔍 Запуск тестов..."
	@go test -v ./...

.PHONY: cover
cover:
	@echo "📊 Генерация отчёта покрытия..."
	@go test -coverprofile=cover.out ./...
	@echo ""
	@echo "🧮 Общий процент покрытия:"
	@go tool cover -func=cover.out | grep total | awk '{print $$3}'
	@echo ""
	@echo "🌐 HTML-отчёт сохранён в: cover.html"
	@go tool cover -html=cover.out -o cover.html
