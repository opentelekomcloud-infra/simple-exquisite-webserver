PROJECT=$(shell pwd);


export GO111MODULE=on
export GOPATH=$(PROJECT)

all: go-get go-install go-build

go-build:
	go build -o bin/too_simple_server github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

go-install:
	go install -i github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

go-get:
	go get -d -u github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

test: lint go-test 

go-test:
	go test github.com/opentelekomcloud-infra/simple-exquisite-webserver/main -v

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint