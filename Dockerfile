FROM alpine:3.10

COPY ks /

ENTRYPOINT ["/ks"]
