# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get -u
GORACE=GORACE="halt_on_error=1"
GOBUILD_RACE = $(GORACE) $(GOBUILD) -race
GOTEST_RACE = $(GORACE) $(GOTEST) -race

SRC_ROOT = .
CLI_DIR = $(SRC_ROOT)/cli
CLIPROTO_DIR = $(CLI_DIR)/proto
PKG_DIR = $(SRC_ROOT)/src
PKGPROTO_DIR = $(PKG_DIR)/schemes
CONTRACTS = $(PKG_DIR)/ethereum
NPM_COMPILE = npm run compile

all: build binary

.PHONY: race
race: cliproto build
	@echo "+ building source using Race Detector"
	$(GOBUILD_RACE) -v -o dist/cli $(CLI_DIR)

.PHONY: racetest
racetest:
	@echo "+ building tests using Race Detector"
	$(GOTEST_RACE) -v $(PKG_DIR)/...

.PHONY: binary
binary: dist cliproto
	@echo "+ building source"
	$(GOBUILD) -v -o dist/cli $(CLI_DIR)

dist:
	mkdir $@

.PHONY: build
build: pkgproto generate
	@echo "+ building source"
	$(GOBUILD) -v $(PKG_DIR)/...

.PHONY: cliproto
cliproto:
	@echo "+ compiling bench proto files"
	@protoc -I=$(CLIPROTO_DIR) --go_out=paths=source_relative:$(CLIPROTO_DIR) $(CLIPROTO_DIR)/*.proto

.PHONY: pkgproto
pkgproto:
	@echo "+ compiling pkg proto files"
	@protoc -I=$(PKGPROTO_DIR) --go_out=paths=source_relative:$(PKGPROTO_DIR) $(PKGPROTO_DIR)/*.proto

generate:
	@echo "+ go generate"
	$(GOCMD) generate $(PKG_DIR)/...

.PHONY: contracts
contracts:
	which npm || ( echo "npm is required to compile the contracts, please install npm: https://github.com/npm/cli" && exit 1)
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
