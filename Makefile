.DEFAULT_GOAL := test

fmt:
	golangci-lint run --disable-all --no-config -Egofmt --fix
	golangci-lint run --disable-all --no-config -Egofumpt --fix

staticcheck: fmt
	staticcheck ./...

revive: fmt
	revive -config revive.toml ./...

golangci: fmt
	golangci-lint run

lint: fmt staticcheck revive golangci

test:
	go test -shuffle on github.com/telemachus/pluggo/internal/cli
	go test -shuffle on github.com/telemachus/pluggo/internal/git

testv:
	go test -shuffle on -v github.com/telemachus/pluggo/internal/cli
	go test -shuffle on -v github.com/telemachus/pluggo/internal/git

testr:
	go test -race -shuffle on github.com/telemachus/pluggo/internal/cli
	go test -race -shuffle on github.com/telemachus/pluggo/internal/git

build: lint testr
	go build .

install: build
	go install .

clean:
	rm -f pluggo
	go clean -i -r -cache

.PHONY: fmt staticcheck revive golangci lint build install test testv testr clean
