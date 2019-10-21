# Makefile for cordctl

# Set bash for fail quickly
SHELL = bash -eu -o pipefail

ifeq ($(GOPATH),)
  $(error Please set your GOPATH)
endif

VERSION     ?= $(shell cat $(GOPATH)/src/github.com/opencord/cordctl/VERSION)
GOVERSION    = $(shell go version 2>&1 | awk '{print $$3;}' )

GITCOMMIT   ?= $(shell git log --pretty=format:"%h" -n 1)
ifeq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
	GITDIRTY  := false
else
	GITDIRTY  := true
endif

# build target creates binaries for host OS and arch
HOST_OS     ?= $(shell uname -s | tr A-Z a-z)

# uname and golang disagree on name of CPU architecture
ifeq ($(shell uname -m),x86_64)
	HOST_ARCH ?= amd64
else
	HOST_ARCH ?= $(shell uname -m)
endif

BUILDTIME    = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS      = -ldflags \
	"-X github.com/opencord/cordctl/cli/version.Version=$(VERSION)  \
	 -X github.com/opencord/cordctl/cli/version.GitCommit=$(GITCOMMIT)  \
	 -X github.com/opencord/cordctl/cli/version.GitDirty=$(GITDIRTY)  \
	 -X github.com/opencord/cordctl/cli/version.GoVersion=$(GOVERSION)  \
	 -X github.com/opencord/cordctl/cli/version.Os=$$GOOS \
	 -X github.com/opencord/cordctl/cli/version.Arch=$$GOARCH \
	 -X github.com/opencord/cordctl/cli/version.BuildTime=$(BUILDTIME)"

# Settings for running with mock server
TEST_PROTOSET = $(shell pwd)/mock/xos-core.protoset
TEST_MOCK_DIR = $(shell pwd)/mock
TEST_SERVER = localhost:50051
TEST_USERNAME = admin@opencord.org
TEST_PASSWORD = letmein

# Default is GO111MODULE=auto, which will refuse to use go mod if running
# go less than 1.13.0 and this repository is checked out in GOPATH. For now,
# force module usage. This affects commands executed from this Makefile, but
# not the environment inside the Docker build (which does not build from
# inside a GOPATH).
export GO111MODULE=on

help:
	@echo "release      - build binaries using cross compliing for the support architectures"
	@echo "build        - build the binary as a local executable"
	@echo "install      - build and install the binary into \$$GOPATH/bin"
	@echo "run          - runs cordctl using the command specified as \$$CMD"
	@echo "lint         - run static code analysis"
	@echo "test         - run unity tests"
	@echo "clean        - remove temporary and generated files"

build:
	export GOOS=$(HOST_OS) ;\
	export GOARCH=$(HOST_ARCH) ;\
	go build $(LDFLAGS) cmd/cordctl/cordctl.go

install:
	export GOOS=$(HOST_OS) ;\
	export GOARCH=$(HOST_ARCH) ;\
	go install -mod=vendor $(LDFLAGS) cmd/cordctl/cordctl.go

run:
	export GOOS=$(HOST_OS) ;\
	export GOARCH=$(HOST_ARCH) ;\
	go run -mod=vendor $(LDFLAGS) cmd/cordctl/cordctl.go $(CMD)

lint:
	find $(GOPATH)/src/github.com/opencord/cordctl -name "*.go" -not -path '$(GOPATH)/src/github.com/opencord/cordctl/vendor/*' | xargs gofmt -l
	go vet -mod=vendor ./...

test:
	@mkdir -p ./tests/results
	@set +e; \
	CORDCTL_PROTOSET=$(TEST_PROTOSET)\
         CORDCTL_SERVER=$(TEST_SERVER) \
         CORDCTL_MOCK_DIR=$(TEST_MOCK_DIR) \
         CORDCTL_USERNAME=$(TEST_USERNAME) \
         CORDCTL_PASSWORD=$(TEST_PASSWORD) \
         go test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	go-junit-report < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	gocover-cobertura < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	cd mock; \
	docker-compose down; \
	exit $$RETURN

# Release related items
# Generates binaries in $RELEASE_DIR with name $RELEASE_NAME-$RELEASE_OS_ARCH
# Inspired by: https://github.com/kubernetes/minikube/releases
RELEASE_DIR     ?= release
RELEASE_NAME    ?= cordctl
RELEASE_OS_ARCH ?= linux-amd64 linux-arm64 windows-amd64 darwin-amd64
RELEASE_BINS    := $(foreach rel,$(RELEASE_OS_ARCH),$(RELEASE_DIR)/$(RELEASE_NAME)-$(rel))

# Functions to extract the OS/ARCH
rel_os    = $(word 2, $(subst -, ,$(notdir $@)))
rel_arch  = $(word 3, $(subst -, ,$(notdir $@)))

$(RELEASE_BINS):
	export GOOS=$(rel_os) ;\
	export GOARCH=$(rel_arch) ;\
	go build -mod=vendor -v $(LDFLAGS) -o "$@" cmd/cordctl/cordctl.go

release: $(RELEASE_BINS)

clean:
	rm -f cordctl $(RELEASE_BINS)
	rm -rf vendor
	rm -f Gopkg.lock

