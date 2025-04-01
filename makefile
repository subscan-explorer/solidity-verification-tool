.PHONY: check run build

GOCMD=go

export CGO_ENABLED=0

build:
	$(GOCMD) build -tags="netgo" -a -ldflags "-s -w -extldflags '-static'" -o verification .