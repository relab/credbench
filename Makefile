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
BENCH_DIR = $(SRC_ROOT)/bench
BENCHPROTO_DIR = $(BENCH_DIR)/proto
PKG_DIR = $(SRC_ROOT)/pkg
PKGPROTO_DIR = $(PKG_DIR)/schemes

all: build bench

bench: dist benchproto
	@echo "+ building source"
	$(GOBUILD) -v -o dist/ctbench $(BENCH_DIR)

dist:
	mkdir $@

build: pkgproto
	@echo "+ building source"
	$(GOBUILD) -v $(PKG_DIR)/...

benchproto:
	@echo "+ compiling bench proto files"
	@protoc -I=$(BENCHPROTO_DIR) --go_out=paths=source_relative:$(BENCHPROTO_DIR) $(BENCHPROTO_DIR)/*.proto

pkgproto:
	@echo "+ compiling pkg proto files"
	@protoc -I=$(PKGPROTO_DIR) --go_out=paths=source_relative:$(PKGPROTO_DIR) $(PKGPROTO_DIR)/*.proto

.PHONY: tools
tools:
	@echo "+ installing tools"
	@go mod download
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -I % go install %

.PHONY: test
test:
	@echo "+ executing tests"
	$(GOTEST) -v -count=1 $(PKG_DIR)/...

.PHONY: racetest
racetest:
	@echo "+ building tests using Race Detector"
	$(GOTEST_RACE) -v $(PKG_DIR)/...

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
