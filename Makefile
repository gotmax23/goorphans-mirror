# Link to systen libsqlite3
GO_BUILDTAGS = libsqlite3
GO_FLAGS = -tags '$(GO_BUILDTAGS)' -buildmode pie $(GO_EXTRA_FLAGS)

.PHONY: all allc clean lint goorphans install installd update-deps

all: lint goorphans install
allc: clean lint goorphans installd

clean:
	rm -vf goorphans
lint:
	golangci-lint run --fix
goorphans:
	go build -v $(GO_FLAGS) .
install:
	go install -v $(GO_FLAGS) .
installd:
	install -p goorphans -t $(shell go env GOPATH)/bin/
update-deps:
	go get -u
	go mod tidy
