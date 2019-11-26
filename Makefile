export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)

all: build

build:
	@ echo building server
	@ go build -o bin/too_simple_server github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

install:
	@ echo installing server
	@ go install github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

lint:
	@ echo lint install
	@ go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@ echo lint run
	@ golangci-lint run -D unused,deadcode

test:
	@ echo test run
	@ go test github.com/opentelekomcloud-infra/simple-exquisite-webserver/main -v