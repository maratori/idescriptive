.DEFAULT_GOAL := test

LINTER_VERSION = v1.27.0
lint:
	go mod download
	$(RUN_LINTER_IN_DOCKER)
.PHONY: lint

test:
	go test -race -p=8 -parallel=8 ./...
.PHONY: test

test-cover:
	go test -race -p=8 -parallel=8 -coverpkg ./... -coverprofile=coverage.out ./...
.PHONY: test-cover

tidy:
	go mod tidy
.PHONY: tidy

update-deps:
	go get -u -t ./...
	go mod tidy
.PHONY: update-deps

check-tidy:
	cp go.mod go.check.mod
	cp go.sum go.check.sum
	go mod tidy -modfile=go.check.mod
	diff -u go.mod go.check.mod
	diff -u go.sum go.check.sum
	rm go.check.mod go.check.sum
.PHONY: check-tidy

# need to mount /go/pkg to provide dependencies into container
# otherwise go will fail to download private repos from github
RUN_LINTER_IN_DOCKER = docker run --rm                                                                                                \
                                  --name idescriptive_lint                                                                            \
                                  --mount type=bind,consistency=delegated,source="`go env GOPATH | cut -d : -f 1`/pkg",target=/go/pkg \
                                  --mount type=bind,consistency=delegated,source="`go env GOCACHE`",target=/root/.cache/go-build      \
                                  --mount type=bind,consistency=delegated,source="`pwd`",target=/app                                  \
                                  -w /app                                                                                             \
                                  golangci/golangci-lint:$(LINTER_VERSION)                                                            \
                                  golangci-lint run
