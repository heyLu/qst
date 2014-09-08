all: build

SHELL = /bin/bash

GOROOT ?= /home/lu/t/go/go
CROSSCOMPILE ?= /home/lu/t/go/golang-crosscompile/crosscompile.bash

PLATFORMS ?= darwin-amd64 freebsd-amd64 linux-amd64

build:
	go build

crossbuild:
	source ${CROSSCOMPILE}; \
	for platform in ${PLATFORMS}; do \
		GOROOT=${GOROOT} go-$$platform build -o qst-$$platform; \
	done

clean:
	rm -f qst
	for platform in ${PLATFORMS}; do \
		rm -f qst-$$platform; \
	done