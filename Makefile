.PHONY: all build clean

BINARY="octopus"

CEPH_VERSION="nautilus"

all: gotool build

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags=nautilus -o ${BINARY}

gotool:
	go fmt ./
	go vet ./

clean:
	@if [ -f ${BINARY} ];then rm ${BINARY};fi
