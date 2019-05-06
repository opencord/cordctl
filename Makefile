ifeq ($(GOPATH),)
GOPATH=$(shell pwd)
endif

help:

build:
	GOPATH=$(GOPATH) go install github.com/opencord/cordctl

lint:
	GOPATH=$(GOPATH) find $(GOPATH)/src/github.com/opencord/cordctl -name "*.go" -not -path '$(GOPATH)/src/github.com/opencord/cordctl/vendor/*' | xargs gofmt -l
	GOPATH=$(GOPATH) go vet github.com/opencord/cordctl/...
	dep check

test:
	GOPATH=$(GOPATH) go test $(TEST_ARGS) -cover -coverprofile=cordctl.cp github.com/opencord/cordctl/...

view-coverage:
	GOPATH=$(GOPATH) go tool cover -html cordctl.cp
