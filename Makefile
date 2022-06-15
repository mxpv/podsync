BINPATH := $(abspath ./bin)

.PHONY: all
all: build test

#
# Build Podsync CLI binary
#
.PHONY: build
build:
	go build -o bin/podsync ./cmd/podsync

#
# Build Docker image
#
TAG ?= localhost/podsync
.PHONY: docker
docker:
	docker build -t $(TAG) .
	docker push $(TAG)

#
# Run unit tests
#
.PHONY: test
test:
	go test -v ./...

#
# Clean
#
.PHONY: clean
clean:
	- rm -rf $(BINPATH)
