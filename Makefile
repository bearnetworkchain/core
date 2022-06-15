#! /usr/bin/make -f

#項目變量。
PROJECT_NAME = ignite
DATE := $(shell date '+%Y-%m-%dT%H:%M:%S')
FIND_ARGS := -name '*.go' -type f -not -name '*.pb.go'
HEAD = $(shell git rev-parse HEAD)
LD_FLAGS = -X github.com/ignite-hq/cli/ignite/version.Head='$(HEAD)' \
	-X github.com/ignite-hq/cli/ignite/version.Date='$(DATE)'
BUILD_FLAGS = -mod=readonly -ldflags='$(LD_FLAGS)'
BUILD_FOLDER = ./dist

## install:安裝 de 二進製文件。
install:
	@echo 安裝熊網鏈...
	@go install $(BUILD_FLAGS) ./...
	@ignite version

## build:構建二進製文件。
build:
	@echo 建立熊網鏈...
	@-mkdir -p $(BUILD_FOLDER) 2> /dev/null
	@go build $(BUILD_FLAGS) -o $(BUILD_FOLDER) ./...

## clean: 清理構建文件。還在內部運行`go clean`。
clean:
	@echo Cleaning build cache...
	@-rm -rf $(BUILD_FOLDER) 2> /dev/null
	@go clean ./...

## govet: Run go vet.
govet:
	@echo Running go vet...
	@go vet ./...

## format: 運行霍夫曼。
format:
	@echo Formatting...
	@find . $(FIND_ARGS) | xargs gofmt -d -s
	@find . $(FIND_ARGS) | xargs goimports -w -local github.com/ignite-hq/cli

## lint：運行Golang CI Lint。
lint:
	@echo Running gocilint...
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
	@golangci-lint run --out-format=tab --issues-exit-code=0

## test-unit：運行單元測試。
test-unit:
	@echo Running unit tests...
	@go test -race -failfast -v ./ignite/...

## test-integration：運行集成測試。
test-integration: install
	@echo Running integration tests...
	@go test -race -failfast -v -timeout 60m ./integration/...

## 測試：運行單元和集成測試。
test: govet test-unit test-integration

help: Makefile
	@echo
	@echo " 選擇一個命令運行 "$(PROJECT_NAME)", 或者只是運行'make'進行"install"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.DEFAULT_GOAL := install
