FROM busybox
COPY smoke /

ENTRYPOINT ["./smoke"]
