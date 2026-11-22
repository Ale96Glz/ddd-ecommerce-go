.PHONY: build test fmt run-catalog run-orders run-payments

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -l .

run-catalog:
	go run ./cmd/catalog

run-orders:
	go run ./cmd/orders

run-payments:
	go run ./cmd/payments
