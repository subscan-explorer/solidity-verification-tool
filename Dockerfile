FROM alpine:3

WORKDIR /app
RUN apk update && apk add gcompat

COPY ./verification verification
COPY ./*.json ./

ENTRYPOINT ["/app/verification"]
