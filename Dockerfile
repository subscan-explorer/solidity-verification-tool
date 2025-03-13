FROM alpine:3

WORKDIR /app

COPY ./verification verification

ENTRYPOINT ["/app/verification"]
