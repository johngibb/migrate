all: install test-all

install:
	go install ./...

test:
	go test ./...

test-all:
	docker-compose rm -f postgres # discard state from prior runs
	docker-compose up --build --abort-on-container-exit --force-recreate

.PHONY: all test test-all
