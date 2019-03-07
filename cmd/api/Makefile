OUT := api
SRC := $(wildcard *.go)

build: $(SRC)
	GOOS=linux go build -v -o $(OUT)

.PHONY: push
push: build
	docker build -t mxpv/podsync_api .
	docker push mxpv/podsync_api

.PHONY: clean
clean:
	- rm -f $(OUT)
