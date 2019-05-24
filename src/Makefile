.PHONY: all build test run clean deps generate
.PHONY: build-linux

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOBIN_DIR = ./bin
GO_BINDINGS_DIR = ./go-bindings

BINARY_FILE = $(GOBIN_DIR)/client

all: 	build

build:	generate
		$(GOBUILD) -o $(BINARY_FILE) -v

test:	generate
		$(GOTEST) -v ./...

run: 	build
		$(BINARY_FILE)

clean:
		$(GOCLEAN)
		rm -rf $(BINARY_FILE) $(GO_BINDINGS_DIR)

deps:
		$(GOGET) github.com/ethereum/go-ethereum
		$(GOGET) github.com/spf13/viper

#TODO: use generate to create the directory structure too
generate:
		mkdir -p $(GO_BINDINGS_DIR)/course
		$(GOCMD) generate ./...

# Cross compilation
# TODO: add others OS (osx, windows)
build-linux: generate
		CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_FILE) -v

# Please, uncomment the line below and test in a OSX
# build-osx: generate
# 		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_FILE) -v

# build-windows: generate
# 		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_FILE) -v