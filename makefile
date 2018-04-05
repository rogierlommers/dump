BINARY = dump
SOURCE := *.go
BUILD_DIR=$(shell pwd)/bin

all: clean linux

linux: 
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/dump-linux-amd64 .

clean:
	rm -rf bin/*

container: linux
	docker build -t rogierlommers/dump .
	docker push rogierlommers/dump:latest
