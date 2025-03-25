.DEFAULT_GOAL := test

PREFIX := $(HOME)/local/gitmirror

fmt:
	golangci-lint run --disable-all --no-config -Egofmt --fix
	golangci-lint run --disable-all --no-config -Egofumpt --fix

lint: fmt
	staticcheck ./...
	revive -config revive.toml -exclude internal/flag ./...
	golangci-lint run

golangci: fmt
	golangci-lint run

staticcheck: fmt
	staticcheck ./...

revive: fmt
	revive -config revive.toml -exclude internal/flag ./...

test:
	go test -shuffle on github.com/telemachus/pluggo/internal/cli
	go test -shuffle on github.com/telemachus/pluggo/internal/git
	go test -shuffle on github.com/telemachus/pluggo/internal/opts

testv:
	go test -shuffle on -v github.com/telemachus/pluggo/internal/cli
	go test -shuffle on -v github.com/telemachus/pluggo/internal/git
	go test -shuffle on -v github.com/telemachus/pluggo/internal/opts

testr:
	go test -race -shuffle on github.com/telemachus/pluggo/internal/cli
	go test -race -shuffle on github.com/telemachus/pluggo/internal/git
	go test -race -shuffle on github.com/telemachus/pluggo/internal/opts

build: lint testr
	go build ./cmd/pluggo

install: build
	go install ./cmd/pluggo

clean:
	rm -f pluggo
	go clean -i -r -cache

.PHONY: fmt lint build install test testv testr clean
