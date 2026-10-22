FROM alpine:3

WORKDIR /app
RUN apk update && apk add gcompat

# CI builds verification with the Go version pinned in go.mod.
COPY ./verification verification
COPY ./*.json ./

ENTRYPOINT ["/app/verification"]
