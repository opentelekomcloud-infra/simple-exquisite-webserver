PROJECT=$(shell pwd);


export GO111MODULE=on
export GOPATH=$(PROJECT)

all: build

build:
	go build -o bin/too_simple_server github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

install:
	go install -i github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

test:
	go test github.com/opentelekomcloud-infra/simple-exquisite-webserver/main -v

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint