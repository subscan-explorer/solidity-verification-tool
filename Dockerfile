FROM golang:1.22-alpine

ENV CONF_DIR=/configs

WORKDIR /app

COPY . .

RUN go build -o verification

ENTRYPOINT ["./verification"]
