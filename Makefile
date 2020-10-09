.PHONY: build clean deploy
GIT_COMMIT := $(shell git rev-list -1 HEAD)

build:
	go build -ldflags="-s -w -X main.gitCommit=$(GIT_COMMIT)" 

clean:
	rm -rf ./bin

test:
	go run handlers/web/main.go

run:
	go build
	./energy-sdk

invoke: clean
	env GOOS=linux go build -ldflags="-s -w -X main.gitCommit=$(GIT_COMMIT)" -o bin/energy-sdk  main.go
	sls invoke --verbose local -f update

deploy: clean
	env GOOS=linux go build -ldflags="-s -w -X main.gitCommit=$(GIT_COMMIT)" -o bin/energy-sdk  main.go
	sls deploy

# Testing CI/CD workflows