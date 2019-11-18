.PHONY: lint test test-cov check-tidy

lint:
	go mod download
# need to mount /go/pkg to provide dependencies into container
# otherwise go will fail to download private repos from github
	docker run --rm \
		--name gpm_lint \
		-v `go env GOPATH`/pkg:/go/pkg \
		-v `go env GOCACHE`:/root/.cache/go-build \
		-v `pwd`:/app \
		-w /app \
		golangci/golangci-lint:v1.21.0 \
		golangci-lint run

test:
	go test -race -p=8 -parallel=8 ./...

test-cov:
	go test -race -p=8 -parallel=8 -coverpkg ./... -coverprofile=coverage.out ./...

check-tidy:
	go mod tidy
	if [[ `git status --porcelain go.mod` ]]; then git diff -- go.mod ; echo "go.mod is outdated, please run go mod tidy" ; exit 1; fi
	if [[ `git status --porcelain go.sum` ]]; then git diff -- go.sum ; echo "go.sum is outdated, please run go mod tidy" ; exit 1; fi
