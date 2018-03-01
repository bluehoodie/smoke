.PHONY: install build build-docker

install:
	go build -o ${GOPATH}/bin/smoke main.go

build:
	CGO_ENABLED=0 go build -o ./smoke -a -ldflags '-s' -installsuffix cgo main.go

build-docker: build
	docker build -t smoke .
	
