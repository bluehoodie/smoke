.PHONY: install build container publish

install:
	go build -o ${GOPATH}/bin/smoke main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o smoke main.go

container: build
	docker build -t bluehoodie/smoke .
	rm smoke

publish: container
	docker push bluehoodie/smoke