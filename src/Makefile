PKG := .

.PHONY: all build test run clean deps generate
.PHONY: fmt lint vet codecheck
.PHONY: build-linux

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOBIN_DIR = $(PKG)/bin
GO_BINDINGS_DIR =  $(PKG)/go-bindings

BINARY_FILE = $(GOBIN_DIR)/client

all: build

build: generate
	@echo "+ building source"
	$(GOBUILD) -v -o $(BINARY_FILE)

test: generate
	@echo "+ executing tests"
	$(GOTEST) -v  $(PKG)/...

run: build
	@echo "+ running"
	$(BINARY_FILE)

clean:
	@echo "+ cleaning"
	$(GOCLEAN) -i $(PKG)/...
	rm -rf $(BINARY_FILE) $(GO_BINDINGS_DIR)

deps:
	@echo "+ installing dependencies"
	$(GOGET) github.com/ethereum/go-ethereum
	$(GOGET) github.com/spf13/viper
	$(GOGET) golang.org/x/lint/golint

#TODO: use generate to create the directory structure too
generate:
	@echo "+ go generate"
	mkdir -p $(GO_BINDINGS_DIR)/course
	$(GOCMD) generate $(PKG)/...

codecheck: fmt lint vet

fmt:
	@echo "+ go fmt"
	$(GOCMD) fmt $(PKG)/...

lint:
	@echo "+ go lint"
	golint -min_confidence=0.1 $(PKG)/...

vet:
	@echo "+ go vet"
	$(GOCMD) vet $(PKG)/...

# Cross compilation
# TODO: add others OS (osx, windows)
build-linux: generate
	@echo "+ building $@"
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_FILE)

# Please, uncomment the line below and test in a OSX
# build-osx: generate
# 		CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_FILE) -v

# build-windows: generate
# 		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_FILE) -v