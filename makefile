.PHONY: run all linux

run:
	@go run *.go

all:
	@mkdir -p bin/
	@bash --norc -i ./scripts/build.sh

linux:
	@mkdir -p bin/
	@export GOOS=linux && export GOARCH=amd64 && bash --norc -i ./scripts/build.sh
