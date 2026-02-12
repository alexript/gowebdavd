BIN_DIR := bin
BIN := $(BIN_DIR)/gowebdavd

.PHONY: all build build-release test cover run clean tidy fmt vet

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/gowebdavd

build-release:
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN) ./cmd/gowebdavd

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	$(BIN) run -dir . -port 8080 -bind 127.0.0.1

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	# Clean Go build/test artifacts (project-local)
	go clean
	# Expire cached test results so next run re-executes tests
	go clean -testcache
	# Remove local build outputs and coverage files
	rm -rf $(BIN_DIR) coverage.out coverage.html coverage
	# Remove any previously built binaries placed at repo root
	rm -f gowebdavd gowebdavd.exe gowebdavd-*
