.PHONY: run test test-race lint build docker-build docker-up

run:
	go run ./cmd/api

test:
	go test ./... -v

test-race:
	go test ./... -race

lint:
	golangci-lint run ./...

build:
	go build -o bin/api ./cmd/api

docker-build:
	docker build -t capim-clinics-api .

docker-up:
	docker compose up --build
