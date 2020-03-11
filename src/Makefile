PKG := .

.PHONY: all build test run clean deps generate solc proto
.PHONY: fmt lint vet codecheck
.PHONY: build-linux build-osx build-windows

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get -u
GOBIN_DIR = $(PKG)/bin
SCHEMES_DIR = $(PKG)/schemes
GO_BINDINGS_DIR = {$(PKG)/core/course/contract,$(PKG)/core/notary/contracts}
CONTRACTS_SRC = ./ethereum/contracts
SOLC_BUILD_DIR = ./ethereum/build
EXTERNAL_CONTRACTS_SRC = ./ethereum/node_modules/bbchain-contracts/contracts,
#../../node_modules/@openzeppelin/contracts/,
CONTRACTS_DEPS = bbchain-contracts/contracts=./ethereum/node_modules/bbchain-contracts/contracts
# @openzeppelin/contracts=../../node_modules/@openzeppelin/contracts
BBCHAIN_LIBS = ./ethereum/node_modules/bbchain-contracts/contracts/CertificateSum.sol,
SOLC_ARGS = --optimize --gas --optimize-runs 200 --overwrite --combined-json abi,asm,ast,bin,bin-runtime,compact-format,devdoc,hashes,interface,metadata,opcodes,srcmap,srcmap-runtime,userdoc $(CONTRACTS_DEPS)
BINARY_FILE = $(GOBIN_DIR)/bbchain

# TODO: install contract dependencies: bbchain-contracts + openzeppelin contracts
# TODO: use generate to create the directory structure too
# TODO: generate abi/bin files in contracts directories (maybe calling the npm)

all: proto generate build

build:
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
	$(GOGET) github.com/spf13/cobra/cobra
	$(GOGET) golang.org/x/lint/golint
	$(GOGET) github.com/golang/protobuf/protoc-gen-go

proto:
	@echo Compiling proto files
	@protoc -I=$(SCHEMES_DIR) --go_out=$(SCHEMES_DIR) $(SCHEMES_DIR)/*.proto

generate:
	@echo "+ go generate"
	mkdir -p $(GO_BINDINGS_DIR)
	$(GOCMD) generate $(PKG)/...

solc:
	solc $(SOLC_ARGS) --allow-paths $(BBCHAIN_LIBS) $(CONTRACTS_SRC)/Course.sol -o $(SOLC_BUILD_DIR)

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

build-linux:
	@echo "+ building $@"
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_FILE)

build-osx:
	@echo "+ building $@"
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_FILE)

build-windows:
	@echo "+ building $@"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_FILE)