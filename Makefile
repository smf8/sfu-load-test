GO_LDFLAGS = -ldflags "-s -w"
GO_VERSION = 1.14
GO_TESTPKGS:=$(shell go list ./... | grep -v cmd | grep -v conf | grep -v node)
GO_COVERPKGS:=$(shell echo $(GO_TESTPKGS) | paste -s -d ',')
TEST_UID:=$(shell id -u)
TEST_GID:=$(shell id -g)

all: nodes

go_deps:
	go mod download

clean:
	rm -rf bin

build: go_deps
	go build -o bin/sfu-load-test $(GO_LDFLAGS) main.go
