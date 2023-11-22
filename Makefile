VERSION=`git describe --tags --always`
BUILD=`date +%FT%T%z`
HASH=`git rev-parse --short HEAD`


LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.buildDate=${BUILD} -X main.gitCommit=${HASH}"

all: test build

build:
	go build -o conomi ${LDFLAGS}

install:
	@if [ -z "$$CONOMI_INSTALL_PATH" ]; then \
		echo "CONOMI_INSTALL_PATH is not defined"; \
    else \
		mkdir -p ${CONOMI_INSTALL_PATH}; \
		cp conomi ${CONOMI_INSTALL_PATH}/conomi; \
		cp -R dist ${CONOMI_INSTALL_PATH}/dist; \
		cp -R assets ${CONOMI_INSTALL_PATH}/assets; \
		mkdir -p ${CONOMI_INSTALL_PATH}/templates; \
		cp templates/*.gtpl ${CONOMI_INSTALL_PATH}/templates; \
    fi

clean:
	rm conomi

test:
	go test ./...

.PHONY: clean install test