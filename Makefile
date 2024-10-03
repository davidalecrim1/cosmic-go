lint:
	golangci-lint run

integration-test:
	go test -tags integration ./...

run:
	go run ./cmd/api/server.go