PROJECT=$(shell pwd);


export GO111MODULE=on
export GOPATH=$(PROJECT)

build:
	go build -o bin/too_simple_server github.com/opentelekomcloud-infra/simple-exquisite-webserver/main

install:


