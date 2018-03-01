.PHONY: install build

install:
	go build -o ${GOPATH}/bin/smoke main.go

build:
	go build -o smoke main.go
