BINPATH := $(abspath ./bin)
GOLANGCI := $(BINPATH)/golangci-lint

$(GOLANGCI):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BINPATH) v1.17.1
	$(GOLANGCI) --version

.PHONY: lint
lint: $(GOLANGCI)
	$(GOLANGCI) run

.PHONY: test
test:
	go test ./...
