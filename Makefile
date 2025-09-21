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

.PHONY: metricstest1
metricstest1: build-server
	metricstest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

.PHONY: metricstest2
metricstest2: build-agent
	 metricstest -test.v -test.run=^TestIteration2[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent

.PHONY: metricstest3
metricstest3: build-server build-agent
	metricstest -test.v -test.run=^TestIteration3[AB]*$$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server

.PHONY: metricstest4
metricstest4: build-server build-agent
	SERVER_PORT=$$(random unused-port) ADDRESS="localhost:$${SERVER_PORT}" TEMP_FILE=$$(random tempfile) metricstest -test.v -test.run=^TestIteration4$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$SERVER_PORT \
            -source-path=.

.PHONY: test
test:
	@go test -v -cover ./...

.PHONY: cover
cover:
	@go test -v -coverprofile=cover.out ./...
	@go tool cover -html=cover.out -o cover.html
