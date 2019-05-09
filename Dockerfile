FROM golang:1.12.5-alpine3.9 as builder
WORKDIR /go/src/github.com/bluehoodie/smoke
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o smoke .

FROM alpine:latest
COPY --from=builder /go/src/github.com/bluehoodie/smoke/smoke .
ENTRYPOINT ["./smoke"]
