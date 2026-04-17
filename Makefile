VERSION ?= $(shell git describe --tags --always)
GOBIN ?= $(shell go env GOPATH)/bin

run:
	go run \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		. -v

bin/tpl: build

build:
	go build \
		-ldflags="-X main.BuildVersion=$(VERSION)" \
		-o bin/tpl .

unit-test: bin/tpl
	go test -v ./...

smoke-test: bin/tpl
	export foo="bar"
	export bar="[foo,bar]"
	export foobar="{foo:bar,bar:foo}"
	export foobaz="{foo:[bar,baz]}"
	export baz="1.0-123"
	export number="59614658972"
	export null="null"
	export empty=
	export money="500€"
	export special="?&>=:/"
	export woot="[]"
	export whoa="{}"
	export backslash="\.\/"
	export urls="{google:[https:://google.com,http:://google.de],github:https:://github.com}"
	export json='{"abc":123,"def":["a","b","c"],"ghi":"[{,!?!,}]"}'

	echo
	echo

	go run . -t test/test.tpl | diff -y test/test.txt - && echo Tests succeeded! || echo Tests failed!

	echo
	echo

	go run . -t test/test.tpl -o test/test.out && diff -y test/test.txt test/test.out && echo Tests succeeded! || echo Tests failed!

install: bin/tpl
	@mkdir -p $(GOBIN)
	@cp bin/tpl $(GOBIN)/tpl
	@echo "Installed tpl to $(GOBIN)/tpl"

$(GOBIN)/goimports:
	@go install golang.org/x/tools/cmd/goimports@v0.44.0

$(GOBIN)/gocyclo:
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0

$(GOBIN)/golangci-lint:
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4

$(GOBIN)/gocritic:
	@go install github.com/go-critic/go-critic/cmd/gocritic@v0.14.3

install-linters: $(GOBIN)/goimports $(GOBIN)/gocyclo $(GOBIN)/golangci-lint $(GOBIN)/gocritic
	@echo "Linters installed successfully."

lint: install-linters
	@pre-commit run -a

clean:
	@rm -rfv bin
	@find example -name '*.Dockerfile' -delete
	@find tests -name '*.Dockerfile' -delete
