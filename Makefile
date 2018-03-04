.PHONY: install dep container publish httpbin-container httpbin-publish

install:
	go build -o ${GOPATH}/bin/smoke main.go

dep:
	dep ensure

container: dep
	docker build -t bluehoodie/smoke .

publish: container
	docker push bluehoodie/smoke

httpbin-container:
	docker build -t bluehoodie/httpbin -f Dockerfile-httpbin .

httpbin-publish: httpbin-container
	docker push bluehoodie/httpbin