VERSION=`git describe --tags --always`
BUILD=`date +%FT%T%z`
HASH=`git rev-parse --short HEAD`


LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD} -X main.gitCommit=${HASH}"

all: test build

build:
	go build -o conomi ${LDFLAGS}

install:
	go install -o conomi ${LDFLAGS}

clean:
	rm conomi

test:
	go test ./...

.PHONY: clean install test