#@IgnoreInspection BashAddShebang

GO_LDFLAGS = -ldflags "-s -w"
GO_VERSION = 1.14
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v conf | grep -v node)
GO_COVERPKGS:=$(shell echo $(GO_TESTPKGS) | paste -s -d ',')
TEST_UID:=$(shell id -u)
TEST_GID:=$(shell id -g)

all: format lint build

check-formatter:
	which goimports || GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports

format: check-formatter
	find $(ROOT) -type f -name "*.go" -not -path "$(ROOT)/vendor/*" | xargs -n 1 -I R goimports -w R
	find $(ROOT) -type f -name "*.go" -not -path "$(ROOT)/vendor/*" | xargs -n 1 -I R gofmt -s -w R

check-linter:
	which golangci-lint || (GO111MODULE=off go get -u -v github.com/golangci/golangci-lint/cmd/golangci-lint)

lint: check-linter
	golangci-lint run --deadline 10m ./...

go_deps:
	go mod download

clean:
	rm -rf bin

build: go_deps
	go build -o bin/sfu-load-test $(GO_LDFLAGS) main.go