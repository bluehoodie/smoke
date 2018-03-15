FROM golang:1.10.0-alpine3.7 as builder
WORKDIR /go/src/github.com/bluehoodie/smoke
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o smoke .

FROM scratch
COPY --from=builder /go/src/github.com/bluehoodie/smoke/smoke .
ENTRYPOINT ["./smoke"]
