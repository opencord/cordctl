ifeq ($(GOPATH),)
GOPATH=$(shell pwd)
endif

VERSION=$(shell cat $(GOPATH)/src/github.com/opencord/cordctl/VERSION)
GITCOMMIT=$(shell git log --pretty=format:"%h" -n 1)
GOVERSION=$(shell go version 2>&1 | sed -E  's/.*(go[0-9]+\.[0-9]+\.[0-9]+).*/\1/g')
OSTYPE=$(shell uname -s | tr A-Z a-z)
OSARCH=$(shell uname -p | tr A-Z a-z)
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags \
	'-X "github.com/opencord/cordctl/cli/version.Version=$(VERSION)"  \
	 -X "github.com/opencord/cordctl/cli/version.GitCommit=$(GITCOMMIT)"  \
	 -X "github.com/opencord/cordctl/cli/version.GoVersion=$(GOVERSION)"  \
	 -X "github.com/opencord/cordctl/cli/version.Os=$(OSTYPE)" \
	 -X "github.com/opencord/cordctl/cli/version.Arch=$(OSARCH)" \
	 -X "github.com/opencord/cordctl/cli/version.BuildTime=$(BUILDTIME)"' \

help:

build: dependencies
	GOPATH=$(GOPATH) go build $(LDFLAGS) cmd/cordctl.go

dependencies:
	dep ensure

lint:
	GOPATH=$(GOPATH) find $(GOPATH)/src/github.com/opencord/cordctl -name "*.go" -not -path '$(GOPATH)/src/github.com/opencord/cordctl/vendor/*' | xargs gofmt -l
	GOPATH=$(GOPATH) go vet ./...    
#	GOPATH=$(GOPATH) go vet github.com/opencord/cordctl/...
	dep check

test:
	GOPATH=$(GOPATH) go test $(TEST_ARGS) -cover -coverprofile=cordctl.cp ./...
#	GOPATH=$(GOPATH) go test $(TEST_ARGS) -cover -coverprofile=cordctl.cp github.com/opencord/cordctl/...

view-coverage:
	GOPATH=$(GOPATH) go tool cover -html cordctl.cp
