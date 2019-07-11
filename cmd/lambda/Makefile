OUT := lambda
SRC := $(wildcard *.go)

build: $(SRC)
	GOOS=linux go build -v -o $(OUT)

.PHONY: push
push: build
	docker build -t mxpv/podsync_lambda .
	docker push mxpv/podsync_lambda

.PHONY: clean
clean:
	- rm -f $(OUT)
