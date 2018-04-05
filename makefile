BINARY = dumper
SOURCE := *.go
BUILD_DIR=$(shell pwd)/bin

all: clean linux

linux: 
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/dumper-linux-amd64 .

clean:
	rm -rf bin/*

container: linux
	docker build -t rogierlommers/dumper .
	docker push rogierlommers/dumper:latest
