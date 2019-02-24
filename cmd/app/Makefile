OUT := server
SRC := $(wildcard *.go)

build: $(SRC)
	GOOS=linux go build -v -o $(OUT)

.PHONY: push
push: build
	docker build -t mxpv/podsync .
	docker push mxpv/podsync

.PHONY: clean
clean:
	- rm -f $(OUT)
