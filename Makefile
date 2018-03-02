.PHONY: install build container publish httpbin-container httpbin-publish

install:
	go build -o ${GOPATH}/bin/smoke main.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o smoke main.go

container: build
	docker build -t bluehoodie/smoke .
	rm smoke

publish: container
	docker push bluehoodie/smoke

httpbin-container:
	docker build -t bluehoodie/httpbin -f Dockerfile-httpbin .

httpbin-publish: test-container
	docker push bluehoodie/httpbin