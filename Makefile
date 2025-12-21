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
test: gen
	@echo "🔍 Запуск тестов..."
	@go test -race ./...
	@echo "📐 Запуск линтера..."
	@golangci-lint run

.PHONY: gen
gen:
	@echo "🧰 Генерация моков..."
	@rm -rf internal/server/service/mocks
	@mkdir -p internal/server/service/mocks
	@rm -rf internal/server/delivery/rest/mocks
	@mkdir -p internal/server/delivery/rest/mocks
	@go generate ./...

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

.PHONY: deps
deps:
	@go get -u ./...
	@go mod tidy
	@go mod verify
	@go build ./...
