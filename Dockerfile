FROM alpine:3

WORKDIR /app

COPY ./verification verification
COPY ./*.json ./

ENTRYPOINT ["/app/verification"]
