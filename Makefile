.DEFAULT_GOAL := server

LDFLAGS = -ldflags "\
	-X main.buildVersion=1.0.0 \
	-X 'main.buildDate=$$(date "+%Y-%m-%d %H:%M:%S")' \
	-X main.buildCommit=$$(git rev-parse --short HEAD)"

.PHONY: server
server: build-server
	@./cmd/server/server $(args)

.PHONY: build-server
build-server:
	@go build ${LDFLAGS} -o ./cmd/server/server ./cmd/server/server.go

.PHONY: agent
agent: build-agent
	@./cmd/agent/agent $(args)

.PHONY: build-agent
build-agent:
	@go build ${LDFLAGS} -o ./cmd/agent/agent ./cmd/agent/agent.go

.PHONY: test
test: gen fmt
	@echo "🔍 Запуск тестов..."
	@go test -race ./...
	@echo "📐 Запуск линтера..."
	@golangci-lint run

.PHONY: gen
gen: gen-reset
	@echo "🧰 Генерация моков..."
	@rm -rf internal/server/service/mocks
	@mkdir -p internal/server/service/mocks
	@rm -rf internal/server/delivery/http/handler/mocks
	@mkdir -p internal/server/delivery/http/handler/mocks
	@go generate ./...

.PHONY: proto-gen
proto-gen:
	@echo "🧰 Генерация protobuf и gRPC кода..."
	@rm -rf internal/proto
	@mkdir -p internal/proto
	@protoc \
    --go_out=. --go_opt=module=github.com/Nekrasov-Sergey/metrics-collector \
    --go-grpc_out=. --go-grpc_opt=module=github.com/Nekrasov-Sergey/metrics-collector \
    api/proto/metrics.proto

.PHONY: gen-reset
gen-reset:
	@go run ./cmd/reset/main.go

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

TARGET_URL := http://localhost:8080/value/
PROFILES_DIR := internal/profiles

BASE_PROFILE := $(PROFILES_DIR)/base.pprof
RESULT_PROFILE := $(PROFILES_DIR)/result.pprof

PPROF_HEAP_URL := http://localhost:8080/debug/pprof/heap

.PHONY: load-test
load-test:
	@echo "Запуск нагрузочного теста (POST /value)"
	hey -n 100000 -c 1000 \
		-m POST \
		-H "Content-Type: application/json" \
		-d '{"id":"PollCount","type":"counter"}' \
		$(TARGET_URL)

.PHONY: pprof-base
pprof-base:
	@mkdir -p $(PROFILES_DIR)
	@echo "Сбор базового heap профиля"
	@go tool pprof -seconds=30 -proto $(PPROF_HEAP_URL) > $(BASE_PROFILE)

.PHONY: pprof-ui-base
pprof-ui-base:
	@echo "pprof ui (base) → http://localhost:6060"
	@go tool pprof -http=:6060 $(BASE_PROFILE)

.PHONY: pprof-result
pprof-result:
	@mkdir -p $(PROFILES_DIR)
	@echo "Сбор итогового heap профиля"
	@go tool pprof -seconds=30 -proto $(PPROF_HEAP_URL) > $(RESULT_PROFILE)

.PHONY: pprof-ui-result
pprof-ui-result:
	@echo "pprof ui (result) → http://localhost:6060"
	@go tool pprof -http=:6060 $(RESULT_PROFILE)

.PHONY: pprof-diff
pprof-diff:
	@echo "pprof diff (base → result)"
	@go tool pprof -top -diff_base=$(BASE_PROFILE) $(RESULT_PROFILE)

.PHONY: pprof-ui-diff
pprof-ui-diff:
	@echo "pprof diff ui → http://localhost:6060"
	@go tool pprof -http=:6060 -diff_base=$(BASE_PROFILE) $(RESULT_PROFILE)

.PHONY: fmt
fmt:
	@echo "Форматирование проекта..."
	@goimports -w .

.PHONY: doc
doc:
	@echo "Документация сервиса http://localhost:6060/pkg/github.com/Nekrasov-Sergey/metrics-collector?m=all"
	@godoc -http=:6060 -play

.PHONY: staticlint
staticlint:
	@go run ./cmd/staticlint $$(go list ./... | \
		grep -v /internal/server/service/mocks | \
		grep -v /internal/server/delivery/rest/mocks)
