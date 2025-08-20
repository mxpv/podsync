BINPATH := $(abspath ./bin)

.PHONY: all
all: build test

#
# Build Podsync CLI binary
# Example:
# 	$ GOOS=amd64 make build
#

GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)

TAG ?= $(shell git tag --points-at HEAD)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE := $(shell date)

#
# Go optimizations
# -ldflags -s Remove symbol table
# -ldflags -w Remove debug information
# -trimpath Remove all file system paths from the compiled binary
# -tags netgo Use the netgo network stack (Go DNS resolver)
#
LDFLAGS := "-s -w -X 'main.version=${TAG}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}' -X 'main.arch=${GOARCH}'"

.PHONY: build
build:
	go build -trimpath -tags netgo -ldflags ${LDFLAGS} -o bin/podsync ./cmd/podsync

#
# Build a local Docker image
# Example:
# 	$ make docker
#	$ docker run -it --rm localhost/podsync:latest
#
IMAGE_TAG ?= localhost/podsync
.PHONY: docker
docker:
	docker buildx build -t $(IMAGE_TAG) .

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
