SUBDIRS := cmd/app cmd/nginx cmd/ytdl
GOLANGCI := ./bin/golangci-lint

.PHONY: push
push:
	for d in $(SUBDIRS); do $(MAKE) -C $$d push; done

$(GOLANGCI):
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ./bin v1.15.0
	./bin/golangci-lint --version

.PHONY: lint
lint: $(GOLANGCI)
	./bin/golangci-lint run

.PHONY: test
test:
	go test -short ./...

.PHONY: up
up:
	docker-compose pull
	docker-compose up -d
