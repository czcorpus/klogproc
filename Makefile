VERSION=`git describe --tags --always`
BUILD=`date +%FT%T%z`
HASH=`git rev-parse --short HEAD`


LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD} -X main.gitCommit=${HASH}"

all: test build

build:
	go build ${LDFLAGS}

build-race:
	go build -race ${LDFLAGS}

install:
	go install ${LDFLAGS}

clean:
	rm klogproc

test:
	go test ./...

.PHONY: clean install test
