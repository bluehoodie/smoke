FROM golang:latest as builder
WORKDIR /go/src/github.com/bluehoodie/smoke
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o smoke .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/bluehoodie/smoke/smoke .
ENTRYPOINT ["./smoke"]
