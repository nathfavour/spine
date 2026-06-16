.PHONY: build clean

build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/spine ./cmd/spine

clean:
	rm -rf bin
