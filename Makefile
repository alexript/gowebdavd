BIN_DIR := bin
BIN := $(BIN_DIR)/gowebdavd
BIN_WIN := $(BIN_DIR)/gowebdavd.exe

# Detect OS
ifeq ($(OS),Windows_NT)
    SHELL := cmd
    MKDIR = if not exist $(subst /,\\,$(1)) mkdir $(subst /,\\,$(1))
    RM = if exist $(subst /,\\,$(1)) rmdir /s /q $(subst /,\\,$(1))
    RM_F = if exist $(subst /,\\,$(1)) del /f $(subst /,\\,$(1))
    BIN_TARGET := $(BIN_WIN)
else
    MKDIR = mkdir -p $(1)
    RM = rm -rf $(1)
    RM_F = rm -f $(1)
    BIN_TARGET := $(BIN)
endif

.PHONY: all build build-release test cover run clean tidy fmt vet

all: build

build:
	@$(call MKDIR,$(BIN_DIR))
	go build -o $(BIN_TARGET) ./cmd/gowebdavd

build-release:
	@$(call MKDIR,$(BIN_DIR))
	go build -ldflags="-s -w" -o $(BIN_TARGET) ./cmd/gowebdavd

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	$(BIN_TARGET) run -dir . -port 8080 -bind 127.0.0.1

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	go clean
	go clean -testcache
	@$(call RM,$(BIN_DIR))
	@$(call RM_F,coverage.out)
	@$(call RM_F,coverage.html)
	@$(call RM,coverage)
	@$(call RM_F,gowebdavd)
	@$(call RM_F,gowebdavd.exe)
