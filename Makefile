.PHONY: install dep container publish binary httpbin-container httpbin-publish smoke-test

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

smoke-test: install
	-docker stop httpbin
	docker run -d -p 8000:8000 --rm --name httpbin bluehoodie/httpbin
	sleep 5
	ENVTOKEN=token235 smoke -f ./smoke_test.json -u http://localhost -p 8000 -v
	ENVTOKEN=token235 smoke -f ./smoke_test.yaml -u http://localhost -p 8000 -v
	docker stop httpbin