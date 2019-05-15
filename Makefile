ifeq ($(GOPATH),)
$(error Please set your GOPATH)
endif

VERSION=$(shell cat $(GOPATH)/src/github.com/opencord/cordctl/VERSION)
GITCOMMIT=$(shell git log --pretty=format:"%h" -n 1)
ifeq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
GITDIRTY=false
else
GITDIRTY=true
endif
GOVERSION=$(shell go version 2>&1 | sed -E  's/.*(go[0-9]+\.[0-9]+\.[0-9]+).*/\1/g')
OSTYPE=$(shell uname -s | tr A-Z a-z)
OSARCH=$(shell uname -p | tr A-Z a-z)
BUILDTIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags \
	'-X "github.com/opencord/cordctl/cli/version.Version=$(VERSION)"  \
	 -X "github.com/opencord/cordctl/cli/version.GitCommit=$(GITCOMMIT)"  \
	 -X "github.com/opencord/cordctl/cli/version.GitDirty=$(GITDIRTY)"  \
	 -X "github.com/opencord/cordctl/cli/version.GoVersion=$(GOVERSION)"  \
	 -X "github.com/opencord/cordctl/cli/version.Os=$(OSTYPE)" \
	 -X "github.com/opencord/cordctl/cli/version.Arch=$(OSARCH)" \
	 -X "github.com/opencord/cordctl/cli/version.BuildTime=$(BUILDTIME)"'

help:

build: dependencies
	go build $(LDFLAGS) cmd/cordctl.go

dependencies:
	[ -d "vendor" ] || dep ensure

lint: dependencies
	find $(GOPATH)/src/github.com/opencord/cordctl -name "*.go" -not -path '$(GOPATH)/src/github.com/opencord/cordctl/vendor/*' | xargs gofmt -l
	go vet ./...
	dep check

test: dependencies
	@mkdir -p ./tests/results
	@go test -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	go-junit-report < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	gocover-cobertura < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

clean:
	rm -f cordctl
	rm -rf vendor
	rm -f Gopkg.lock
