# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get -u

SRC_ROOT = .
CLI_DIR = $(SRC_ROOT)/cli
BENCHPROTO_DIR = $(CLI_DIR)/proto
PKG_DIR = $(SRC_ROOT)/src
PKGPROTO_DIR = $(PKG_DIR)/schemes
CONTRACTS = $(PKG_DIR)/ethereum
NPM_COMPILE = npm run compile

all: binary

.PHONY: binary
binary: dist benchproto
	@echo "+ building source"
	$(GOBUILD) -v -o dist/cli $(CLI_DIR)

dist:
	mkdir $@

.PHONY: build
build: pkgproto npm generate
	@echo "+ building source"
	$(GOBUILD) -v $(PKG_DIR)/...

.PHONY: benchproto
benchproto:
	@echo "+ compiling bench proto files"
	@protoc -I=$(BENCHPROTO_DIR) --go_out=paths=source_relative:$(BENCHPROTO_DIR) $(BENCHPROTO_DIR)/*.proto

.PHONY: pkgproto
pkgproto:
	@echo "+ compiling pkg proto files"
	@protoc -I=$(PKGPROTO_DIR) --go_out=paths=source_relative:$(PKGPROTO_DIR) $(PKGPROTO_DIR)/*.proto

generate:
	@echo "+ go generate"
	$(GOCMD) generate $(PKG_DIR)/...

.PHONY: npm
npm:
	which npm || ( echo "install npm for your system from https://github.com/npm/cli" && exit 1)
	cd $(CONTRACTS) && $(NPM_COMPILE)

.PHONY: test
test:
	@echo "+ executing tests"
	$(GOTEST) -v -count=1 $(PKG_DIR)/...

.PHONY: clean
clean:
	@echo "+ cleaning"
	$(GOCLEAN) -i ./...
	rm -rf dist $(CONTRACTS)/build

.PHONY: codecheck
codecheck: fmt lint vet

.PHONY: fmt
fmt:
	@echo "+ go fmt"
	$(GOCMD) fmt $(SRC_ROOT)/...

.PHONY: lint
lint:
	@echo "+ go lint"
	golint -min_confidence=0.1 $(SRC_ROOT)/...

.PHONY: vet
vet:
	@echo "+ go vet"
	$(GOCMD) vet $(SRC_ROOT)/...
