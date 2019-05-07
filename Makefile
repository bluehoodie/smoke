.PHONY: install dep container publish binary httpbin-container httpbin-publish test

install:
	go build -o ${GOPATH}/bin/smoke main.go

dep:
	dep ensure

container: dep
	docker build -t bluehoodie/smoke .

publish: container
	docker push bluehoodie/smoke

binary: dep
	goreleaser --rm-dist

httpbin-container:
	docker build -t bluehoodie/httpbin -f Dockerfile-httpbin .

httpbin-publish: httpbin-container
	docker push bluehoodie/httpbin

test: install
	ENVTOKEN=token235 smoke -f ./smoke_test.json -u https://httpbin.org -v
	ENVTOKEN=token235 smoke -f ./smoke_test.yaml -u https://httpbin.org -v