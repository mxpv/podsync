SUBDIRS := cmd/api cmd/nginx
BINPATH := $(abspath ./bin)
GOLANGCI := $(BINPATH)/golangci-lint

.PHONY: push
push:
	for d in $(SUBDIRS); do $(MAKE) -C $$d push; done

$(GOLANGCI):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BINPATH) v1.16.0
	$(GOLANGCI) --version

.PHONY: lint
lint: $(GOLANGCI)
	$(GOLANGCI) run

.PHONY: test
test:
	go test -short ./...

.PHONY: up
up:
	docker-compose pull
	docker-compose up -d

.PHONY: static
static:
	- rm -rf dist/
	npm install
	npm run build
