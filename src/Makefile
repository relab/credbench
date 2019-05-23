BINARY_NAME = client-linux

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Solidity parameters
SOLC_BUILD_DIR = ./ethereum/build
GO_BINDINGS_DIR = ./go-bindings

all: generate build

build: 
		$(GOBUILD) -o $(BINARY_NAME) -v
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -rf $(BINARY_NAME) $(SOLC_BUILD_DIR) $(GO_BINDINGS_DIR)
run:
		$(GOBUILD) -o $(BINARY_NAME) -v ./...
		$(BINARY_NAME)
deps:
		$(GOGET) github.com/spf13/viper
#TODO: use generate to create the directory structure too
generate:
		mkdir -p $(GO_BINDINGS_DIR)/{course,notary,owners}
		$(GOCMD) generate ./...
# Cross compilation
# TODO: add others OS (osx, windows)
build-linux:
		CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME) -v
