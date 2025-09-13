.DEFAULT_GOAL := server

.PHONY: server
server: build-server
	@./cmd/server/server

.PHONY: build-server
build-server:
	@go build -o ./cmd/server/server ./cmd/server/server.go

.PHONY: agent
agent: build-agent
	@./cmd/agent/agent

.PHONY: build-agent
build-agent:
	@go build -o ./cmd/agent/agent ./cmd/agent/agent.go

.PHONY: metricstest1
metricstest1: build-server
	metricstest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

.PHONY: test
test:
	@go test -v -cover

.PHONY: cover
cover:
	@go test -v -coverprofile=cover.out
	@go tool cover -html=cover.out -o cover.html
