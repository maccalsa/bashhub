.PHONY: run build

run:
	go run ./cmd/bashhub

build:
	go build -o bashhub ./cmd/bashhub
