.PHONY: run build install clean

run:
	go run ./cmd/bashhub

build:
	go build -o bashhub ./cmd/bashhub

install: build
	cp bashhub $(GOPATH)/bin/bashhub

clean:
	rm -f bashhub

