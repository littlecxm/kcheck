.PHONY: all build windows linux m1 clean fmt help
BUILD_DATE := $(shell date +'%Y-%m-%d %H:%M:%S')
SHA_SHORT := $(shell git rev-parse --short HEAD)
VERSION := $(shell cat version)
GOOS ?= windows
GOARCH ?= amd64

all: fmt build
build:
	@go mod download
	@echo build kcheck...
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o build/ -ldflags "-s -X \"main.version=${VERSION}\" -X \"main.buildDate=${BUILD_DATE}\" -X \"main.commitID=${SHA_SHORT}\"" -v ./cmd/kcheck
	@echo build makecheck...
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o build/ -ldflags "-s -X \"main.version=${VERSION}\" -X \"main.buildDate=${BUILD_DATE}\" -X \"main.commitID=${SHA_SHORT}\"" -v ./cmd/makecheck
windows: GOOS=windows
windows: GOARCH=amd64
windows: build
linux: GOOS=linux
linux: GOARCH=amd64
linux: build
m1: GOOS=darwin 
m1: GOARCH=arm64
m1: build
clean:
	@go clean
	@find ./build -type f -exec rm -r {} +
fmt:
	@echo Formatting...
	@go fmt ./cmd/*
	@go fmt ./pkg/*
	@go fmt ./tests/*
	@go vet ./cmd/*
	@go vet ./pkg/*
	@go vet ./tests/*
help:
	@echo "make: make"
	@echo "make <windows|linux|m1>: build for specific target"
.EXPORT_ALL_VARIABLES:
GO111MODULE = on
