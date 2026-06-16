.PHONY: build clean

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/spine ./cmd/spine

clean:
	rm -rf bin
