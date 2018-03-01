.PHONY: install build

install:
	go build -o ${GOPATH}/bin/smoke main.go

build:
	CGO_ENABLED=0 go build -o ./smoke -a -ldflags '-s' -installsuffix cgo main.go

docker-build: build
	docker build -t smoke .
	rm smoke
